package cache

import (
	"context"
	stderrors "errors"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// latestQubicDataID is the id of the general data document the processor keeps up to date.
const latestQubicDataID = "latest"

type ServiceConfiguration struct {
	MongoDatabase             string
	MongoQubicDataCollection  string
	MongoRichListCollection   string
	MongoEpochStatsCollection string

	CacheValidityDuration time.Duration

	RichListPageSize int32

	CacheUpdateTimeout time.Duration
}

type Service struct {
	Cache *Cache

	mongoClient               *mongo.Client
	mongoDatabase             string
	mongoQubicDataCollection  string
	mongoRichListCollection   string
	mongoEpochStatsCollection string

	cacheValidityDuration time.Duration

	cacheUpdateTimeout time.Duration

	richListPageSize int32
}

func NewCacheService(configuration *ServiceConfiguration, mongoClient *mongo.Client) *Service {
	return &Service{
		Cache: &Cache{},

		mongoClient:               mongoClient,
		mongoDatabase:             configuration.MongoDatabase,
		mongoQubicDataCollection:  configuration.MongoQubicDataCollection,
		mongoRichListCollection:   configuration.MongoRichListCollection,
		mongoEpochStatsCollection: configuration.MongoEpochStatsCollection,

		cacheValidityDuration: configuration.CacheValidityDuration,

		richListPageSize: configuration.RichListPageSize,

		cacheUpdateTimeout: configuration.CacheUpdateTimeout,
	}

}

func (s *Service) Start() chan bool {
	exit := make(chan bool)
	println("Starting Cache service...")
	go func() {
		ticker := time.NewTicker(time.Microsecond)
		for range ticker.C {
			ticker.Reset(s.cacheValidityDuration)
			println("Updating...")

			err := s.updateCache()
			if err != nil {
				fmt.Printf("Failed to update Cache. Error: %v\n", err)
				continue
			}
			println("Done updating.")
		}
		exit <- true
	}()

	return exit
}

// updateCache reloads the latest data and the supply history. The two are independent, so failing
// to load one still refreshes the other and keeps the previously cached one.
func (s *Service) updateCache() error {

	ctx, cancel := context.WithTimeout(context.Background(), s.cacheUpdateTimeout)
	defer cancel()

	var errs []error

	qubicData, err := s.fetchQubicData(ctx)
	if err != nil {
		errs = append(errs, errors.Wrap(err, "fetching qubic data"))
	} else {
		s.Cache.UpdateQubicData(qubicData)
	}

	// The supply history is small, so it is simply reloaded in full.
	supplyHistory, err := s.fetchSupplyHistory(ctx)
	if err != nil {
		errs = append(errs, errors.Wrap(err, "fetching supply history"))
	} else {
		s.Cache.UpdateSupplyHistory(supplyHistory)
	}

	return stderrors.Join(errs...)
}

// fetchSupplyHistory loads every epoch stats record, oldest epoch first. An empty collection is not
// an error: the records only appear once the processor has parsed a spectrum file.
func (s *Service) fetchSupplyHistory(ctx context.Context) (SupplyHistory, error) {
	collection := s.mongoClient.Database(s.mongoDatabase).Collection(s.mongoEpochStatsCollection)

	opts := options.Find().SetSort(bson.D{{Key: "_id", Value: 1}})

	cursor, err := collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, errors.Wrap(err, "querying epoch stats")
	}

	var supplyHistory SupplyHistory
	if err := cursor.All(ctx, &supplyHistory); err != nil {
		return nil, errors.Wrap(err, "decoding epoch stats")
	}

	return supplyHistory, nil
}

func (s *Service) fetchQubicData(ctx context.Context) (QubicData, error) {
	collection := s.mongoClient.Database(s.mongoDatabase).Collection(s.mongoQubicDataCollection)

	var qubicData QubicData

	// The processor keeps a single document, which it overwrites on every scrape.
	result := collection.FindOne(ctx, bson.D{{Key: "_id", Value: latestQubicDataID}})
	err := result.Decode(&qubicData)
	if err != nil {
		return QubicData{}, errors.Wrap(err, "decoding database response")
	}

	return qubicData, nil
}
