package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ardanlabs/conf"
	"github.com/pkg/errors"
	"github.com/qubic/qubic-stats-processor/service"
	"github.com/qubic/qubic-stats-processor/spectrum"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const prefix = "QUBIC_STATS_PROCESSOR"

type Configuration struct {
	App struct {
		Mode string `conf:"default:service"`
	}
	SpectrumParser struct {
		SpectrumSize string `conf:"default:16777216"`
		SpectrumFile string `conf:"default:./latest.118"`
		OutputMode   string `conf:"default:db"`
		OutputFile   string `conf:"default:spectrumData.json"`
	}
	Service struct {
		QueryServiceGrpcAddress string `conf:"default:localhost:8001"`
		LiveServiceGrpcAddress  string `conf:"default:localhost:8002"`

		CoinGeckoToken     string        `cong:"default:XXXXXXXXXXXXXXXXXXXXX"`
		DataScrapeInterval time.Duration `conf:"default:1m"`
		DataScrapeTimeout  time.Duration `conf:"default:5s"`
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

func main() {

	if err := run(); err != nil {
		log.Fatalf("main: exited with error: %s\n", err.Error())
	}
}

func validateConfig(config *Configuration) error {

	//TODO: improve validation

	switch config.App.Mode {
	case "service":
		break
	case "spectrum_parser":
		break
	default:
		return errors.New("Bad app mode. Accepted values: 'service', 'spectrum_parser'")
	}

	switch config.SpectrumParser.OutputMode {
	case "db":
		break
	case "file":
		break
	default:
		return errors.New("Bad parser output mode. Accepted values: 'db', 'file'")
	}

	return nil
}

func run() error {
	var config Configuration

	if err := conf.Parse(os.Args[1:], prefix, &config); err != nil {
		switch err {
		case conf.ErrHelpWanted:
			usage, err := conf.Usage(prefix, &config)
			if err != nil {
				return errors.Wrap(err, "generating config usage")
			}
			fmt.Println(usage)
			return nil
		case conf.ErrVersionWanted:
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

	err = validateConfig(&config)
	if err != nil {
		return errors.Wrap(err, "failed to validate configuration")
	}

	switch config.App.Mode {
	case "service":
		println("Processor")

		mongoConnection := MongoConfiguration{
			Username:          config.Mongo.Username,
			Password:          config.Mongo.Password,
			Hostname:          config.Mongo.Hostname,
			Port:              config.Mongo.Port,
			ConnectionOptions: config.Mongo.Options,
		}

		println("Connecting to database...")
		client, err := createMongoClient(&mongoConnection)
		if err != nil {
			return errors.Wrap(err, "connecting to database")
		}

		defer func() {
			if err = client.Disconnect(context.Background()); err != nil {
				log.Fatalf("main: exited with error: %s\n", err.Error())
			}
		}()

		s := service.Service{
			CoinGeckoToken:          config.Service.CoinGeckoToken,
			QueryServiceGrpcAddress: config.Service.QueryServiceGrpcAddress,
			LiveServiceGrpcAddress:  config.Service.LiveServiceGrpcAddress,

			MongoClient:              client,
			MongoDatabase:            config.Mongo.Database,
			MongoSpectrumCollection:  config.Mongo.SpectrumCollection,
			MongoQubicDataCollection: config.Mongo.DataCollection,

			ScrapeInterval: config.Service.DataScrapeInterval,
			ScrapeTimeout:  config.Service.DataScrapeTimeout,
		}

		err = s.RunService()
		if err != nil {
			return errors.Wrap(err, "running the web service")
		}

		break
	case "spectrum_parser":

		println("Spectrum parser")

		spectrumSize, err := strconv.ParseInt(config.SpectrumParser.SpectrumSize, 10, 64)
		if err != nil {
			return errors.Wrap(err, "parsing spectrum size")
		}

		s, err := spectrum.ReadSpectrumFromFile(config.SpectrumParser.SpectrumFile, spectrumSize)
		if err != nil {
			return errors.Wrap(err, "loading spectrum from file")
		}

		results, err := spectrum.CalculateSpectrumData(s)
		if err != nil {
			return errors.Wrap(err, "calculating spectrum data")
		}

		if config.SpectrumParser.OutputMode == "file" {
			err = results.Data.SaveSpectrumDataToFile(config.SpectrumParser.OutputFile)
			if err != nil {
				return errors.Wrap(err, "saving spectrum data")
			}
			break
		}

		epoch := ""

		split := strings.Split(config.SpectrumParser.SpectrumFile, ".")
		if len(split) < 2 {
			return errors.New("cannot parse file extension into epoch")
		}
		epoch = split[len(split)-1]

		mongoConnection := MongoConfiguration{
			Username:          config.Mongo.Username,
			Password:          config.Mongo.Password,
			Hostname:          config.Mongo.Hostname,
			Port:              config.Mongo.Port,
			ConnectionOptions: config.Mongo.Options,
		}

		println("Connecting to database...")
		client, err := createMongoClient(&mongoConnection)
		if err != nil {
			return errors.Wrap(err, "connecting to database")
		}

		defer func() {
			if err = client.Disconnect(context.Background()); err != nil {
				log.Fatalf("main: exited with error: %s\n", err.Error())
			}
		}()

		err = results.Data.SaveSpectrumDataToDatabase(context.Background(), client, config.Mongo.Database, config.Mongo.SpectrumCollection)
		if err != nil {
			return errors.Wrap(err, "saving spectrum data")
		}

		err = spectrum.SaveRichListToDatabase(context.Background(), client, config.Mongo.Database, config.Mongo.RichListCollection+"_"+epoch, results.List)

		latestData, err := spectrum.LoadSpectrumDataFromDatabase(context.Background(), client, config.Mongo.Database, config.Mongo.SpectrumCollection)

		fmt.Printf("Latest data: %d, %d %d", latestData.CirculatingSupply, latestData.ActiveAddresses, latestData.Timestamp)

		break
	}

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
