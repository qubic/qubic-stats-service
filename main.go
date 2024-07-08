package main

import (
	"context"
	"fmt"
	"github.com/ardanlabs/conf"
	"github.com/pkg/errors"
	"github.com/qubic/frontend-service-processor/src/db"
	"github.com/qubic/frontend-service-processor/src/spectrum"
	"log"
	"os"
	"strconv"
)

const prefix = "QUBIC_FRONTEND_PROCESSOR"

type Configuration struct {
	App struct {
		Mode string `conf:"default:service"`
	}
	SpectrumParser struct {
		SpectrumSize string `conf:"default:16777216"`
		SpectrumFile string `conf:"default:./latest.qs"`
		OutputMode   string `conf:"default:db"`
		OutputFile   string `conf:"default:spectrumData.json"`
	}
	Service struct {
		SpectrumData   string `conf:"default:spectrumData.json"`
		CoinGeckoToken string `cong:"default:XXXXXXXXXXXXXXXXXXXXX"`
		ServerPort     string `conf:"default:80"`
	}
	Mongo struct {
		Username string `conf:"default:user"`
		Password string `conf:"default:pass"`
		Hostname string `conf:"default:localhost"`
		Port     string `conf:"default:27017"`
		Options  string

		Database           string `conf:"default:qubic_frontend"`
		SpectrumCollection string `conf:"default:spectrum_data"`
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
		println("Web service")

		//var result bson.M

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

		spectrumData, err := s.CalculateSpectrumData()
		if err != nil {
			return errors.Wrap(err, "calculating spectrum data")
		}

		if config.SpectrumParser.OutputMode == "file" {
			err = spectrumData.SaveSpectrumDataToFile(config.SpectrumParser.OutputFile)
			if err != nil {
				return errors.Wrap(err, "saving spectrum data")
			}
			break
		}

		mongoConnection := db.Connection{
			Username:          config.Mongo.Username,
			Password:          config.Mongo.Password,
			Hostname:          config.Mongo.Hostname,
			Port:              config.Mongo.Port,
			ConnectionOptions: config.Mongo.Options,
		}

		println("Connecting to database...")
		client, err := db.CreateClient(&mongoConnection)
		if err != nil {
			return errors.Wrap(err, "connecting to database")
		}

		defer func() {
			if err = client.Disconnect(context.Background()); err != nil {
				log.Fatalf("main: exited with error: %s\n", err.Error())
			}
		}()

		err = spectrumData.SaveSpectrumDataToDatabase(client, config.Mongo.Database, config.Mongo.SpectrumCollection)
		if err != nil {
			return errors.Wrap(err, "saving spectrum data")
		}
		break

		break
	}

	return nil
}
