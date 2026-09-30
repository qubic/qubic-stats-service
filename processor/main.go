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
	"github.com/qubic/qubic-stats-processor/epochstats"
	"github.com/qubic/qubic-stats-processor/service"
	"github.com/qubic/qubic-stats-processor/spectrum"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const prefix = "QUBIC_STATS_PROCESSOR"

type Configuration struct {
	App struct {
		Mode string `conf:"default:service"`
	}
	SpectrumParser struct {
		SpectrumFile string `conf:"default:./latest.118"`
		OutputMode   string `conf:"default:db"`
		OutputFile   string `conf:"default:spectrumData.json"`
	}
	Service struct {
		QueryServiceGrpcAddress string `conf:"default:localhost:8001"`
		LiveServiceGrpcAddress  string `conf:"default:localhost:8002"`

		CoinGeckoToken     string        `cong:"default:XXXXXXXXXXXXXXXXXXXXX"`
		DataScrapeInterval time.Duration `conf:"default:1m"`
		DataScrapeTimeout  time.Duration `conf:"default:15s"`
	}
	Mongo struct {
		Username string `conf:"default:user"`
		Password string `conf:"default:pass"`
		Hostname string `conf:"default:localhost"`
		Port     string `conf:"default:27017"`
		Options  string

		Database             string `conf:"default:qubic_frontend"`
		DataCollection       string `conf:"default:general_data"`
		RichListCollection   string `conf:"default:rich_list"`
		EpochStatsCollection string `conf:"default:epoch_stats"`
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

			MongoClient:               client,
			MongoDatabase:             config.Mongo.Database,
			MongoQubicDataCollection:  config.Mongo.DataCollection,
			MongoEpochStatsCollection: config.Mongo.EpochStatsCollection,

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

		//start := time.Now()

		s, err := spectrum.ReadSpectrumFromFile(config.SpectrumParser.SpectrumFile)
		if err != nil {
			return errors.Wrap(err, "loading spectrum from file")
		}

		//elapsed := time.Since(start)
		//fmt.Printf("Spectrum file read took: %s\n", elapsed.String())
		//start = time.Now()

		results, err := spectrum.CalculateSpectrumData(s)
		if err != nil {
			return errors.Wrap(err, "calculating spectrum data")
		}

		//elapsed = time.Since(start)
		//fmt.Printf("Spectrum file read took: %s\n", elapsed.String())

		if config.SpectrumParser.OutputMode == "file" {
			err = results.Data.SaveSpectrumDataToFile(config.SpectrumParser.OutputFile)
			if err != nil {
				return errors.Wrap(err, "saving spectrum data")
			}
			break
		}

		spectrumEpoch, err := parseEpochFromFileName(config.SpectrumParser.SpectrumFile)
		if err != nil {
			return errors.Wrap(err, "reading epoch from spectrum file name")
		}

		record, err := epochstats.NewRecord(spectrumEpoch, results.Data.CirculatingSupply, results.Data.ActiveAddresses)
		if err != nil {
			return errors.Wrap(err, "building epoch stats record")
		}

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

		err = saveSpectrumResults(context.Background(), client, &config, record, results.List)
		if err != nil {
			return err
		}

		break
	}

	return nil
}

// parseEpochFromFileName reads the epoch from the extension of a spectrum file, as in
// "spectrum.119".
func parseEpochFromFileName(fileName string) (uint32, error) {

	index := strings.LastIndex(fileName, ".")
	if index < 0 || index == len(fileName)-1 {
		return 0, errors.Errorf("no epoch extension in file name [%s]", fileName)
	}

	epoch, err := strconv.ParseUint(fileName[index+1:], 10, 32)
	if err != nil {
		return 0, errors.Wrapf(err, "parsing epoch from file name [%s]", fileName)
	}

	return uint32(epoch), nil
}

// saveSpectrumResults stores what a spectrum file measured. The epoch record is always written, so
// that older spectrum files can be parsed to fill in the supply history. The rich list is only
// replaced when the file is at least as recent as every other one parsed so far.
func saveSpectrumResults(ctx context.Context, client *mongo.Client, config *Configuration, record epochstats.Record, richList spectrum.RichList) error {

	latest, found, err := epochstats.LoadLatest(ctx, client, config.Mongo.Database, config.Mongo.EpochStatsCollection)
	if err != nil {
		return errors.Wrap(err, "loading latest epoch stats record")
	}

	// The rich list is replaced before the record is written, so that the rich list is already in
	// place once the record makes the api report the new epoch.
	if !found || record.Epoch >= latest.Epoch {
		err = spectrum.ReplaceRichList(ctx, client, config.Mongo.Database, config.Mongo.RichListCollection, richList)
		if err != nil {
			return errors.Wrap(err, "replacing rich list")
		}
	} else {
		fmt.Printf("Epoch %d is older than the latest stored epoch %d. Keeping the current rich list.\n", record.Epoch, latest.Epoch)
	}

	err = epochstats.Save(ctx, client, config.Mongo.Database, config.Mongo.EpochStatsCollection, record)
	if err != nil {
		return errors.Wrap(err, "saving epoch stats record")
	}

	fmt.Printf("Saved epoch %d: Circ supply: %d, Total emitted: %d, Active addr: %d, Epoch end: %d\n",
		record.Epoch, record.CirculatingSupply, record.TotalEmitted, record.ActiveAddresses, record.EpochEndTimestamp)

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
	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, errors.Wrap(err, "creating database client")
	}

	return client, nil

}
