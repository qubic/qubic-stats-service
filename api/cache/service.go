package cache

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strconv"
	"time"
)

type ServiceConfiguration struct {
	MongoDatabase            string
	MongoSpectrumCollection  string
	MongoQubicDataCollection string
	MongoRichListCollection  string

	CacheValidityDuration    time.Duration
	SpectrumValidityDuration time.Duration

	RichListPageSize int32
}

type Service struct {
	Cache *Cache

	mongoClient              *mongo.Client
	mongoDatabase            string
	mongoSpectrumCollection  string
	mongoQubicDataCollection string
	mongoRichListCollection  string

	cacheValidityDuration    time.Duration
	spectrumValidityDuration time.Duration

	richListPageSize int32
}

func NewCacheService(configuration *ServiceConfiguration, mongoClient *mongo.Client) *Service {
	return &Service{
		Cache: &Cache{},

		mongoClient:              mongoClient,
		mongoDatabase:            configuration.MongoDatabase,
		mongoSpectrumCollection:  configuration.MongoSpectrumCollection,
		mongoQubicDataCollection: configuration.MongoQubicDataCollection,
		mongoRichListCollection:  configuration.MongoRichListCollection,

		cacheValidityDuration:    configuration.CacheValidityDuration,
		spectrumValidityDuration: configuration.SpectrumValidityDuration,

		richListPageSize: configuration.RichListPageSize,
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

			updateSpectrum := false
			if nextSpectrumUpdate.Compare(time.Now()) > 0 || s.Cache.spectrumData.CirculatingSupply == 0 {
				updateSpectrum = true
			}

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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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

	err = s.calculateRichListCache(ctx)
	if err != nil {
		return errors.Wrap(err, "calculating rich list length and page count")
	}

	return nil
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

func (s *Service) calculateRichListCache(ctx context.Context) error {

	if s.Cache.GetRichListLength() == 0 || s.Cache.GetRichListPageCount() == 0 {

		epoch := strconv.Itoa(int(s.Cache.GetQubicData().Epoch))

		collection := s.mongoClient.Database(s.mongoDatabase).Collection(s.mongoRichListCollection + "_" + epoch)

		count, err := collection.CountDocuments(ctx, bson.D{}, options.Count().SetHint("_id_"))
		if err != nil {
			return errors.Wrap(err, "getting rich list length")
		}
		s.Cache.SetRichListLength(int32(count))

		pageCount := int32(count) / s.richListPageSize
		remainder := int32(count) % s.richListPageSize
		if remainder != 0 {
			pageCount += 1
		}
		s.Cache.SetRichListPageCount(pageCount)
	}

	return nil

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
