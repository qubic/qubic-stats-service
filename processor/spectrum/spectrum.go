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

	"github.com/qubic/go-node-connector/types"
	"github.com/schollz/progressbar/v3"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
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
			return nil, fmt.Errorf("getting balance of entity #%d: %w", index, err)
		}
		circulatingSupply += entityBalance

		if entity.publicKey != EmptyAddress {
			activeAddresses += 1

			var identity types.Identity
			identity, err := identity.FromPubKey(entity.publicKey, false)
			if err != nil {
				return nil, fmt.Errorf("getting identity for spectrum entity: %w", err)
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
		return nil, fmt.Errorf("reading spectrum file: %w", err)
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
			return nil, fmt.Errorf("deserializing %dth spectrum entity: %w", index, err)
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
		return nil, fmt.Errorf("loading spectrum data file: %w", err)
	}

	var spectrumData Data

	err = json.Unmarshal(data, &spectrumData)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling spectrum data from json: %w", err)
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
		return fmt.Errorf("marshalling spectrum data to json: %w", err)
	}

	_ = os.Remove(spectrumDataFile) // Ignore error on purpose

	file, err := os.Create(spectrumDataFile)
	if err != nil {
		return fmt.Errorf("creating spectrum data file: %w", err)
	}
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("saving to spectrum data file: %w", err)
	}

	return nil

}

// ReplaceRichList swaps the stored rich list for the given one.
//
// The new list is written to a staging collection and indexed first, then renamed over the live
// collection in a single step, so the api never sees a missing or half written rich list.
func ReplaceRichList(ctx context.Context, dbClient *mongo.Client, database string, richListCollection string, richList RichList) error {
	println("Saving rich list to database...")

	stagingCollectionName := richListCollection + "_staging"
	staging := dbClient.Database(database).Collection(stagingCollectionName)

	// A parse that failed half way may have left a staging collection behind.
	err := staging.Drop(ctx)
	if err != nil {
		return fmt.Errorf("dropping leftover rich list staging collection: %w", err)
	}

	list := make([]interface{}, len(richList))
	for index, value := range richList {
		list[index] = value
	}

	_, err = staging.InsertMany(ctx, list)
	if err != nil {
		return fmt.Errorf("saving rich list to staging collection: %w", err)
	}

	// Built after the insert, which is cheaper than maintaining it while writing half a million
	// entries. The api pages through the list by rank.
	_, err = staging.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "rank", Value: 1}},
	})
	if err != nil {
		return fmt.Errorf("creating rich list rank index: %w", err)
	}

	err = dbClient.Database("admin").RunCommand(ctx, bson.D{
		{Key: "renameCollection", Value: database + "." + stagingCollectionName},
		{Key: "to", Value: database + "." + richListCollection},
		{Key: "dropTarget", Value: true},
	}).Err()
	if err != nil {
		return fmt.Errorf("renaming rich list staging collection: %w", err)
	}

	println("Done.")

	return nil
}
