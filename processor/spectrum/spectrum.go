package spectrum

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/pkg/errors"
	"github.com/qubic/go-node-connector/types"
	"github.com/schollz/progressbar/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var EmptyAddress [32]byte

type Data struct {
	CirculatingSupply int64
	ActiveAddresses   int
	Timestamp         int64
}

type Results struct {
	Data Data
	List RichList
}

type Spectrum []Entity

func CalculateSpectrumData(spectrum *Spectrum) (*Results, error) {

	println("Calculating spectrum data...")

	var circulatingSupply int64
	var activeAddresses int

	var richList RichList
	spectrumLength := len(*spectrum)

	bar := progressbar.Default(int64(spectrumLength), "Processing spectrum data")

	for index, entity := range *spectrum {
		_ = bar.Add(1)
		entityBalance, err := entity.GetBalance()
		if err != nil {
			return nil, errors.Wrapf(err, "getting balance of entity #%d", index)
		}
		circulatingSupply += entityBalance

		if entity.publicKey != EmptyAddress {
			activeAddresses += 1

			var identity types.Identity
			identity, err := identity.FromPubKey(entity.publicKey, false)
			if err != nil {
				return nil, errors.Wrap(err, "getting identity for spectrum entity")
			}

			richListEntity := RichListEntity{
				Balance:  entityBalance,
				Identity: identity.String(),
			}
			richList = append(richList, richListEntity)
		}

	}

	slices.SortFunc(richList, func(a, b RichListEntity) int {
		return cmp.Compare(a.Balance, b.Balance)
	})

	println("Done.")
	fmt.Printf("Circulating supply: %d\n", circulatingSupply)
	fmt.Printf("Active addresses: %d\n", activeAddresses)

	return &Results{
		Data: Data{
			CirculatingSupply: circulatingSupply,
			ActiveAddresses:   activeAddresses,
			Timestamp:         time.Now().Unix(),
		},
		List: richList,
	}, nil
}

func ReadSpectrumFromFile(filePath string) (*Spectrum, error) {

	println("Reading spectrum...")

	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrap(err, "reading spectrum file")
	}

	spectrumFileLength := len(fileData)

	if spectrumFileLength%entitySize != 0 {
		return nil, fmt.Errorf("spectrum file may be incomplete. fileLength mod %d != 0", entitySize)
	}

	fmt.Printf("Spectrum file size: %d\n", spectrumFileLength)

	entityCount := spectrumFileLength / entitySize

	var spectrum Spectrum

	for index := range entityCount {

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

/*func LoadSpectrumDataFromFile(spectrumDataFile string) (*Data, error) {

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

}*/

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

func LoadSpectrumDataFromDatabase(ctx context.Context, dbClient *mongo.Client, database string, spectrumCollection string) (*Data, error) {

	println("Loading spectrum data from database...")

	collection := dbClient.Database(database).Collection(spectrumCollection)

	cursor, err := collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, errors.Wrap(err, "getting spectrum data from database")
	}

	var results []Data
	var result Data
	var latestTimestamp int64

	if err = cursor.All(ctx, &results); err != nil {
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

func (d *Data) SaveSpectrumDataToDatabase(ctx context.Context, dbClient *mongo.Client, database string, spectrumCollection string) error {

	println("Saving spectrum data to database...")

	collection := dbClient.Database(database).Collection(spectrumCollection)

	_, err := collection.InsertOne(ctx, d)
	if err != nil {
		return errors.Wrap(err, "saving spectrum data to database")
	}

	println("Done.")

	return nil
}

func SaveRichListToDatabase(ctx context.Context, dbClient *mongo.Client, database string, richListCollection string, richList RichList) error {
	println("Saving rich list to database...")

	collection := dbClient.Database(database).Collection(richListCollection)

	list := make([]interface{}, len(richList))
	for index, value := range richList {
		list[index] = value
	}

	_, err := collection.InsertMany(ctx, list)
	if err != nil {
		return errors.Wrap(err, "saving rich list to database")
	}

	println("Done.")

	return nil
}
