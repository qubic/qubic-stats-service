package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ardanlabs/conf"
	"github.com/jellydator/ttlcache/v3"
	"github.com/pkg/errors"
	qubic "github.com/qubic/go-node-connector"
	"github.com/qubic/go-node-connector/types"
	"github.com/qubic/qubic-stats-api/cache"
	"github.com/qubic/qubic-stats-api/rpc"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const prefix = "QUBIC_STATS_API"

func main() {
	if err := run(); err != nil {
		log.Fatalf("main: exited with error: %s", err.Error())
	}
}

func run() error {
	var config struct {
		Service struct {
			HttpAddress                string        `conf:"default:0.0.0.0:8080"`
			GrpcAddress                string        `conf:"default:0.0.0.0:8081"`
			CacheValidityDuration      time.Duration `conf:"default:10s"`
			SpectrumDataUpdateInterval time.Duration `conf:"default:24h"`
			RichListPageSize           int32         `conf:"default:100"`
			CacheUpdateTimeout         time.Duration `conf:"default:30s"`
			RichListLimit              int           `conf:"default:10000"`
		}
		Mongo struct {
			Username string `conf:"default:user"`
			Password string `conf:"default:pass"`
			Hostname string `conf:"default:localhost"`
			Port     string `conf:"default:27017"`
			Options  string

			Database           string        `conf:"default:qubic_frontend"`
			SpectrumCollection string        `conf:"default:spectrum_data"`
			DataCollection     string        `conf:"default:general_data"`
			RichListCollection string        `conf:"default:rich_list"`
			Timeout            time.Duration `conf:"default:15s"`
		}
		Pool struct {
			NodeFetcherUrl     string        `conf:"default:http://127.0.0.1:8070/status"`
			NodeFetcherTimeout time.Duration `conf:"default:2s"`
			NodePort           string        `conf:"default:21841"`
			InitialCap         int           `conf:"default:5"`
			MaxIdle            int           `conf:"default:20"`
			MaxCap             int           `conf:"default:30"`
			IdleTimeout        time.Duration `conf:"default:15s"`
		}
		AssetService struct {
			Ttl time.Duration `conf:"default:10m"`
		}
	}

	if err := conf.Parse(os.Args[1:], prefix, &config); err != nil {
		switch {
		case errors.Is(err, conf.ErrHelpWanted):
			usage, err := conf.Usage(prefix, &config)
			if err != nil {
				return errors.Wrap(err, "generating config usage")
			}
			fmt.Println(usage)
			return nil
		case errors.Is(err, conf.ErrVersionWanted):
			version, err := conf.VersionString(prefix, &config)
			if err != nil {
				return errors.Wrap(err, "generating config version")
			}
			fmt.Println(version)
			return nil
		}
		return errors.Wrap(err, "parsing config")
	}

	out, err := conf.String(&config)
	if err != nil {
		return errors.Wrap(err, "generating config for output")
	}
	log.Printf("main: Config :\n%v\n", out)

	mongoConnection := MongoConfiguration{
		Username:          config.Mongo.Username,
		Password:          config.Mongo.Password,
		Hostname:          config.Mongo.Hostname,
		Port:              config.Mongo.Port,
		ConnectionOptions: config.Mongo.Options,
		Timeout:           config.Mongo.Timeout,
	}

	println("Connecting to database...")

	dbClient, err := createMongoClient(&mongoConnection)
	if err != nil {
		return errors.Wrap(err, "creating database client")
	}

	defer func() {
		if err = dbClient.Disconnect(context.Background()); err != nil {
			log.Fatalf("database: exited with error: %s\n", err.Error())
		}
	}()

	serviceConfiguration := cache.ServiceConfiguration{
		MongoDatabase:            config.Mongo.Database,
		MongoSpectrumCollection:  config.Mongo.SpectrumCollection,
		MongoQubicDataCollection: config.Mongo.DataCollection,
		MongoRichListCollection:  config.Mongo.RichListCollection,

		CacheValidityDuration:    config.Service.CacheValidityDuration,
		SpectrumValidityDuration: config.Service.SpectrumDataUpdateInterval,

		RichListPageSize: config.Service.RichListPageSize,

		CacheUpdateTimeout: config.Service.CacheUpdateTimeout,
	}

	cacheService := cache.NewCacheService(&serviceConfiguration, dbClient)
	exit := cacheService.Start()

	pool, err := qubic.NewPoolConnection(qubic.PoolConfig{
		InitialCap:         config.Pool.InitialCap,
		MaxCap:             config.Pool.MaxCap,
		MaxIdle:            config.Pool.MaxIdle,
		IdleTimeout:        config.Pool.IdleTimeout,
		NodeFetcherUrl:     config.Pool.NodeFetcherUrl,
		NodeFetcherTimeout: config.Pool.NodeFetcherTimeout,
		NodePort:           config.Pool.NodePort,
	})
	if err != nil {
		return errors.Wrap(err, "creating qubic pool")
	}
	log.Print("Created qubic node pool.")
	assetOwnersCache := ttlcache.New[string, *types.AssetOwnerships](
		ttlcache.WithTTL[string, *types.AssetOwnerships](config.AssetService.Ttl),
		ttlcache.WithDisableTouchOnHit[string, *types.AssetOwnerships](),
		ttlcache.WithCapacity[string, *types.AssetOwnerships](uint64(1024*1024*100)), // 100 MB
	)
	go assetOwnersCache.Start()
	log.Printf("Created cache for asset owners with ttl [%s]", config.AssetService.Ttl)
	assetsService := rpc.NewAssetService(pool, assetOwnersCache)
	log.Print("Created assets service.")

	server := rpc.NewServer(
		config.Service.HttpAddress,
		config.Service.GrpcAddress,
		cacheService.Cache,
		dbClient,
		config.Mongo.Database,
		assetsService,
		config.Mongo.RichListCollection,
		config.Service.RichListPageSize,
		config.Service.RichListLimit,
	)
	err = server.Start()
	if err != nil {
		return errors.Wrap(err, "starting the web server")
	}
	log.Print("Started web server.")

	<-exit

	return nil
}

type MongoConfiguration struct {
	Username          string
	Password          string
	Hostname          string
	Port              string
	ConnectionOptions string
	Timeout           time.Duration
}

func (c *MongoConfiguration) AssembleConnectionURI() string {
	return fmt.Sprintf("mongodb://%s:%s@%s:%s/%s", c.Username, c.Password, c.Hostname, c.Port, c.ConnectionOptions)
}

func createMongoClient(configuration *MongoConfiguration) (*mongo.Client, error) {

	serverApi := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(configuration.AssembleConnectionURI()).SetServerAPIOptions(serverApi).SetTimeout(configuration.Timeout)
	client, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		return nil, errors.Wrap(err, "creating database client")
	}

	return client, nil
}
