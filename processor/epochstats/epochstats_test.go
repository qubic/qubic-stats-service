package epochstats

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_TotalEmitted(t *testing.T) {

	assert.Equal(t, int64(0), TotalEmitted(0))
	assert.Equal(t, int64(1_000_000_000_000), TotalEmitted(1))
	assert.Equal(t, int64(180_000_000_000_000), TotalEmitted(180))
}

func Test_MeasurementWithin_givenMeasurementInRange_thenFound(t *testing.T) {

	measurements := []spectrumMeasurement{
		{Timestamp: 100, CirculatingSupply: 1},
		{Timestamp: 250, CirculatingSupply: 2},
		{Timestamp: 400, CirculatingSupply: 3},
	}

	measurement, found := measurementWithin(measurements, 200, 300)

	assert.True(t, found)
	assert.Equal(t, int64(2), measurement.CirculatingSupply)
}

func Test_MeasurementWithin_givenSeveralInRange_thenLastWins(t *testing.T) {

	measurements := []spectrumMeasurement{
		{Timestamp: 210, CirculatingSupply: 1},
		{Timestamp: 290, CirculatingSupply: 2},
	}

	measurement, found := measurementWithin(measurements, 200, 300)

	assert.True(t, found)
	assert.Equal(t, int64(2), measurement.CirculatingSupply)
}

func Test_MeasurementWithin_givenBoundaries_thenInclusive(t *testing.T) {

	measurements := []spectrumMeasurement{{Timestamp: 200, CirculatingSupply: 7}}

	_, found := measurementWithin(measurements, 200, 300)
	assert.True(t, found)

	_, found = measurementWithin(measurements, 100, 200)
	assert.True(t, found)

	_, found = measurementWithin(measurements, 201, 300)
	assert.False(t, found)
}

func Test_MeasurementWithin_givenNoneInRange_thenNotFound(t *testing.T) {

	measurements := []spectrumMeasurement{{Timestamp: 100}, {Timestamp: 400}}

	_, found := measurementWithin(measurements, 200, 300)

	assert.False(t, found)
}

func Test_MissingEpochs(t *testing.T) {

	tests := []struct {
		name     string
		epochs   []uint32
		expected []uint32
	}{
		{name: "no gaps", epochs: []uint32{118, 119, 120}},
		{name: "one gap", epochs: []uint32{118, 120}, expected: []uint32{119}},
		{name: "several gaps", epochs: []uint32{118, 121, 124}, expected: []uint32{119, 120, 122, 123}},
		{name: "single epoch", epochs: []uint32{118}},
		{name: "nothing", epochs: []uint32{}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, missingEpochs(test.epochs))
		})
	}
}

// A spectrum file named for epoch N is dumped entering N and holds the end state of N-1, so it is
// the measurement of N-1. Getting this backwards pairs a supply with one epoch too much issuance.
func Test_MeasuredEpoch(t *testing.T) {

	epoch, ok := MeasuredEpoch(180)
	assert.True(t, ok)
	assert.Equal(t, uint32(179), epoch)

	// the supply measured by spectrum.180 belongs with the issuance of 179 epochs
	assert.Equal(t, int64(179_000_000_000_000), TotalEmitted(epoch))
}

func Test_MeasuredEpoch_givenNoNameableEpoch_thenNotOk(t *testing.T) {

	for _, spectrumEpoch := range []uint32{0, 1} {
		_, ok := MeasuredEpoch(spectrumEpoch)
		assert.False(t, ok, "spectrum epoch %d", spectrumEpoch)
	}
}
