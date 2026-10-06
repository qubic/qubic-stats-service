// Package epochstats stores one record per completed epoch: the circulating supply and the active
// addresses at the end of the epoch, the cumulative issuance up to it and when it ended.
//
// A record is derived from nothing but a spectrum file and its name, so parsing the same file again
// always yields the same record, no matter when it happens.
package epochstats

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// EmissionPerEpoch is the fixed amount of QUs that is emitted at the end of every epoch.
const EmissionPerEpoch int64 = 1_000_000_000_000

const (
	// epochDuration is the length of an epoch in seconds. Epochs change every Wednesday at 12:00 UTC.
	epochDuration int64 = 7 * 24 * 60 * 60
	// referenceEpoch ended at referenceEpochEnd (Wednesday, 2026-09-23 12:00 UTC). Every other epoch
	// boundary is a whole number of weeks away from it.
	referenceEpoch    uint32 = 231
	referenceEpochEnd int64  = 1790164800
)

// Record is one data point of the per epoch supply history.
type Record struct {
	Epoch             uint32 `bson:"_id"`
	CirculatingSupply int64  `bson:"circulatingSupply"`
	TotalEmitted      int64  `bson:"totalEmitted"`
	ActiveAddresses   int    `bson:"activeAddresses"`
	EpochEndTimestamp int64  `bson:"epochEndTimestamp"`
}

// BurnedQUs is the amount of QUs that were emitted up to the end of the epoch but are no longer in
// circulation.
func (r Record) BurnedQUs() int64 {
	return r.TotalEmitted - r.CirculatingSupply
}

// NewRecord builds the record that the spectrum file of the given epoch stands for.
//
// A spectrum file named for epoch N is dumped at the transition into N and holds the end state of
// epoch N-1, so it is the record of N-1. The epoch in progress has no record until it has closed.
func NewRecord(spectrumEpoch uint32, circulatingSupply int64, activeAddresses int) (Record, error) {

	if spectrumEpoch < 2 {
		return Record{}, fmt.Errorf("spectrum epoch %d does not follow a completed epoch", spectrumEpoch)
	}
	epoch := spectrumEpoch - 1

	return Record{
		Epoch:             epoch,
		CirculatingSupply: circulatingSupply,
		TotalEmitted:      TotalEmitted(epoch),
		ActiveAddresses:   activeAddresses,
		EpochEndTimestamp: EpochEndTimestamp(epoch),
	}, nil
}

// TotalEmitted returns the cumulative issuance up to and including the given epoch.
func TotalEmitted(epoch uint32) int64 {
	return int64(epoch) * EmissionPerEpoch
}

// EpochEndTimestamp returns, in unix seconds, when the given epoch ended. This is the scheduled
// transition, Wednesday 12:00 UTC.
func EpochEndTimestamp(epoch uint32) int64 {
	return referenceEpochEnd + (int64(epoch)-int64(referenceEpoch))*epochDuration
}

// Save writes the record of an epoch, replacing whatever was stored for it.
func Save(ctx context.Context, client *mongo.Client, database, collectionName string, record Record) error {

	collection := client.Database(database).Collection(collectionName)
	_, err := collection.ReplaceOne(
		ctx,
		bson.D{{Key: "_id", Value: record.Epoch}},
		record,
		options.Replace().SetUpsert(true),
	)
	if err != nil {
		return fmt.Errorf("saving epoch stats record for epoch %d: %w", record.Epoch, err)
	}

	return nil
}

// LoadLatest returns the record of the most recent epoch. It reports false when there is none yet.
func LoadLatest(ctx context.Context, client *mongo.Client, database, collectionName string) (Record, bool, error) {

	collection := client.Database(database).Collection(collectionName)

	var record Record
	err := collection.FindOne(ctx, bson.D{}, options.FindOne().SetSort(bson.D{{Key: "_id", Value: -1}})).Decode(&record)
	if err == mongo.ErrNoDocuments {
		return Record{}, false, nil
	}
	if err != nil {
		return Record{}, false, fmt.Errorf("loading latest epoch stats record: %w", err)
	}

	return record, true, nil
}
