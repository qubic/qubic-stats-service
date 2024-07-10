package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/qubic/frontend-service-processor/src/archiver"
	"github.com/qubic/frontend-service-processor/src/db"
	"github.com/qubic/frontend-service-processor/src/spectrum"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"log"
	"net/http"
	"time"
)

type Service struct {
	CoinGeckoToken         string
	ArchiverUrl            string
	ArchiverStatusPath     string
	ArchiverLatestTickPath string

	DatabaseConnection db.Connection
	Database           string
	SpectrumCollection string
	DataCollection     string

	ScrapeInterval time.Duration

	dbClient     *mongo.Client
	spectrumData *spectrum.Data // we keep this here for caching purposes
}

type Data struct {
	Timestamp                int64   `json:"timestamp"`
	Price                    float32 `json:"qubic_price"`
	MarketCap                int64   `json:"market_cap"`
	Epoch                    uint32  `json:"epoch"`
	CurrentTick              uint32  `json:"current_tick"`
	TicksInCurrentEpoch      uint32  `json:"ticks_in_current_epoch"`
	EmptyTicksInCurrentEpoch uint32  `json:"empty_ticks_in_current_epoch"`
	EpochTickQuality         float32 `json:"epoch_tick_quality"`
}

func (s *Service) RunService() error {

	println("Starting web service...")

	err := s.createDatabaseClient()
	if err != nil {
		return errors.Wrap(err, "creating database client")
	}
	defer func() {
		if err = s.dbClient.Disconnect(context.Background()); err != nil {
			log.Fatalf("service: exited with error: %s\n", err.Error())
		}
	}()

	println("Starting scrape loop...")

	ticker := time.NewTicker(s.ScrapeInterval)

	for range ticker.C {
		println("Scraping for data... ")

		data, err := s.scrapeData()
		if err != nil {
			log.Printf("Failed to fetch data. Error: %v", err)
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

	spectrumData, err := spectrum.LoadSpectrumDataFromDatabase(s.dbClient, s.Database, s.SpectrumCollection)
	if err != nil {
		if s.spectrumData == nil {
			return nil, errors.Wrap(err, "fetching initial spectrum data from database")
		}
		fmt.Printf("Failed to update spectrum data: %v", err)
	}

	marketCap := int64(float64(price) * float64(spectrumData.CirculatingSupply))

	archiverStatus, err := fetchArchiverStatus(s.ArchiverUrl + s.ArchiverStatusPath)
	if err != nil {
		return nil, errors.Wrap(err, "fetching archiver status")
	}

	epoch := archiverStatus.LastProcessedTick.Epoch

	latestTick, err := fetchLatestTick(s.ArchiverUrl + s.ArchiverLatestTickPath)
	if err != nil {
		return nil, errors.Wrap(err, "fetching latest tick")
	}

	epochStartingTick, err := getEpochStartingTick(archiverStatus, epoch)
	if err != nil {
		return nil, errors.Wrap(err, "getting starting tick for current interval")
	}

	ticksThisEpoch := latestTick - epochStartingTick

	serviceData := Data{
		Price:                    price,
		MarketCap:                marketCap,
		Epoch:                    epoch,
		CurrentTick:              latestTick,
		TicksInCurrentEpoch:      ticksThisEpoch,
		EmptyTicksInCurrentEpoch: 0,
		EpochTickQuality:         0,
	}

	return &serviceData, nil
}

func (s *Service) saveData(data *Data) error {

	collection := s.dbClient.Database(s.Database).Collection(s.DataCollection)
	_, err := collection.InsertOne(context.Background(), data)
	if err != nil {
		return errors.Wrap(err, "inserting data in collection")
	}

	return nil
}

func fetchArchiverStatus(statusURL string) (*archiver.StatusResponse, error) {

	req, err := http.NewRequest("GET", statusURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "creating request")
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "executing request")
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "reading request response")
	}

	var archiverStatus archiver.StatusResponse

	err = json.Unmarshal(body, &archiverStatus)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshalling archiver status")
	}

	return &archiverStatus, err
}

func fetchLatestTick(latestTickURL string) (uint32, error) {
	req, err := http.NewRequest("GET", latestTickURL, nil)
	if err != nil {
		return 0, errors.Wrap(err, "creating request")
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, errors.Wrap(err, "executing request")
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, errors.Wrap(err, "reading request response")
	}

	var archiverStatus archiver.LatestTickResponse

	err = json.Unmarshal(body, &archiverStatus)
	if err != nil {
		return 0, errors.Wrap(err, "unmarshalling archiver status")
	}

	return archiverStatus.LatestTick, err
}

func getEpochStartingTick(archiverStatus *archiver.StatusResponse, epoch uint32) (uint32, error) {
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

func (s *Service) createDatabaseClient() error {
	println("Connecting to database...")
	client, err := db.CreateClient(&s.DatabaseConnection)
	if err != nil {
		return errors.Wrap(err, "connecting to database")
	}
	s.dbClient = client
	return nil
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
