package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ServiceConfiguration struct {
	MongoDatabase             string
	MongoSpectrumCollection   string
	MongoQubicDataCollection  string
	MongoRichListCollection   string
	MongoEpochStatsCollection string

	CacheValidityDuration    time.Duration
	SpectrumValidityDuration time.Duration

	RichListPageSize int32

	CacheUpdateTimeout time.Duration
}

type Service struct {
	Cache *Cache

	mongoClient               *mongo.Client
	mongoDatabase             string
	mongoSpectrumCollection   string
	mongoQubicDataCollection  string
	mongoRichListCollection   string
	mongoEpochStatsCollection string

	cacheValidityDuration    time.Duration
	spectrumValidityDuration time.Duration

	cacheUpdateTimeout time.Duration

	richListPageSize int32
}

func NewCacheService(configuration *ServiceConfiguration, mongoClient *mongo.Client) *Service {
	return &Service{
		Cache: &Cache{},

		mongoClient:               mongoClient,
		mongoDatabase:             configuration.MongoDatabase,
		mongoSpectrumCollection:   configuration.MongoSpectrumCollection,
		mongoQubicDataCollection:  configuration.MongoQubicDataCollection,
		mongoRichListCollection:   configuration.MongoRichListCollection,
		mongoEpochStatsCollection: configuration.MongoEpochStatsCollection,

		cacheValidityDuration:    configuration.CacheValidityDuration,
		spectrumValidityDuration: configuration.SpectrumValidityDuration,

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

			lastSpectrumDataUpdate := s.Cache.GetLastSpectrumDataUpdate()
			nextSpectrumUpdate := lastSpectrumDataUpdate.Add(s.spectrumValidityDuration)

			// Either the refresh interval has passed, or nothing has been cached yet.
			updateSpectrum := !nextSpectrumUpdate.After(time.Now()) || s.Cache.GetSpectrumData().CirculatingSupply == 0

			err := s.updateCache(updateSpectrum, true)
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

func (s *Service) updateCache(updateSpectrumData bool, updateQubicData bool) error {
	var qubicData QubicData
	var spectrumData SpectrumData
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), s.cacheUpdateTimeout)
	defer cancel()

	if updateQubicData {
		println("Updated Qubic data")
		qubicData, err = s.fetchQubicData(ctx)
		if err != nil {
			return errors.Wrap(err, "fetching qubic data")
		}
	}

	if updateSpectrumData {
		spectrumData, err = s.fetchSpectrumData(ctx)
		println("Updated spectrum data")
		if err != nil {
			return errors.Wrap(err, "fetching spectrum data")
		}
	}

	s.Cache.UpdateDataCache(spectrumData, qubicData)

	// The supply history is small and immutable per epoch, so it is simply reloaded in full. Failing
	// to load it must not hold back the rest of the cache.
	supplyHistory, err := s.fetchSupplyHistory(ctx)
	if err != nil {
		fmt.Printf("Failed to update supply history. Error: %v\n", err)
	} else {
		s.Cache.UpdateSupplyHistory(supplyHistory)
	}

	return nil
}

// fetchSupplyHistory loads every epoch stats record, oldest epoch first. An empty collection is not
// an error: the records only appear once the processor has written or backfilled them.
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

func (s *Service) fetchSpectrumData(ctx context.Context) (SpectrumData, error) {
	collection := s.mongoClient.Database(s.mongoDatabase).Collection(s.mongoSpectrumCollection)

	var spectrumData SpectrumData

	opts := options.FindOne().SetSort(bson.M{"$natural": -1})

	result := collection.FindOne(ctx, bson.D{}, opts)
	err := result.Decode(&spectrumData)
	if err != nil {
		return SpectrumData{}, errors.Wrap(err, "decoding database response")
	}

	return spectrumData, nil
}

func (s *Service) fetchQubicData(ctx context.Context) (QubicData, error) {
	collection := s.mongoClient.Database(s.mongoDatabase).Collection(s.mongoQubicDataCollection)

	var qubicData QubicData

	opts := options.FindOne().SetSort(bson.M{"$natural": -1})

	result := collection.FindOne(ctx, bson.D{}, opts)
	err := result.Decode(&qubicData)
	if err != nil {
		return QubicData{}, errors.Wrap(err, "decoding database response")
	}

	return qubicData, nil
}
