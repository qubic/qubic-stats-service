package main

import (
	"fmt"
	"github.com/ardanlabs/conf"
	"github.com/pkg/errors"
	"github.com/qubic/qubic-stats-api/db"
	"github.com/qubic/qubic-stats-api/rpc"
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
		}
	}

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

	mongoConnection := db.Connection{
		Username:          config.Mongo.Username,
		Password:          config.Mongo.Password,
		Hostname:          config.Mongo.Hostname,
		Port:              config.Mongo.Port,
		ConnectionOptions: config.Mongo.Options,
	}

	configuration := rpc.ServerConfiguration{
		HttpAddress:        config.Service.HttpAddress,
		GrpcAddress:        config.Service.GrpcAddress,
		Connection:         &mongoConnection,
		Database:           config.Mongo.Database,
		SpectrumCollection: config.Mongo.SpectrumCollection,
		DataCollection:     config.Mongo.DataCollection,
		//CacheStrategy:              config.Service.CachingStrategy,
		CacheValidityDuration:      config.Service.CacheValidityDuration,
		SpectrumDataUpdateInterval: config.Service.SpectrumDataUpdateInterval,
	}

	server := rpc.NewServer(&configuration)
	err = server.Start()
	if err != nil {
		return errors.Wrap(err, "starting the web server")
	}

	return nil
}
