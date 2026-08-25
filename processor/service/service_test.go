package service

import (
	"context"
	"errors"
	"sync"
	"testing"

	queryProto "github.com/qubic/archive-query-service/legacy/protobuf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// fakeTickListClient mimics the paging behaviour of the query service tick list endpoint. Ticks are
// numbered firstTick to firstTick+tickCount-1 and pages are served in descending order.
type fakeTickListClient struct {
	firstTick  uint32
	tickCount  int32
	emptyTicks map[uint32]bool
	failOnPage int32

	mutex    sync.Mutex
	requests []*queryProto.GetEpochTickListRequestV2
}

func (f *fakeTickListClient) GetEpochTickListV2(_ context.Context, in *queryProto.GetEpochTickListRequestV2, _ ...grpc.CallOption) (*queryProto.GetEpochTickListResponseV2, error) {

	f.mutex.Lock()
	f.requests = append(f.requests, in)
	f.mutex.Unlock()

	if f.failOnPage != 0 && f.failOnPage == in.GetPage() {
		return nil, errors.New("page unavailable")
	}

	start := (in.GetPage() - 1) * in.GetPageSize()
	end := min(start+in.GetPageSize(), f.tickCount)

	ticks := make([]*queryProto.Tick, 0, max(0, end-start))
	for index := start; index < end; index++ {
		tickNumber := f.firstTick + uint32(f.tickCount-1-index) // the first page holds the newest ticks
		ticks = append(ticks, &queryProto.Tick{
			TickNumber: tickNumber,
			IsEmpty:    f.emptyTicks[tickNumber],
		})
	}

	return &queryProto.GetEpochTickListResponseV2{Ticks: ticks}, nil
}

func (f *fakeTickListClient) requestedPages() []int32 {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	pages := make([]int32, 0, len(f.requests))
	for _, request := range f.requests {
		pages = append(pages, request.GetPage())
	}
	return pages
}

// everyHundredthTick marks every tick number that is a multiple of 100 as empty.
func everyHundredthTick(firstTick uint32, tickCount int32) map[uint32]bool {
	empty := make(map[uint32]bool)
	for tickNumber := firstTick; tickNumber < firstTick+uint32(tickCount); tickNumber++ {
		if tickNumber%100 == 0 {
			empty[tickNumber] = true
		}
	}
	return empty
}

func Test_FetchTickQualityWindow_givenMoreTicksThanWindow_thenWindowIsCapped(t *testing.T) {

	const firstTick = uint32(1000)
	const tickCount = int32(25000)

	client := &fakeTickListClient{
		firstTick:  firstTick,
		tickCount:  tickCount,
		emptyTicks: everyHundredthTick(firstTick, tickCount),
	}

	window, err := fetchTickQualityWindow(context.Background(), client, 42, tickCount)
	require.NoError(t, err)

	// the window covers ticks 16000 to 25999, which holds 100 multiples of 100
	assert.Equal(t, uint32(10000), window.TickCount)
	assert.Equal(t, uint32(100), window.EmptyTickCount)
	assert.InDelta(t, 99.0, calculateTickQuality(window.TickCount, window.EmptyTickCount), 0.0001)

	assert.ElementsMatch(t, []int32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, client.requestedPages())
	for _, request := range client.requests {
		assert.Equal(t, uint32(42), request.GetEpoch())
		assert.Equal(t, int32(tickListPageSize), request.GetPageSize())
		assert.True(t, request.GetDesc())
	}
}

func Test_FetchTickQualityWindow_givenFewerTicksThanWindow_thenWindowIsPartial(t *testing.T) {

	const firstTick = uint32(0)
	const tickCount = int32(3500)

	client := &fakeTickListClient{
		firstTick:  firstTick,
		tickCount:  tickCount,
		emptyTicks: everyHundredthTick(firstTick, tickCount),
	}

	window, err := fetchTickQualityWindow(context.Background(), client, 42, tickCount)
	require.NoError(t, err)

	// a fresh epoch only offers the ticks it already has: 0 to 3499, holding 35 multiples of 100
	assert.Equal(t, uint32(3500), window.TickCount)
	assert.Equal(t, uint32(35), window.EmptyTickCount)
	assert.InDelta(t, 99.0, calculateTickQuality(window.TickCount, window.EmptyTickCount), 0.0001)

	assert.ElementsMatch(t, []int32{1, 2, 3, 4}, client.requestedPages())
}

func Test_FetchTickQualityWindow_givenPartialLastPage_thenAllTicksCounted(t *testing.T) {

	const tickCount = int32(1234)

	client := &fakeTickListClient{
		tickCount:  tickCount,
		emptyTicks: map[uint32]bool{},
	}

	window, err := fetchTickQualityWindow(context.Background(), client, 42, tickCount)
	require.NoError(t, err)

	assert.Equal(t, uint32(1234), window.TickCount)
	assert.Equal(t, uint32(0), window.EmptyTickCount)
	assert.ElementsMatch(t, []int32{1, 2}, client.requestedPages())
}

func Test_FetchTickQualityWindow_givenNoTicks_thenEmptyWindow(t *testing.T) {

	for _, tickCount := range []int32{0, -1} {
		client := &fakeTickListClient{tickCount: tickCount}

		window, err := fetchTickQualityWindow(context.Background(), client, 42, tickCount)
		require.NoError(t, err)

		assert.Equal(t, tickQualityWindow{}, window)
		assert.Empty(t, client.requestedPages()) // nothing to query
		assert.Zero(t, calculateTickQuality(window.TickCount, window.EmptyTickCount))
	}
}

// overlappingTickListClient returns pages that share ticks, which happens when the epoch grows while
// the pages are fetched and the page boundaries shift.
type overlappingTickListClient struct {
	pages map[int32][]*queryProto.Tick
}

func (o *overlappingTickListClient) GetEpochTickListV2(_ context.Context, in *queryProto.GetEpochTickListRequestV2, _ ...grpc.CallOption) (*queryProto.GetEpochTickListResponseV2, error) {
	return &queryProto.GetEpochTickListResponseV2{Ticks: o.pages[in.GetPage()]}, nil
}

func Test_FetchTickQualityWindow_givenOverlappingPages_thenTicksCountedOnce(t *testing.T) {

	buildPage := func(from, to uint32, emptyTick uint32) []*queryProto.Tick {
		var ticks []*queryProto.Tick
		for tickNumber := from; tickNumber <= to; tickNumber++ {
			ticks = append(ticks, &queryProto.Tick{
				TickNumber: tickNumber,
				IsEmpty:    tickNumber == emptyTick,
			})
		}
		return ticks
	}

	client := &overlappingTickListClient{
		pages: map[int32][]*queryProto.Tick{
			1: buildPage(1001, 2000, 1003),
			2: buildPage(6, 1005, 1003), // ticks 1001 to 1005 are served twice
		},
	}

	window, err := fetchTickQualityWindow(context.Background(), client, 42, 2000)
	require.NoError(t, err)

	assert.Equal(t, uint32(1995), window.TickCount)   // 2000 ticks minus the 5 duplicates
	assert.Equal(t, uint32(1), window.EmptyTickCount) // the duplicated empty tick is counted once
}

func Test_FetchTickQualityWindow_givenFailingPage_thenError(t *testing.T) {

	client := &fakeTickListClient{
		tickCount:  25000,
		emptyTicks: map[uint32]bool{},
		failOnPage: 3,
	}

	window, err := fetchTickQualityWindow(context.Background(), client, 42, 25000)

	require.Error(t, err)
	assert.ErrorContains(t, err, "fetching tick list page 3 for epoch 42")
	assert.Equal(t, tickQualityWindow{}, window) // no partial result
}

func Test_CalculateTickQuality(t *testing.T) {

	tests := []struct {
		name           string
		tickCount      uint32
		emptyTickCount uint32
		expected       float32
	}{
		{name: "no ticks", tickCount: 0, emptyTickCount: 0, expected: 0}, // must not produce NaN
		{name: "no empty ticks", tickCount: 100, emptyTickCount: 0, expected: 100},
		{name: "all empty ticks", tickCount: 100, emptyTickCount: 100, expected: 0},
		{name: "issue example", tickCount: 10000, emptyTickCount: 41, expected: 99.59},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.InDelta(t, test.expected, calculateTickQuality(test.tickCount, test.emptyTickCount), 0.0001)
		})
	}
}
