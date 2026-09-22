package epochstats

import (
	"context"
	"fmt"
	"slices"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// The general data documents carry no bson tags, so the driver stored them under the lower cased go
// field names.
const (
	generalDataEpochField     = "epoch"
	generalDataTimestampField = "timestamp"
	generalDataBurnedField    = "burnedqus"
)

// epochAggregate is the per epoch summary of the historical general data.
type epochAggregate struct {
	Epoch          uint32 `bson:"_id"`
	FirstTimestamp int64  `bson:"firstTimestamp"`
	LastTimestamp  int64  `bson:"lastTimestamp"`
	BurnedQus      int64  `bson:"burnedQus"`
}

// spectrumMeasurement is the part of a spectrum data document the backfill needs. Historical
// documents carry no epoch, which is why they have to be matched by timestamp.
//
// spectrum.Data carries no bson tags, so the driver stored its fields under the lower cased go field
// names. The tags below have to spell them exactly that way.
type spectrumMeasurement struct {
	CirculatingSupply int64 `bson:"circulatingsupply"`
	ActiveAddresses   int   `bson:"activeaddresses"`
	Timestamp         int64 `bson:"timestamp"`
}

// BackfillReport describes what the backfill found, so that the result can be checked before the
// historical general data it was reconstructed from is expired.
type BackfillReport struct {
	Records []Record
	// MissingEpochs are epochs between the first and the last reconstructed one that the general data
	// holds no document for, most likely because the processor was not running.
	MissingEpochs []uint32
	// CarriedOverEpochs are epochs that no spectrum file was parsed during, so their supply is
	// repeated from an earlier epoch and their burn reads as zero.
	CarriedOverEpochs []uint32
	// SupplyMismatches are epochs where the spectrum measurement disagrees with the supply derived
	// from the burned QUs. They point at a real data problem and should be looked at.
	SupplyMismatches []uint32
}

// BuildBackfill reconstructs one record per epoch from the historical general data.
//
// The general data never stored the circulating supply, only the burned QUs, which the service
// computed as epoch * EmissionPerEpoch - circulatingSupply. Inverting that is exact.
//
// The supply in force during an epoch A comes from the spectrum file named A, which holds the end
// state of epoch A-1, so the reconstructed value is the supply of epoch A-1 and is recorded under
// that epoch. The spectrum data is joined in by timestamp to recover the active addresses and to
// tell the epochs that were really measured from the ones that only repeat an earlier supply.
func BuildBackfill(ctx context.Context, client *mongo.Client, database, generalDataCollection, spectrumCollection string) (BackfillReport, error) {

	aggregates, err := aggregateGeneralDataPerEpoch(ctx, client, database, generalDataCollection)
	if err != nil {
		return BackfillReport{}, err
	}
	if len(aggregates) == 0 {
		return BackfillReport{}, nil
	}

	measurements, err := loadSpectrumMeasurements(ctx, client, database, spectrumCollection)
	if err != nil {
		return BackfillReport{}, err
	}

	now := time.Now().Unix()
	report := BackfillReport{Records: make([]Record, 0, len(aggregates))}

	recordEpochs := make([]uint32, 0, len(aggregates))

	for _, aggregate := range aggregates {

		// The burned QUs of epoch A were computed against the supply that A ran on, which the
		// spectrum file named A measured at the end of epoch A-1.
		measuredEpoch, ok := MeasuredEpoch(aggregate.Epoch)
		if !ok {
			continue
		}

		record := Record{
			Epoch:             measuredEpoch,
			CirculatingSupply: TotalEmitted(aggregate.Epoch) - aggregate.BurnedQus,
			TotalEmitted:      TotalEmitted(measuredEpoch),
			// Epoch A-1 closed where epoch A began.
			EpochEndTimestamp: aggregate.FirstTimestamp,
			SupplySource:      SupplySourceCarriedOver,
			TimestampSource:   TimestampSourceFirstObserved,
			UpdatedAt:         now,
		}

		// A spectrum file parsed during epoch A is the measurement of epoch A-1.
		if measurement, found := measurementWithin(measurements, aggregate.FirstTimestamp, aggregate.LastTimestamp); found {
			record.SupplySource = SupplySourceDerived
			record.ActiveAddresses = measurement.ActiveAddresses
			record.SpectrumTimestamp = measurement.Timestamp
			if measurement.CirculatingSupply != record.CirculatingSupply {
				report.SupplyMismatches = append(report.SupplyMismatches, measuredEpoch)
			}
		} else {
			report.CarriedOverEpochs = append(report.CarriedOverEpochs, measuredEpoch)
		}

		report.Records = append(report.Records, record)
		recordEpochs = append(recordEpochs, measuredEpoch)
	}

	report.MissingEpochs = missingEpochs(recordEpochs)

	return report, nil
}

// aggregateGeneralDataPerEpoch reduces the general data to one summary per epoch. The burned QUs are
// taken from the last document of an epoch, because an epoch starts out with the supply of the
// previous one and only picks up its own once the spectrum file has been parsed.
func aggregateGeneralDataPerEpoch(ctx context.Context, client *mongo.Client, database, collectionName string) ([]epochAggregate, error) {

	pipeline := mongo.Pipeline{
		// epoch 0 means the document predates the epoch field or was written before the archiver
		// answered, and holds nothing we can attribute.
		bson.D{{Key: "$match", Value: bson.D{{Key: generalDataEpochField, Value: bson.D{{Key: "$gt", Value: 0}}}}}},
		bson.D{{Key: "$sort", Value: bson.D{
			{Key: generalDataEpochField, Value: 1},
			{Key: generalDataTimestampField, Value: 1},
		}}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$" + generalDataEpochField},
			{Key: "firstTimestamp", Value: bson.D{{Key: "$first", Value: "$" + generalDataTimestampField}}},
			{Key: "lastTimestamp", Value: bson.D{{Key: "$last", Value: "$" + generalDataTimestampField}}},
			{Key: "burnedQus", Value: bson.D{{Key: "$last", Value: "$" + generalDataBurnedField}}},
		}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "_id", Value: 1}}}},
	}

	collection := client.Database(database).Collection(collectionName)
	// The general data is neither indexed nor small, but the backfill runs once and offline.
	cursor, err := collection.Aggregate(ctx, pipeline, options.Aggregate().SetAllowDiskUse(true))
	if err != nil {
		return nil, fmt.Errorf("aggregating general data per epoch: %w", err)
	}

	var aggregates []epochAggregate
	if err := cursor.All(ctx, &aggregates); err != nil {
		return nil, fmt.Errorf("decoding general data aggregate: %w", err)
	}

	return aggregates, nil
}

func loadSpectrumMeasurements(ctx context.Context, client *mongo.Client, database, collectionName string) ([]spectrumMeasurement, error) {

	collection := client.Database(database).Collection(collectionName)
	cursor, err := collection.Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{Key: "timestamp", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("loading spectrum measurements: %w", err)
	}

	var measurements []spectrumMeasurement
	if err := cursor.All(ctx, &measurements); err != nil {
		return nil, fmt.Errorf("decoding spectrum measurements: %w", err)
	}

	return measurements, nil
}

// measurementWithin returns the last spectrum measurement that was taken between the given
// timestamps. The last one wins because a spectrum file may be parsed more than once per epoch.
func measurementWithin(measurements []spectrumMeasurement, from, to int64) (spectrumMeasurement, bool) {

	var found bool
	var latest spectrumMeasurement

	for _, measurement := range measurements {
		if measurement.Timestamp >= from && measurement.Timestamp <= to {
			latest = measurement
			found = true
		}
	}

	return latest, found
}

// missingEpochs returns the epochs between the first and the last given one that are absent.
func missingEpochs(epochs []uint32) []uint32 {

	if len(epochs) == 0 {
		return nil
	}

	present := make(map[uint32]bool, len(epochs))
	for _, epoch := range epochs {
		present[epoch] = true
	}

	var missing []uint32
	for epoch := slices.Min(epochs); epoch <= slices.Max(epochs); epoch++ {
		if !present[epoch] {
			missing = append(missing, epoch)
		}
	}

	return missing
}
