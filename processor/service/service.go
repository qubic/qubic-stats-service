package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	queryProto "github.com/qubic/archive-query-service/legacy/protobuf"
	liveProto "github.com/qubic/qubic-http/protobuff"
	"github.com/qubic/qubic-stats-processor/spectrum"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	// tickQualityWindowSize is the number of most recent ticks the rolling tick quality is based on.
	tickQualityWindowSize = 10000
	// tickListPageSize is the maximum page size accepted by the query service tick list endpoint.
	tickListPageSize = 1000
)

type Service struct {
	CoinGeckoToken          string
	QueryServiceGrpcAddress string
	LiveServiceGrpcAddress  string

	MongoClient              *mongo.Client
	MongoDatabase            string
	MongoSpectrumCollection  string
	MongoQubicDataCollection string

	ScrapeInterval time.Duration
	ScrapeTimeout  time.Duration

	spectrumData *spectrum.Data // we keep this here for caching purposes
}

type Data struct {
	Timestamp                int64
	Price                    float32
	MarketCap                int64
	Epoch                    uint32
	CurrentTick              uint32
	TicksInCurrentEpoch      uint32
	EmptyTicksInCurrentEpoch uint32
	EpochTickQuality         float32
	BurnedQUs                uint64
	TicksInLast10000         uint32
	EmptyTicksInLast10000    uint32
	Last10000TickQuality     float32
}

func (s *Service) RunService() error {

	println("Starting processor service...")

	ticker := time.NewTicker(time.Second)

	for range ticker.C {
		ticker.Reset(s.ScrapeInterval)
		println("Scraping for data... ")

		data, err := s.scrapeData()
		if err != nil {
			log.Printf("Failed to fetch data. Error: %v", err)
			continue
		}

		println("Done scraping data.")

		println("Compiled data:")
		fmt.Printf("    Price: %.9f\n", data.Price)
		fmt.Printf("    Market Cap: %d\n", data.MarketCap)
		fmt.Printf("    Epoch: %d\n", data.Epoch)
		fmt.Printf("    Current Tick: %d\n", data.CurrentTick)
		fmt.Printf("    Ticks this Epoch: %d\n", data.TicksInCurrentEpoch)
		fmt.Printf("    Empty Ticks this Epoch: %d\n", data.EmptyTicksInCurrentEpoch)
		fmt.Printf("    Tick Quality: %f\n", data.EpochTickQuality)
		fmt.Printf("    Burned QUs: %d\n", data.BurnedQUs)
		fmt.Printf("    Ticks in last %d: %d\n", tickQualityWindowSize, data.TicksInLast10000)
		fmt.Printf("    Empty Ticks in last %d: %d\n", tickQualityWindowSize, data.EmptyTicksInLast10000)
		fmt.Printf("    Last %d Tick Quality: %f\n", tickQualityWindowSize, data.Last10000TickQuality)

		println("Saving data to database...")
		err = s.saveData(data)
		if err != nil {
			log.Printf("Failed to save the data. Error: %v", err)
		}
		println("Done saving.")
	}

	return nil
}

func (s *Service) scrapeData() (Data, error) {

	ctx, cancel := context.WithTimeout(context.Background(), s.ScrapeTimeout)
	defer cancel()

	queryServiceConnection, err := grpc.NewClient(s.QueryServiceGrpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return Data{}, fmt.Errorf("creating query service grpc connection: %w", err)
	}
	defer func(connection *grpc.ClientConn) {
		err := connection.Close()
		if err != nil {
			fmt.Printf("failed to close query service grpc connection")
		}
	}(queryServiceConnection)
	queryServiceClient := queryProto.NewTransactionsServiceClient(queryServiceConnection)

	liveServiceConnection, err := grpc.NewClient(s.LiveServiceGrpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return Data{}, fmt.Errorf("creating live serive grpc connection")
	}
	defer func(connection *grpc.ClientConn) {
		err := connection.Close()
		if err != nil {
			fmt.Printf("failed to close live service grpc connection")
		}
	}(liveServiceConnection)
	liveServiceClient := liveProto.NewQubicLiveServiceClient(liveServiceConnection)

	price, err := FetchCoinGeckoPrice(ctx, s.CoinGeckoToken)
	if err != nil {
		return Data{}, fmt.Errorf("fetching qubic price from coingecko: %w", err)
	}

	spectrumData, err := spectrum.LoadSpectrumDataFromDatabase(ctx, s.MongoClient, s.MongoDatabase, s.MongoSpectrumCollection)
	if err != nil {
		if s.spectrumData == nil {
			return Data{}, fmt.Errorf("fetching initial spectrum data from database: %w", err)
		}
		fmt.Printf("Failed to update spectrum data: %v", err)
	}

	marketCap := int64(float64(price) * float64(spectrumData.CirculatingSupply))

	archiverStatus, err := fetchQueryServiceStatus(ctx, queryServiceClient)
	if err != nil {
		return Data{}, fmt.Errorf("fetching query service archiver status: %w", err)
	}

	epoch := archiverStatus.LastProcessedTick.Epoch

	latestTick, err := fetchLiveServiceNetworkTick(ctx, liveServiceClient)
	if err != nil {
		return Data{}, fmt.Errorf("fetching live service latest tick: %w", err)
	}

	ticksThisEpoch, err := fetchQueryServiceEpochTotalTickCount(ctx, queryServiceClient, epoch)
	if err != nil {
		return Data{}, fmt.Errorf("fetching query service tick count for epoch %d: %w", epoch, err)
	}

	burnedQUs := (uint64(epoch) * uint64(1000000000000)) - uint64(spectrumData.CirculatingSupply)

	emptyTickCount, err := fetchQueryServiceEpochEmptyTickCount(ctx, queryServiceClient, epoch)
	if err != nil {
		return Data{}, fmt.Errorf("fetching query service empty tick count for epoch %d: %w", epoch, err)
	}

	qualityWindow, err := fetchTickQualityWindow(ctx, queryServiceClient, epoch, ticksThisEpoch)
	if err != nil {
		return Data{}, fmt.Errorf("fetching tick quality window for epoch %d: %w", epoch, err)
	}

	serviceData := Data{
		Timestamp:                time.Now().Unix(),
		Price:                    price,
		MarketCap:                marketCap,
		Epoch:                    epoch,
		CurrentTick:              latestTick,
		TicksInCurrentEpoch:      uint32(ticksThisEpoch), // the code originally uses uint32 here
		EmptyTicksInCurrentEpoch: uint32(emptyTickCount), // I am not currently sure of the implications related to changing the data type to normal int
		EpochTickQuality:         calculateTickQuality(uint32(ticksThisEpoch), uint32(emptyTickCount)),
		BurnedQUs:                burnedQUs,
		TicksInLast10000:         qualityWindow.TickCount,
		EmptyTicksInLast10000:    qualityWindow.EmptyTickCount,
		Last10000TickQuality:     calculateTickQuality(qualityWindow.TickCount, qualityWindow.EmptyTickCount),
	}

	return serviceData, nil
}

func (s *Service) saveData(data Data) error {

	collection := s.MongoClient.Database(s.MongoDatabase).Collection(s.MongoQubicDataCollection)
	_, err := collection.InsertOne(context.Background(), data)
	if err != nil {
		return fmt.Errorf("inserting data in collection: %w", err)
	}

	return nil
}

func fetchQueryServiceStatus(ctx context.Context, client queryProto.TransactionsServiceClient) (*queryProto.GetArchiverStatusResponse, error) {
	status, err := client.GetArchiverStatus(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("getting query service processed tick intervals: %w", err)
	}

	return status, nil
}

func fetchQueryServiceEpochTotalTickCount(ctx context.Context, client queryProto.TransactionsServiceClient, epoch uint32) (int32, error) {
	tickListResponse, err := client.GetEpochTickListV2(ctx, &queryProto.GetEpochTickListRequestV2{
		Epoch: epoch,
	})
	if err != nil {
		return 0, fmt.Errorf("fetching query service tick list")
	}

	return tickListResponse.Pagination.TotalRecords, nil
}

func fetchQueryServiceEpochEmptyTickCount(ctx context.Context, client queryProto.TransactionsServiceClient, epoch uint32) (int32, error) {
	emptyTickListResponse, err := client.GetEmptyTickListV2(ctx, &queryProto.GetEpochEmptyTickListRequestV2{
		Epoch: epoch,
	})
	if err != nil {
		return 0, fmt.Errorf("fetching query service empty tick list")
	}

	return emptyTickListResponse.Pagination.TotalRecords, nil
}

// epochTickListClient is the subset of queryProto.TransactionsServiceClient that is needed to page
// through the tick list of an epoch.
type epochTickListClient interface {
	GetEpochTickListV2(ctx context.Context, in *queryProto.GetEpochTickListRequestV2, opts ...grpc.CallOption) (*queryProto.GetEpochTickListResponseV2, error)
}

// tickQualityWindow holds the tick counts of the most recent ticks of an epoch.
type tickQualityWindow struct {
	TickCount      uint32
	EmptyTickCount uint32
}

// fetchTickQualityWindow counts the ticks and the empty ticks among the last tickQualityWindowSize
// ticks of the given epoch.
//
// The query service offers no tick range query, so the window is assembled from descending pages of
// the epoch tick list. Only the current epoch is taken into account, which means that the window is
// shorter than tickQualityWindowSize right after an epoch change. The returned tick count states how
// many ticks the window really holds.
func fetchTickQualityWindow(ctx context.Context, client epochTickListClient, epoch uint32, epochTickCount int32) (tickQualityWindow, error) {

	if epochTickCount <= 0 {
		return tickQualityWindow{}, nil
	}

	windowSize := min(tickQualityWindowSize, epochTickCount)
	pageCount := int((windowSize + tickListPageSize - 1) / tickListPageSize)

	pages := make([][]*queryProto.Tick, pageCount)

	errorGroup, groupContext := errgroup.WithContext(ctx)
	for index := range pages {
		page := int32(index) + 1 // page numbering starts at 1
		errorGroup.Go(func() error {
			response, err := client.GetEpochTickListV2(groupContext, &queryProto.GetEpochTickListRequestV2{
				Epoch:    epoch,
				Page:     page,
				PageSize: tickListPageSize,
				Desc:     true, // the first page holds the most recent ticks
			})
			if err != nil {
				return fmt.Errorf("fetching tick list page %d for epoch %d: %w", page, epoch, err)
			}
			pages[index] = response.GetTicks()
			return nil
		})
	}
	if err := errorGroup.Wait(); err != nil {
		return tickQualityWindow{}, err
	}

	// The epoch keeps growing while the pages are fetched, which shifts the page boundaries and can
	// return the same tick on two pages. Collecting the ticks in a set keeps the counts exact.
	ticks := make(map[uint32]bool, windowSize)
	for _, page := range pages {
		for _, tick := range page {
			ticks[tick.GetTickNumber()] = tick.GetIsEmpty()
		}
	}

	var window tickQualityWindow
	for _, isEmpty := range ticks {
		window.TickCount++
		if isEmpty {
			window.EmptyTickCount++
		}
	}

	return window, nil
}

// calculateTickQuality returns the percentage of non-empty ticks.
func calculateTickQuality(tickCount, emptyTickCount uint32) float32 {
	if tickCount == 0 {
		return 0
	}
	return (float32(tickCount-emptyTickCount) / float32(tickCount)) * 100
}

func fetchLiveServiceNetworkTick(ctx context.Context, client liveProto.QubicLiveServiceClient) (uint32, error) {
	tickInfo, err := client.GetTickInfo(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("getting live serivice tick info: %w", err)
	}
	return tickInfo.TickInfo.Tick, nil
}

type coinGeckoResponse struct {
	QubicNetwork struct {
		Usd float32 `json:"usd"`
	} `json:"qubic-network"`
}

func FetchCoinGeckoPrice(ctx context.Context, token string) (float32, error) {

	url := "https://api.coingecko.com/api/v3/simple/price?ids=qubic-network&vs_currencies=usd&precision=9"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("x-cg-demo-api-key", token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("executing request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, fmt.Errorf("reading request response: %w", err)
	}

	var response coinGeckoResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0, fmt.Errorf("unmarshalling response: %w", err)
	}

	return response.QubicNetwork.Usd, nil
}
