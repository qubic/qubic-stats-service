package cache

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type ServiceConfiguration struct {
	MongoDatabase            string
	MongoSpectrumCollection  string
	MongoQubicDataCollection string

	CacheValidityDuration    time.Duration
	SpectrumValidityDuration time.Duration
}

type Service struct {
	Cache *Cache

	mongoClient              *mongo.Client
	mongoDatabase            string
	mongoSpectrumCollection  string
	mongoQubicDataCollection string

	cacheValidityDuration    time.Duration
	spectrumValidityDuration time.Duration
}

func NewCacheService(configuration *ServiceConfiguration, mongoClient *mongo.Client) *Service {
	return &Service{
		Cache: &Cache{},

		mongoClient:              mongoClient,
		mongoDatabase:            configuration.MongoDatabase,
		mongoSpectrumCollection:  configuration.MongoSpectrumCollection,
		mongoQubicDataCollection: configuration.MongoQubicDataCollection,

		cacheValidityDuration:    configuration.CacheValidityDuration,
		spectrumValidityDuration: configuration.SpectrumValidityDuration,
	}

}

func (s *Service) Start() chan bool {

	exit := make(chan bool)

	println("Starting Cache service...")
	go func() {
		ticker := time.NewTicker(time.Second)
		for range ticker.C {
			ticker.Reset(s.cacheValidityDuration)
			println("Updating...")

			lastSpectrumDataUpdate := s.Cache.GetLastSpectrumDataUpdate()
			nextSpectrumUpdate := lastSpectrumDataUpdate.Add(s.spectrumValidityDuration)

			updateSpectrum := false
			if nextSpectrumUpdate.Compare(time.Now()) > 0 {
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

	var qubicData *QubicData
	var spectrumData *SpectrumData
	var err error

	if updateQubicData {
		qubicData, err = s.fetchQubicData()
		if err != nil {
			return errors.Wrap(err, "fetching qubic data")
		}
	}

	if updateSpectrumData {
		spectrumData, err = s.fetchSpectrumData()
		if err != nil {
			return errors.Wrap(err, "fetching spectrum data")
		}
	}

	s.Cache.UpdateDataCache(spectrumData, qubicData)

	return nil
}

func (s *Service) fetchSpectrumData() (*SpectrumData, error) {

	collection := s.mongoClient.Database(s.mongoDatabase).Collection(s.mongoSpectrumCollection)

	var spectrumData SpectrumData

	opts := options.FindOne().SetSort(bson.M{"$natural": -1})

	result := collection.FindOne(context.Background(), bson.D{}, opts)
	err := result.Decode(&spectrumData)
	if err != nil {
		return nil, errors.Wrap(err, "decoding database response")
	}

	return &spectrumData, nil
}

func (s *Service) fetchQubicData() (*QubicData, error) {
	collection := s.mongoClient.Database(s.mongoDatabase).Collection(s.mongoQubicDataCollection)

	var qubicData QubicData

	opts := options.FindOne().SetSort(bson.M{"$natural": -1})

	result := collection.FindOne(context.Background(), bson.D{}, opts)
	err := result.Decode(&qubicData)
	if err != nil {
		return nil, errors.Wrap(err, "decoding database response")
	}

	return &qubicData, nil
}
