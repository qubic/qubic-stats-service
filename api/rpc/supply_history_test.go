package rpc

import (
	"context"
	"testing"

	"github.com/qubic/qubic-stats-api/cache"
	"github.com/qubic/qubic-stats-api/protobuff"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// supplyHistoryServer builds a server over a cache holding the given epochs, where epoch N has a
// supply of N * 100 and closed at N * 1000.
func supplyHistoryServer(epochs ...uint32) *Server {

	history := make(cache.SupplyHistory, 0, len(epochs))
	for _, epoch := range epochs {
		history = append(history, cache.EpochStats{
			Epoch:             epoch,
			CirculatingSupply: int64(epoch) * 100,
			TotalEmitted:      int64(epoch) * 1_000_000_000_000,
			EpochEndTimestamp: int64(epoch) * 1000,
			SupplySource:      "spectrum",
		})
	}

	dataCache := &cache.Cache{}
	dataCache.UpdateSupplyHistory(history)
	dataCache.UpdateDataCache(
		cache.SpectrumData{Timestamp: 1, CirculatingSupply: 4200, ActiveAddresses: 7},
		cache.QubicData{Timestamp: 1, Epoch: 42},
	)

	return &Server{cache: dataCache}
}

func epochsOf(points []*protobuff.SupplyHistoryPoint) []uint32 {
	epochs := make([]uint32, 0, len(points))
	for _, point := range points {
		epochs = append(epochs, point.GetEpoch())
	}
	return epochs
}

func Test_GetSupplyHistory_givenNoRange_thenWholeHistoryAscending(t *testing.T) {

	server := supplyHistoryServer(40, 41, 42)

	response, err := server.GetSupplyHistory(context.Background(), &protobuff.GetSupplyHistoryRequest{})
	require.NoError(t, err)

	assert.Equal(t, []uint32{40, 41, 42}, epochsOf(response.GetPoints()))
	assert.Equal(t, int64(1_000_000_000_000_000), response.GetSupplyCap())
	assert.Equal(t, uint32(42), response.GetCurrentEpoch())
	assert.Equal(t, int64(4200), response.GetCurrentCirculatingSupply())

	assert.Equal(t, &protobuff.SupplyHistoryPoint{
		Epoch:             41,
		CirculatingSupply: 4100,
		TotalEmitted:      41_000_000_000_000,
		Timestamp:         41000,
		SupplySource:      "spectrum",
	}, response.GetPoints()[1])
}

func Test_GetSupplyHistory_givenRange_thenBoundsAreInclusive(t *testing.T) {

	server := supplyHistoryServer(38, 39, 40, 41, 42)

	tests := []struct {
		name      string
		fromEpoch uint32
		toEpoch   uint32
		expected  []uint32
	}{
		{name: "both bounds", fromEpoch: 39, toEpoch: 41, expected: []uint32{39, 40, 41}},
		{name: "from only", fromEpoch: 41, expected: []uint32{41, 42}},
		{name: "to only", toEpoch: 39, expected: []uint32{38, 39}},
		{name: "single epoch", fromEpoch: 40, toEpoch: 40, expected: []uint32{40}},
		{name: "beyond the history", fromEpoch: 100, toEpoch: 200, expected: []uint32{}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response, err := server.GetSupplyHistory(context.Background(), &protobuff.GetSupplyHistoryRequest{
				FromEpoch: test.fromEpoch,
				ToEpoch:   test.toEpoch,
			})
			require.NoError(t, err)
			assert.Equal(t, test.expected, epochsOf(response.GetPoints()))
		})
	}
}

func Test_GetSupplyHistory_givenLimit_thenMostRecentPointsKept(t *testing.T) {

	server := supplyHistoryServer(38, 39, 40, 41, 42)

	response, err := server.GetSupplyHistory(context.Background(), &protobuff.GetSupplyHistoryRequest{Limit: 2})
	require.NoError(t, err)

	assert.Equal(t, []uint32{41, 42}, epochsOf(response.GetPoints()))
}

func Test_GetSupplyHistory_givenLimitLargerThanRange_thenEverythingKept(t *testing.T) {

	server := supplyHistoryServer(41, 42)

	response, err := server.GetSupplyHistory(context.Background(), &protobuff.GetSupplyHistoryRequest{Limit: 500})
	require.NoError(t, err)

	assert.Equal(t, []uint32{41, 42}, epochsOf(response.GetPoints()))
}

func Test_GetSupplyHistory_givenEmptyHistory_thenNoPointsAndNoError(t *testing.T) {

	server := supplyHistoryServer()

	response, err := server.GetSupplyHistory(context.Background(), &protobuff.GetSupplyHistoryRequest{})
	require.NoError(t, err)

	assert.Empty(t, response.GetPoints())
	assert.Equal(t, uint32(42), response.GetCurrentEpoch()) // the rest of the response still stands
}

func Test_GetSupplyHistory_givenInvalidArguments_thenError(t *testing.T) {

	server := supplyHistoryServer(41, 42)

	tests := []struct {
		name    string
		request *protobuff.GetSupplyHistoryRequest
	}{
		{name: "inverted range", request: &protobuff.GetSupplyHistoryRequest{FromEpoch: 42, ToEpoch: 41}},
		{name: "negative limit", request: &protobuff.GetSupplyHistoryRequest{Limit: -1}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := server.GetSupplyHistory(context.Background(), test.request)
			require.Error(t, err)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		})
	}
}

// The ticket requires that the supply history cannot drift away from the latest stats endpoint.
func Test_GetSupplyHistory_thenCurrentSupplyMatchesLatestStats(t *testing.T) {

	server := supplyHistoryServer(41, 42)

	history, err := server.GetSupplyHistory(context.Background(), &protobuff.GetSupplyHistoryRequest{})
	require.NoError(t, err)

	latest, err := server.GetLatestData(context.Background(), nil)
	require.NoError(t, err)

	assert.Equal(t, latest.GetData().GetCirculatingSupply(), history.GetCurrentCirculatingSupply())
	assert.Equal(t, latest.GetData().GetEpoch(), history.GetCurrentEpoch())
}
