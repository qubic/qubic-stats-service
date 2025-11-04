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
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	goodTicks := ticksThisEpoch - emptyTickCount

	var tickQuality = (float32(goodTicks) / float32(ticksThisEpoch)) * 100

	serviceData := Data{
		Timestamp:                time.Now().Unix(),
		Price:                    price,
		MarketCap:                marketCap,
		Epoch:                    epoch,
		CurrentTick:              latestTick,
		TicksInCurrentEpoch:      uint32(ticksThisEpoch), // the code originally uses uint32 here
		EmptyTicksInCurrentEpoch: uint32(emptyTickCount), // I am not currently sure of the implications related to changing the data type to normal int
		EpochTickQuality:         tickQuality,
		BurnedQUs:                burnedQUs,
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
