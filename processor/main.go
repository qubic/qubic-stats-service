package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ardanlabs/conf"
	"github.com/pkg/errors"
	"github.com/qubic/qubic-stats-processor/epochstats"
	"github.com/qubic/qubic-stats-processor/service"
	"github.com/qubic/qubic-stats-processor/spectrum"
	"go.mongodb.org/mongo-driver/v2/bson"
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
		SpectrumCollection   string `conf:"default:spectrum_data"`
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
	case "backfill_epoch_stats":
		break
	default:
		return errors.New("Bad app mode. Accepted values: 'service', 'spectrum_parser', 'backfill_epoch_stats'")
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
			MongoSpectrumCollection:   config.Mongo.SpectrumCollection,
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

		epochNumber, err := parseEpochFromFileName(config.SpectrumParser.SpectrumFile)
		if err != nil {
			return errors.Wrap(err, "reading epoch from spectrum file name")
		}
		epoch := strconv.FormatUint(uint64(epochNumber), 10)

		// The epoch the measurement belongs to is what lets the service tell a fresh supply from one
		// that is carried over from an earlier epoch.
		results.Data.Epoch = epochNumber

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

		richListCollection := config.Mongo.RichListCollection + "_" + epoch

		err = spectrum.SaveRichListToDatabase(context.Background(), client, config.Mongo.Database, richListCollection, results.List)
		if err != nil {
			return errors.Wrap(err, "saving rich list")
		}

		// The rich lists of earlier epochs are only dropped once the new one is in place, so that the
		// api never queries a collection that has just been removed.
		err = purgeOldEpochCollections(client, config.Mongo.Database, config.Mongo.RichListCollection, richListCollection)
		if err != nil {
			return fmt.Errorf("purging old epoch rich lists: %w", err)
		}

		latestData, err := spectrum.LoadSpectrumDataFromDatabase(context.Background(), client, config.Mongo.Database, config.Mongo.SpectrumCollection)
		if err != nil {
			return errors.Wrap(err, "reading back spectrum data")
		}

		fmt.Printf("Latest data read from db: Epoch: %d, Circ supply: %d, Active addr: %d Update timestamp: %d\n", latestData.Epoch, latestData.CirculatingSupply, latestData.ActiveAddresses, latestData.Timestamp)
		break

	case "backfill_epoch_stats":

		println("Epoch stats backfill")

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

		err = runEpochStatsBackfill(context.Background(), client, &config)
		if err != nil {
			return errors.Wrap(err, "backfilling epoch stats")
		}

		break
	}

	return nil
}

// purgeOldEpochCollections drops the per epoch rich list collections, except the one that was just
// written. Only collections named like "<base>_<epoch>" are considered, so that anything else that
// happens to share the prefix is left alone.
func purgeOldEpochCollections(mongoClient *mongo.Client, database string, richListCollectionBase string, keep string) error {

	db := mongoClient.Database(database)
	collections, err := db.ListCollectionNames(context.Background(), bson.D{})
	if err != nil {
		return fmt.Errorf("getting list of mongo collection names: %w", err)
	}

	pattern, err := regexp.Compile("^" + regexp.QuoteMeta(richListCollectionBase) + `_\d+$`)
	if err != nil {
		return fmt.Errorf("compiling rich list collection pattern: %w", err)
	}

	for _, c := range collections {
		if c == keep || !pattern.MatchString(c) {
			continue
		}
		err := db.Collection(c).Drop(context.Background())
		if err != nil {
			return fmt.Errorf("dropping collection %s: %w", c, err)
		}
		fmt.Printf("Dropped epoch rich list collection: %s\n", c)
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

// runEpochStatsBackfill reconstructs the per epoch supply history from the historical general data
// and writes the records that do not exist yet. Existing records are never modified.
func runEpochStatsBackfill(ctx context.Context, client *mongo.Client, config *Configuration) error {

	println("Reconstructing epoch stats from the general data...")

	report, err := epochstats.BuildBackfill(ctx, client, config.Mongo.Database, config.Mongo.DataCollection, config.Mongo.SpectrumCollection)
	if err != nil {
		return fmt.Errorf("building backfill: %w", err)
	}

	if len(report.Records) == 0 {
		println("No general data to reconstruct from. Nothing written.")
		return nil
	}

	first := report.Records[0]
	last := report.Records[len(report.Records)-1]

	fmt.Printf("Reconstructed %d epochs, %d to %d.\n", len(report.Records), first.Epoch, last.Epoch)
	fmt.Printf("  Epoch %d: supply %d, emitted %d, ended %d\n", first.Epoch, first.CirculatingSupply, first.TotalEmitted, first.EpochEndTimestamp)
	fmt.Printf("  Epoch %d: supply %d, emitted %d, ended %d\n", last.Epoch, last.CirculatingSupply, last.TotalEmitted, last.EpochEndTimestamp)

	if len(report.MissingEpochs) > 0 {
		fmt.Printf("WARNING: no general data for %d epoch(s) in the range: %v\n", len(report.MissingEpochs), report.MissingEpochs)
	}
	if len(report.CarriedOverEpochs) > 0 {
		fmt.Printf("WARNING: %d epoch(s) have no spectrum measurement of their own and repeat an earlier supply: %v\n", len(report.CarriedOverEpochs), report.CarriedOverEpochs)
	}
	if len(report.SupplyMismatches) > 0 {
		fmt.Printf("WARNING: %d epoch(s) where the spectrum measurement disagrees with the derived supply: %v\n", len(report.SupplyMismatches), report.SupplyMismatches)
	}

	written, err := epochstats.InsertMissing(ctx, client, config.Mongo.Database, config.Mongo.EpochStatsCollection, report.Records)
	if err != nil {
		return err
	}

	fmt.Printf("Wrote %d new epoch stats record(s), left %d existing one(s) untouched.\n", written, len(report.Records)-written)

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
