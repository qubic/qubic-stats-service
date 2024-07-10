package spectrum

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"os"
	"time"
)

var EmptyAddress = [32]byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
}

type Data struct {
	CirculatingSupply int64 `json:"circulating_supply"`
	ActiveAddresses   int   `json:"active_addresses"`
	Timestamp         int64 `json:"unix_timestamp"`
}

type Spectrum []Entity

func (s Spectrum) CalculateSpectrumData() (*Data, error) {

	println("Calculating spectrum data...")

	var circulatingSupply int64
	var activeAddresses int

	for index, entity := range s {
		entityBalance, err := entity.GetBalance()
		if err != nil {
			return nil, errors.Wrapf(err, "getting balance of entity #%d", index)
		}
		circulatingSupply += entityBalance

		if entity.publicKey != EmptyAddress {
			activeAddresses += 1
		}

	}

	println("Done.")
	fmt.Printf("Circulating supply: %d\n", circulatingSupply)
	fmt.Printf("Active addresses: %d\n", activeAddresses)

	return &Data{
		CirculatingSupply: circulatingSupply,
		ActiveAddresses:   activeAddresses,
		Timestamp:         time.Now().Unix(),
	}, nil
}

func ReadSpectrumFromFile(filePath string, spectrumSize int64) (*Spectrum, error) {

	println("Reading spectrum...")

	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrap(err, "reading spectrum file")
	}

	if len(fileData)%64 != 0 {
		return nil, errors.New("Spectrum file may be incomplete! fileLength % 64 != 0")
	}

	fmt.Printf("Spectrum file size: %d\n", len(fileData))

	var spectrum Spectrum

	for index := range spectrumSize {

		beginningIndex := index * 64
		endingIndex := beginningIndex + 64

		dataSlice := fileData[beginningIndex:endingIndex]

		reader := bytes.NewReader(dataSlice)

		var entity Entity
		err := entity.UnmarshallFromBinary(reader)
		if err != nil {
			return nil, errors.Wrapf(err, "deserializing %dth spectrum entity", index)
		}

		spectrum = append(spectrum, entity)
	}

	return &spectrum, nil
}

func LoadSpectrumDataFromFile(spectrumDataFile string) (*Data, error) {

	println("Loading spectrum data...")

	data, err := os.ReadFile(spectrumDataFile)
	if err != nil {
		return nil, errors.Wrap(err, "loading spectrum data file")
	}

	var spectrumData Data

	err = json.Unmarshal(data, &spectrumData)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshalling spectrum data from json")
	}

	println("Spectrum data loaded.")
	fmt.Printf("Circulating supply: %d\n", spectrumData.CirculatingSupply)
	fmt.Printf("Active addresses: %d\n", spectrumData.ActiveAddresses)

	return &spectrumData, nil

}

func (d *Data) SaveSpectrumDataToFile(spectrumDataFile string) error {

	println("Saving spectrum data to file...")

	data, err := json.Marshal(d)
	if err != nil {
		return errors.Wrap(err, "marshalling spectrum data to json")
	}

	_ = os.Remove(spectrumDataFile) // Ignore error on purpose

	file, err := os.Create(spectrumDataFile)
	if err != nil {
		return errors.Wrap(err, "creating spectrum data file")
	}
	_, err = file.Write(data)
	if err != nil {
		return errors.Wrap(err, "saving to spectrum data file")
	}

	return nil

}

func LoadSpectrumDataFromDatabase(dbClient *mongo.Client, database string, spectrumCollection string) (*Data, error) {

	println("Loading spectrum data from database...")

	collection := dbClient.Database(database).Collection(spectrumCollection)

	cursor, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		return nil, errors.Wrap(err, "getting spectrum data from database")
	}

	var results []Data
	var result Data
	var latestTimestamp int64

	if err = cursor.All(context.Background(), &results); err != nil {
		return nil, errors.Wrap(err, "getting spectrum data from cursor")
	}

	for _, data := range results {
		if data.Timestamp > latestTimestamp {
			result = data
		}
	}

	println("Done.")

	return &result, nil

}

func (d *Data) SaveSpectrumDataToDatabase(dbClient *mongo.Client, database string, spectrumCollection string) error {

	println("Saving spectrum data to database...")

	collection := dbClient.Database(database).Collection(spectrumCollection)

	_, err := collection.InsertOne(context.Background(), d)
	if err != nil {
		return errors.Wrap(err, "saving spectrum data to database")
	}

	println("Done.")

	return nil
}
