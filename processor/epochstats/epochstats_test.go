package epochstats

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_TotalEmitted(t *testing.T) {

	assert.Equal(t, int64(0), TotalEmitted(0))
	assert.Equal(t, int64(1_000_000_000_000), TotalEmitted(1))
	assert.Equal(t, int64(180_000_000_000_000), TotalEmitted(180))
}

func Test_EpochEndTimestamp(t *testing.T) {

	tests := []struct {
		name     string
		epoch    uint32
		expected int64
	}{
		{name: "reference", epoch: 231, expected: 1790164800},       // Wed 2026-09-23 12:00 UTC
		{name: "next", epoch: 232, expected: 1790769600},            // Wed 2026-09-30 12:00 UTC
		{name: "previous", epoch: 230, expected: 1789560000},        // Wed 2026-09-16 12:00 UTC
		{name: "oldest spectrum", epoch: 185, expected: 1762344000}, // Wed 2025-11-05 12:00 UTC
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, EpochEndTimestamp(test.epoch))
		})
	}
}

// A spectrum file named for epoch N is dumped entering N and holds the end state of N-1. Getting
// this backwards pairs a supply with one epoch too much issuance.
func Test_NewRecord_thenRecordIsForThePreviousEpoch(t *testing.T) {

	record, err := NewRecord(232, 150_000_000_000_000, 500_000)
	require.NoError(t, err)

	assert.Equal(t, Record{
		Epoch:             231,
		CirculatingSupply: 150_000_000_000_000,
		TotalEmitted:      231_000_000_000_000,
		ActiveAddresses:   500_000,
		EpochEndTimestamp: 1790164800,
	}, record)
	assert.Equal(t, int64(81_000_000_000_000), record.BurnedQUs())
}

func Test_NewRecord_givenNoCompletedEpoch_thenError(t *testing.T) {

	for _, spectrumEpoch := range []uint32{0, 1} {
		_, err := NewRecord(spectrumEpoch, 1, 1)
		assert.Error(t, err, "spectrum epoch %d", spectrumEpoch)
	}
}
