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
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var EmptyAddress [32]byte

type Data struct {
	// Epoch is the epoch the spectrum file was dumped at, which is the epoch the measured supply
	// belongs to. Documents written before this field was introduced hold a zero epoch.
	Epoch             uint32
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

	rankRichList(richList)

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

// rankRichList orders the rich list by balance, highest first, and records each entry's position.
// Equal balances are broken by identity so that a rank stays the same across re-parses of the same
// spectrum file. The api pages through the stored rank instead of sorting the collection.
func rankRichList(richList RichList) {

	slices.SortFunc(richList, func(a, b RichListEntity) int {
		if c := -cmp.Compare(a.Balance, b.Balance); c != 0 {
			return c
		}
		return cmp.Compare(a.Identity, b.Identity)
	})

	for index := range richList {
		richList[index].Rank = int32(index)
	}
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

		if entity.publicKey != EmptyAddress {
			spectrum = append(spectrum, entity)
		}

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

	var result Data

	opts := options.FindOne().SetSort(bson.D{{Key: "timestamp", Value: -1}})

	err := collection.FindOne(ctx, bson.D{}, opts).Decode(&result)
	if err != nil {
		return nil, errors.Wrap(err, "getting spectrum data from database")
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

	// Built after the insert, which is cheaper than maintaining it while writing half a million
	// entries. The api pages through the list by rank.
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "rank", Value: 1}},
	})
	if err != nil {
		return errors.Wrap(err, "creating rich list rank index")
	}

	println("Done.")

	return nil
}
