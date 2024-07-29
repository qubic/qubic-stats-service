package main

import (
	"context"
	"fmt"
	"github.com/ardanlabs/conf"
	"github.com/pkg/errors"
	"github.com/qubic/qubic-stats-api/cache"
	"github.com/qubic/qubic-stats-api/rpc"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"time"
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
			HttpAddress string `conf:"default:0.0.0.0:8080"`
			GrpcAddress string `conf:"default:0.0.0.0:8081"`

			/* Possible options:
			recurrent: update information recurrently based on an interval
			on_request: update information during a request, if the cache validity expired
			none: update information on every request
			*/
			// To be implemented in the future
			//CachingStrategy            string        `conf:"default:recurrent"`
			CacheValidityDuration      time.Duration `conf:"default:10s"`
			SpectrumDataUpdateInterval time.Duration `conf:"default:24h"`
		}
		Mongo struct {
			Username string `conf:"default:user"`
			Password string `conf:"default:pass"`
			Hostname string `conf:"default:localhost"`
			Port     string `conf:"default:27017"`
			Options  string

			Database           string `conf:"default:qubic_frontend"`
			SpectrumCollection string `conf:"default:spectrum_data"`
			DataCollection     string `conf:"default:general_data"`
			RichListCollection string `conf:"default:rich_list"`
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

		CacheValidityDuration:    config.Service.CacheValidityDuration,
		SpectrumValidityDuration: config.Service.SpectrumDataUpdateInterval,
	}

	cacheService := cache.NewCacheService(&serviceConfiguration, dbClient)

	exit := cacheService.Start()

	server := rpc.NewServer(
		config.Service.HttpAddress,
		config.Service.GrpcAddress,
		cacheService.Cache,
		dbClient,
		config.Mongo.Database,
		config.Mongo.RichListCollection,
	)
	err = server.Start()
	if err != nil {
		return errors.Wrap(err, "starting the web server")
	}

	<-exit

	return nil
}

type MongoConfiguration struct {
	Username          string
	Password          string
	Hostname          string
	Port              string
	ConnectionOptions string
}

func (c *MongoConfiguration) AssembleConnectionURI() string {

	return fmt.Sprintf("mongodb://%s:%s@%s:%s/%s", c.Username, c.Password, c.Hostname, c.Port, c.ConnectionOptions)
}

func createMongoClient(configuration *MongoConfiguration) (*mongo.Client, error) {

	serverApi := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(configuration.AssembleConnectionURI()).SetServerAPIOptions(serverApi)
	client, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		return nil, errors.Wrap(err, "creating database client")
	}

	return client, nil

}
