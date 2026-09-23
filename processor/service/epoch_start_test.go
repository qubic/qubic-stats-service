package service

import (
	"context"
	"errors"
	"testing"

	queryProto "github.com/qubic/archive-query-service/legacy/protobuf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// fakeEpochStartClient serves the ticks of an epoch in ascending order, which is how the epoch start
// lookup asks for them.
type fakeEpochStartClient struct {
	firstTick  uint32
	tickCount  int32
	emptyTicks map[uint32]bool
	// timestamps holds the tick data timestamps in milliseconds. A tick that is missing here has no
	// tick data.
	timestamps   map[uint32]uint64
	tickDataErrs map[uint32]bool
	listErr      error

	listRequests []*queryProto.GetEpochTickListRequestV2
	tickRequests []uint32
}

func (f *fakeEpochStartClient) GetEpochTickListV2(_ context.Context, in *queryProto.GetEpochTickListRequestV2, _ ...grpc.CallOption) (*queryProto.GetEpochTickListResponseV2, error) {

	f.listRequests = append(f.listRequests, in)

	if f.listErr != nil {
		return nil, f.listErr
	}

	start := (in.GetPage() - 1) * in.GetPageSize()
	end := min(start+in.GetPageSize(), f.tickCount)

	ticks := make([]*queryProto.Tick, 0, max(0, end-start))
	for index := start; index < end; index++ {
		tickNumber := f.firstTick + uint32(index) // the first page holds the oldest ticks
		ticks = append(ticks, &queryProto.Tick{
			TickNumber: tickNumber,
			IsEmpty:    f.emptyTicks[tickNumber],
		})
	}

	return &queryProto.GetEpochTickListResponseV2{Ticks: ticks}, nil
}

func (f *fakeEpochStartClient) GetTickData(_ context.Context, in *queryProto.GetTickDataRequest, _ ...grpc.CallOption) (*queryProto.GetTickDataResponse, error) {

	f.tickRequests = append(f.tickRequests, in.GetTickNumber())

	if f.tickDataErrs[in.GetTickNumber()] {
		return nil, errors.New("tick data unavailable")
	}

	timestamp, found := f.timestamps[in.GetTickNumber()]
	if !found {
		return &queryProto.GetTickDataResponse{}, nil
	}

	return &queryProto.GetTickDataResponse{
		TickData: &queryProto.TickData{TickNumber: in.GetTickNumber(), Timestamp: timestamp},
	}, nil
}

// emptyUntil marks every tick before the given one as empty.
func emptyUntil(firstTick, until uint32) map[uint32]bool {
	empty := make(map[uint32]bool)
	for tickNumber := firstTick; tickNumber < until; tickNumber++ {
		empty[tickNumber] = true
	}
	return empty
}

func Test_FetchEpochStartTimestamp_givenFirstTickHasData_thenTimestampInSeconds(t *testing.T) {

	client := &fakeEpochStartClient{
		firstTick:  1000,
		tickCount:  500,
		emptyTicks: map[uint32]bool{},
		// the query service reports tick timestamps in milliseconds
		timestamps: map[uint32]uint64{1000: 1753862400123},
	}

	timestamp, err := fetchEpochStartTimestamp(context.Background(), client, 42)
	require.NoError(t, err)

	assert.Equal(t, int64(1753862400), timestamp)
	assert.Equal(t, []uint32{1000}, client.tickRequests)

	require.Len(t, client.listRequests, 1)
	assert.Equal(t, uint32(42), client.listRequests[0].GetEpoch())
	assert.Equal(t, int32(1), client.listRequests[0].GetPage())
	assert.Equal(t, int32(epochStartProbePageSize), client.listRequests[0].GetPageSize())
	assert.False(t, client.listRequests[0].GetDesc()) // oldest ticks first
}

func Test_FetchEpochStartTimestamp_givenEmptyOpeningTicks_thenFirstTickWithDataUsed(t *testing.T) {

	client := &fakeEpochStartClient{
		firstTick:  1000,
		tickCount:  500,
		emptyTicks: emptyUntil(1000, 1005),
		timestamps: map[uint32]uint64{1005: 2_000_000},
	}

	timestamp, err := fetchEpochStartTimestamp(context.Background(), client, 42)
	require.NoError(t, err)

	assert.Equal(t, int64(2000), timestamp)
	assert.Equal(t, []uint32{1005}, client.tickRequests) // the empty ticks are never asked about
}

func Test_FetchEpochStartTimestamp_givenWholeFirstPageEmpty_thenNextPageProbed(t *testing.T) {

	const firstTick = uint32(1000)

	client := &fakeEpochStartClient{
		firstTick:  firstTick,
		tickCount:  500,
		emptyTicks: emptyUntil(firstTick, firstTick+epochStartProbePageSize+3),
		timestamps: map[uint32]uint64{firstTick + epochStartProbePageSize + 3: 3_000_000},
	}

	timestamp, err := fetchEpochStartTimestamp(context.Background(), client, 42)
	require.NoError(t, err)

	assert.Equal(t, int64(3000), timestamp)
	assert.Len(t, client.listRequests, 2)
}

func Test_FetchEpochStartTimestamp_givenUnreadableTick_thenNextTickTried(t *testing.T) {

	client := &fakeEpochStartClient{
		firstTick:    1000,
		tickCount:    500,
		emptyTicks:   map[uint32]bool{},
		timestamps:   map[uint32]uint64{1001: 4_000_000},
		tickDataErrs: map[uint32]bool{1000: true},
	}

	timestamp, err := fetchEpochStartTimestamp(context.Background(), client, 42)
	require.NoError(t, err)

	assert.Equal(t, int64(4000), timestamp)
	assert.Equal(t, []uint32{1000, 1001}, client.tickRequests)
}

func Test_FetchEpochStartTimestamp_givenTickWithoutTimestamp_thenNextTickTried(t *testing.T) {

	client := &fakeEpochStartClient{
		firstTick:  1000,
		tickCount:  500,
		emptyTicks: map[uint32]bool{},
		timestamps: map[uint32]uint64{1002: 5_000_000}, // 1000 and 1001 answer without tick data
	}

	timestamp, err := fetchEpochStartTimestamp(context.Background(), client, 42)
	require.NoError(t, err)

	assert.Equal(t, int64(5000), timestamp)
}

func Test_FetchEpochStartTimestamp_givenNoTickWithData_thenError(t *testing.T) {

	client := &fakeEpochStartClient{
		firstTick:  1000,
		tickCount:  10,
		emptyTicks: emptyUntil(1000, 1010),
	}

	_, err := fetchEpochStartTimestamp(context.Background(), client, 42)

	require.Error(t, err)
	assert.ErrorContains(t, err, "no tick with data")
}

func Test_FetchEpochStartTimestamp_givenFailingTickList_thenError(t *testing.T) {

	client := &fakeEpochStartClient{listErr: errors.New("epoch unavailable")}

	_, err := fetchEpochStartTimestamp(context.Background(), client, 42)

	require.Error(t, err)
	assert.ErrorContains(t, err, "fetching tick list page 1 for epoch 42")
}

func Test_CacheEpochStartTimestamp_thenResolvedOnce(t *testing.T) {

	client := &fakeEpochStartClient{
		firstTick:  1000,
		tickCount:  500,
		emptyTicks: map[uint32]bool{},
		timestamps: map[uint32]uint64{1000: 6_000_000},
	}

	service := &Service{}

	service.cacheEpochStartTimestamp(context.Background(), client, 42)
	service.cacheEpochStartTimestamp(context.Background(), client, 42)

	assert.Equal(t, int64(6000), service.epochStartTimestamps[42])
	assert.Len(t, client.listRequests, 1) // the second call is served from the cache
}

func Test_CacheEpochStartTimestamp_givenFailure_thenNothingCachedAndRetried(t *testing.T) {

	client := &fakeEpochStartClient{listErr: errors.New("epoch unavailable")}

	service := &Service{}

	service.cacheEpochStartTimestamp(context.Background(), client, 42)
	assert.Empty(t, service.epochStartTimestamps)

	service.cacheEpochStartTimestamp(context.Background(), client, 42)
	assert.Len(t, client.listRequests, 2) // retried rather than remembered as failed
}
