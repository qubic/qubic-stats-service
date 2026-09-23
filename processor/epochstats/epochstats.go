// Package epochstats stores one record per epoch: the circulating supply at the start of the epoch,
// the cumulative issuance up to it and when the epoch started. Records are keyed by epoch, so a
// record becomes immutable once its epoch has closed and writing one is idempotent.
package epochstats

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// EmissionPerEpoch is the fixed amount of QUs that is emitted in every epoch.
const EmissionPerEpoch int64 = 1_000_000_000_000

// Provenance of the circulating supply of a record.
const (
	// SupplySourceSpectrum means the supply was measured from the spectrum file of that same epoch.
	SupplySourceSpectrum = "spectrum"
	// SupplySourceCarriedOver means no spectrum file could be attributed to the epoch, so the supply
	// of an earlier epoch is repeated. Such a record is an estimate and shows no burn for its epoch.
	// Only the backfill produces these; a live epoch without a measurement gets no record at all.
	SupplySourceCarriedOver = "carried-over"
	// SupplySourceDerived means the supply was reconstructed from the historical general data, where
	// only the burned QUs were kept. The value is exact, but which epoch it was measured in is not
	// known, so it may be carried over from an earlier epoch.
	SupplySourceDerived = "derived"
)

// Provenance of the epoch start timestamp of a record.
const (
	// TimestampSourceTick means the timestamp comes from the first tick of the epoch that holds tick
	// data, which is the closest we get to the real epoch boundary.
	TimestampSourceTick = "tick"
	// TimestampSourceFirstObserved means the timestamp is when the epoch transition was first
	// noticed rather than a real boundary, which is accurate to the scrape interval at best and off
	// by the downtime if the service was not running across the transition.
	TimestampSourceFirstObserved = "first-observed"
)

// Record is one data point of the per epoch supply history.
type Record struct {
	Epoch             uint32 `bson:"_id"`
	CirculatingSupply int64  `bson:"circulatingSupply"`
	TotalEmitted      int64  `bson:"totalEmitted"`
	ActiveAddresses   int    `bson:"activeAddresses"`
	// EpochEndTimestamp is when the epoch closed, which is the boundary its supply was measured at.
	EpochEndTimestamp int64  `bson:"epochEndTimestamp"`
	SpectrumTimestamp int64  `bson:"spectrumTimestamp"`
	SupplySource      string `bson:"supplySource"`
	TimestampSource   string `bson:"timestampSource"`
	UpdatedAt         int64  `bson:"updatedAt"`
}

// TotalEmitted returns the cumulative issuance up to and including the given epoch.
func TotalEmitted(epoch uint32) int64 {
	return int64(epoch) * EmissionPerEpoch
}

// MeasuredEpoch returns the epoch a spectrum file stands for.
//
// A spectrum file named for epoch N is dumped at the transition into N and holds the end state of
// epoch N-1, so it is the authoritative measurement of epoch N-1 and pairs with the issuance of
// N-1 epochs. The epoch that is in progress has no measurement of its own until it too has closed.
//
// It reports false when the spectrum epoch is unknown or too low to name a completed epoch.
func MeasuredEpoch(spectrumEpoch uint32) (uint32, bool) {
	if spectrumEpoch < 2 {
		return 0, false
	}
	return spectrumEpoch - 1, true
}

// Upsert writes the record of an epoch, creating it if it does not exist yet.
//
// An epoch start timestamp that was resolved from tick data replaces whatever is stored. An
// estimated one is only written when the record is created, so that a good timestamp is never
// replaced by an estimate once it has been established.
func Upsert(ctx context.Context, client *mongo.Client, database, collectionName string, record Record) error {

	supplyFields := bson.D{
		{Key: "circulatingSupply", Value: record.CirculatingSupply},
		{Key: "totalEmitted", Value: record.TotalEmitted},
		{Key: "activeAddresses", Value: record.ActiveAddresses},
		{Key: "spectrumTimestamp", Value: record.SpectrumTimestamp},
		{Key: "supplySource", Value: record.SupplySource},
		{Key: "updatedAt", Value: record.UpdatedAt},
	}

	timestampFields := bson.D{
		{Key: "epochEndTimestamp", Value: record.EpochEndTimestamp},
		{Key: "timestampSource", Value: record.TimestampSource},
	}

	var update bson.D
	if record.TimestampSource == TimestampSourceTick {
		update = bson.D{{Key: "$set", Value: append(supplyFields, timestampFields...)}}
	} else {
		update = bson.D{
			{Key: "$set", Value: supplyFields},
			{Key: "$setOnInsert", Value: timestampFields},
		}
	}

	collection := client.Database(database).Collection(collectionName)
	_, err := collection.UpdateOne(
		ctx,
		bson.D{{Key: "_id", Value: record.Epoch}},
		update,
		options.UpdateOne().SetUpsert(true),
	)
	if err != nil {
		return fmt.Errorf("upserting epoch stats record for epoch %d: %w", record.Epoch, err)
	}

	return nil
}

// InsertMissing creates the given records but never touches records that already exist. The backfill
// uses it so that re-running it cannot downgrade a record that the service has meanwhile written
// from a real spectrum measurement. It returns the number of records that were created.
func InsertMissing(ctx context.Context, client *mongo.Client, database, collectionName string, records []Record) (int, error) {

	if len(records) == 0 {
		return 0, nil
	}

	models := make([]mongo.WriteModel, 0, len(records))
	for _, record := range records {
		models = append(models, mongo.NewUpdateOneModel().
			SetFilter(bson.D{{Key: "_id", Value: record.Epoch}}).
			SetUpdate(bson.D{{Key: "$setOnInsert", Value: bson.D{
				{Key: "circulatingSupply", Value: record.CirculatingSupply},
				{Key: "totalEmitted", Value: record.TotalEmitted},
				{Key: "activeAddresses", Value: record.ActiveAddresses},
				{Key: "epochEndTimestamp", Value: record.EpochEndTimestamp},
				{Key: "spectrumTimestamp", Value: record.SpectrumTimestamp},
				{Key: "supplySource", Value: record.SupplySource},
				{Key: "timestampSource", Value: record.TimestampSource},
				{Key: "updatedAt", Value: record.UpdatedAt},
			}}}).
			SetUpsert(true))
	}

	collection := client.Database(database).Collection(collectionName)
	result, err := collection.BulkWrite(ctx, models, options.BulkWrite().SetOrdered(false))
	if err != nil {
		return 0, fmt.Errorf("bulk inserting %d epoch stats records: %w", len(records), err)
	}

	return int(result.UpsertedCount), nil
}
