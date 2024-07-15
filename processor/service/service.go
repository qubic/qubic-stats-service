package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/qubic/go-archiver/protobuff"
	"github.com/qubic/qubic-stats-processor/spectrum"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"log"
	"net/http"
	"time"
)

type Service struct {
	CoinGeckoToken      string
	ArchiverGrpcAddress string

	MongoClient              *mongo.Client
	MongoDatabase            string
	MongoSpectrumCollection  string
	MongoQubicDataCollection string

	ScrapeInterval time.Duration

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

func (s *Service) scrapeData() (*Data, error) {

	price, err := FetchCoinGeckoPrice(s.CoinGeckoToken)
	if err != nil {
		return nil, errors.Wrap(err, "fetching qubic price from coingecko")
	}

	spectrumData, err := spectrum.LoadSpectrumDataFromDatabase(s.MongoClient, s.MongoDatabase, s.MongoSpectrumCollection)
	if err != nil {
		if s.spectrumData == nil {
			return nil, errors.Wrap(err, "fetching initial spectrum data from database")
		}
		fmt.Printf("Failed to update spectrum data: %v", err)
	}

	marketCap := int64(float64(price) * float64(spectrumData.CirculatingSupply))

	archiverStatus, err := fetchArchiverStatus(s.ArchiverGrpcAddress)
	if err != nil {
		return nil, errors.Wrap(err, "fetching archiver status")
	}

	epoch := archiverStatus.LastProcessedTick.Epoch

	latestTick, err := fetchLatestTick(s.ArchiverGrpcAddress)
	if err != nil {
		return nil, errors.Wrap(err, "fetching latest tick")
	}

	epochStartingTick, err := getEpochStartingTick(archiverStatus, epoch)
	if err != nil {
		return nil, errors.Wrap(err, "getting starting tick for current interval")
	}

	ticksThisEpoch := latestTick - epochStartingTick

	burnedQUs := (uint64(epoch) * uint64(1000000000000)) - uint64(spectrumData.CirculatingSupply)

	serviceData := Data{
		Timestamp:                time.Now().Unix(),
		Price:                    price,
		MarketCap:                marketCap,
		Epoch:                    epoch,
		CurrentTick:              latestTick,
		TicksInCurrentEpoch:      ticksThisEpoch,
		EmptyTicksInCurrentEpoch: 0,
		EpochTickQuality:         0,
		BurnedQUs:                burnedQUs,
	}

	return &serviceData, nil
}

func (s *Service) saveData(data *Data) error {

	collection := s.MongoClient.Database(s.MongoDatabase).Collection(s.MongoQubicDataCollection)
	_, err := collection.InsertOne(context.Background(), data)
	if err != nil {
		return errors.Wrap(err, "inserting data in collection")
	}

	return nil
}

func fetchArchiverStatus(archiverGRPCAddress string) (*protobuff.GetStatusResponse, error) {

	connection, err := grpc.NewClient(archiverGRPCAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, errors.Wrap(err, "creating grpc connection")
	}
	client := protobuff.NewArchiveServiceClient(connection)
	status, err := client.GetStatus(context.Background(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "getting archiver status")
	}
	return status, nil

}

func fetchLatestTick(archiverGRPCAddress string) (uint32, error) {
	connection, err := grpc.NewClient(archiverGRPCAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return 0, errors.Wrap(err, "creating grpc connection")
	}
	client := protobuff.NewArchiveServiceClient(connection)
	latestTick, err := client.GetLatestTick(context.Background(), nil)
	if err != nil {
		return 0, errors.Wrap(err, "getting latest tick")
	}
	return latestTick.LatestTick, nil
}

func getEpochStartingTick(archiverStatus *protobuff.GetStatusResponse, epoch uint32) (uint32, error) {
	intervals := archiverStatus.ProcessedTickIntervalsPerEpoch

	// we start from the bottom because this function will usually wil be used for the latest epoch
	for i := len(intervals) - 1; i >= 0; i-- {
		interval := intervals[i]
		if interval.Epoch != epoch {
			continue
		}
		startingTick := interval.Intervals[0].InitialProcessedTick
		return startingTick, nil
	}

	return 0, errors.New("Could not find the specified epoch")

}

type coinGeckoResponse struct {
	QubicNetwork struct {
		Usd float32 `json:"usd"`
	} `json:"qubic-network"`
}

func FetchCoinGeckoPrice(token string) (float32, error) {

	url := "https://api.coingecko.com/api/v3/simple/price?ids=qubic-network&vs_currencies=usd&precision=9"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, errors.Wrap(err, "creating request")
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("x-cg-demo-api-key", token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, errors.Wrap(err, "executing request")
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, errors.Wrap(err, "reading request response")
	}

	var response coinGeckoResponse

	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0, errors.Wrap(err, "unmarshalling response")
	}

	return response.QubicNetwork.Usd, nil
}
