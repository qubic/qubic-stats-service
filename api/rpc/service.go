package rpc

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/qubic/qubic-stats-api/db"
	"github.com/qubic/qubic-stats-api/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func (s *Server) createDatabaseClient() error {
	client, err := db.CreateClient(s.connection)
	if err != nil {
		return errors.Wrap(err, "connecting to database")
	}
	s.dbClient = client
	return nil
}

func (s *Server) updateCache(updateSpectrumData bool, updateQubicData bool) error {

	var qubicData *types.QubicData
	var spectrumData *types.SpectrumData
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

	s.dataCache.UpdateDataCache(spectrumData, qubicData)

	return nil
}

func (s *Server) fetchSpectrumData() (*types.SpectrumData, error) {

	collection := s.dbClient.Database(s.database).Collection(s.spectrumCollection)

	var spectrumData types.SpectrumData

	opts := options.FindOne().SetSort(bson.M{"$natural": -1})

	result := collection.FindOne(context.Background(), bson.D{}, opts)
	err := result.Decode(&spectrumData)
	if err != nil {
		return nil, errors.Wrap(err, "decoding database response")
	}

	return &spectrumData, nil
}

func (s *Server) fetchQubicData() (*types.QubicData, error) {
	collection := s.dbClient.Database(s.database).Collection(s.dataCollection)

	var qubicData types.QubicData

	opts := options.FindOne().SetSort(bson.M{"$natural": -1})

	result := collection.FindOne(context.Background(), bson.D{}, opts)
	err := result.Decode(&qubicData)
	if err != nil {
		return nil, errors.Wrap(err, "decoding database response")
	}

	return &qubicData, nil

}

func (s *Server) startUpdateService() chan bool {

	exit := make(chan bool)

	println("Starting update service...")
	go func() {
		ticker := time.NewTicker(time.Second)
		for range ticker.C {
			ticker.Reset(s.cacheValidityDuration)
			println("Updating...")

			lastSpectrumDataUpdate := s.dataCache.getLastSpectrumDataUpdate()
			nextSpectrumUpdate := lastSpectrumDataUpdate.Add(s.spectrumDataUpdateInterval)

			updateSpectrum := false
			if nextSpectrumUpdate.Compare(time.Now()) > 0 {
				updateSpectrum = true
			}

			err := s.updateCache(updateSpectrum, true)
			if err != nil {
				fmt.Printf("Failed to update cache. Error: %v", err)
			}
		}
		exit <- true
	}()

	return exit

}
