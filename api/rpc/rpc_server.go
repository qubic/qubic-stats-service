package rpc

import (
	"context"
	"log"
	"math"
	"net"
	"net/http"
	"regexp"
	"strconv"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"github.com/qubic/go-node-connector/types"
	"github.com/qubic/qubic-stats-api/cache"
	"github.com/qubic/qubic-stats-api/protobuff"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	protobuff.UnimplementedStatsServiceServer
	httpAddress string
	grpcAddress string

	cache *cache.Cache

	dbClient                *mongo.Client
	mongoDatabase           string
	mongoRichListCollection string

	richListPageSize int32
	richListLimit    int

	assetService AssetService
}

func (s *Server) GetLatestData(_ context.Context, _ *emptypb.Empty) (*protobuff.GetLatestDataResponse, error) {

	qubicData := s.cache.GetQubicData()
	spectrumData := s.cache.GetSpectrumData()

	return &protobuff.GetLatestDataResponse{
		Data: &protobuff.QubicData{
			Timestamp:                qubicData.Timestamp,
			Price:                    qubicData.Price,
			CirculatingSupply:        spectrumData.CirculatingSupply,
			ActiveAddresses:          int32(spectrumData.ActiveAddresses),
			MarketCap:                qubicData.MarketCap,
			Epoch:                    qubicData.Epoch,
			CurrentTick:              qubicData.CurrentTick,
			TicksInCurrentEpoch:      qubicData.TicksInCurrentEpoch,
			EmptyTicksInCurrentEpoch: qubicData.EmptyTicksInCurrentEpoch,
			EpochTickQuality:         qubicData.EpochTickQuality,
			BurnedQus:                qubicData.BurnedQUs,
			TicksInLast10000:         qubicData.TicksInLast10000,
			EmptyTicksInLast10000:    qubicData.EmptyTicksInLast10000,
			Last10000TickQuality:     qubicData.Last10000TickQuality,
		},
	}, nil

}

func (s *Server) GetRichListSlice(ctx context.Context, request *protobuff.GetRichListSliceRequest) (*protobuff.GetRichListSliceResponse, error) {

	var pageSize int
	if request.PageSize > s.richListPageSize {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid page size (maximum is %d).", s.richListPageSize)
	}
	if request.PageSize == 0 {
		pageSize = int(defaultPageSize)
	} else {
		pageSize = int(request.PageSize)
	}

	pageNumber := max(0, int(request.Page)-1) // API index starts with 1, implementation index starts with 0
	start := pageNumber * pageSize
	limit := pageSize

	if start+limit > s.richListLimit {
		limit -= (start + limit) - s.richListLimit // ensure that we are not requesting records than the limit for the last page
	}

	epoch := s.cache.GetQubicData().Epoch
	epochString := strconv.Itoa(int(epoch))

	dbRecordCount := s.cache.GetSpectrumData().ActiveAddresses
	totalRecords := min(dbRecordCount, s.richListLimit) // we do not want to expose the full rich list

	pagination, err := getPaginationInformation(totalRecords, pageNumber+1, pageSize)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "creating pagination info")
	}
	if pageNumber+1 > int(pagination.TotalPages) {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid page for current page size (maximum is %d)", pagination.TotalPages)
	}

	collection := s.dbClient.Database(s.mongoDatabase).Collection(s.mongoRichListCollection + "_" + epochString)

	results, err := queryRichListByRank(ctx, collection, start, limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot get rich list section from the database. error: %v", err)
	}

	// A rich list written before the processor stored the rank holds no rank to range over. Sorting
	// keeps such an epoch served until the next spectrum parse has rewritten the collection.
	if len(results) == 0 && start < totalRecords {
		results, err = queryRichListByBalance(ctx, collection, start, limit)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "cannot get rich list section from the database. error: %v", err)
		}
	}

	var list []*protobuff.RichListEntity

	for _, entity := range results {
		list = append(list, &protobuff.RichListEntity{
			Identity: entity.Identity,
			Balance:  entity.Balance,
		})
	}

	return &protobuff.GetRichListSliceResponse{
		Pagination: pagination,
		Epoch:      epoch,
		RichList: &protobuff.RichList{
			Entities: list,
		},
	}, nil

}

// queryRichListByRank reads a page of the rich list with an indexed range scan over the rank the
// processor stored, which costs nothing to page deeply into.
func queryRichListByRank(ctx context.Context, collection *mongo.Collection, start, limit int) (cache.RichList, error) {

	filter := bson.D{{Key: "rank", Value: bson.D{
		{Key: "$gte", Value: start},
		{Key: "$lt", Value: start + limit},
	}}}
	findOptions := options.Find().SetSort(bson.D{{Key: "rank", Value: 1}})

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, errors.Wrap(err, "querying rich list by rank")
	}

	var results cache.RichList
	if err := cursor.All(ctx, &results); err != nil {
		return nil, errors.Wrap(err, "decoding rich list")
	}

	return results, nil
}

// queryRichListByBalance is the pre rank way of reading a page and sorts the whole collection. It
// only serves rich lists that were written before the rank existed.
func queryRichListByBalance(ctx context.Context, collection *mongo.Collection, start, limit int) (cache.RichList, error) {

	findOptions := options.Find().
		SetSkip(int64(start)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "balance", Value: -1}}).
		SetAllowDiskUse(true) // the collection is neither indexed nor small enough to sort in memory

	cursor, err := collection.Find(ctx, bson.D{}, findOptions)
	if err != nil {
		return nil, errors.Wrap(err, "querying rich list by balance")
	}

	var results cache.RichList
	if err := cursor.All(ctx, &results); err != nil {
		return nil, errors.Wrap(err, "decoding rich list")
	}

	return results, nil
}

// supplyCap is the maximum amount of QUs that can ever exist.
const supplyCap int64 = 1_000_000_000_000_000

// maxSupplyHistoryPoints bounds how many points one supply history response may hold.
const maxSupplyHistoryPoints = 1000

func (s *Server) GetSupplyHistory(_ context.Context, request *protobuff.GetSupplyHistoryRequest) (*protobuff.GetSupplyHistoryResponse, error) {

	fromEpoch := request.GetFromEpoch()
	toEpoch := request.GetToEpoch()
	if toEpoch == 0 {
		toEpoch = math.MaxUint32
	}
	if fromEpoch > toEpoch {
		return nil, status.Errorf(codes.InvalidArgument, "fromEpoch (%d) is after toEpoch (%d)", request.GetFromEpoch(), request.GetToEpoch())
	}

	limit := int(request.GetLimit())
	if limit < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid limit (%d)", limit)
	}
	if limit == 0 || limit > maxSupplyHistoryPoints {
		limit = maxSupplyHistoryPoints
	}

	// The history is held in the cache, so this serves without touching the database.
	history := s.cache.GetSupplyHistory()

	points := make([]*protobuff.SupplyHistoryPoint, 0, min(len(history), limit))
	for _, record := range history {
		if record.Epoch < fromEpoch || record.Epoch > toEpoch {
			continue
		}
		points = append(points, &protobuff.SupplyHistoryPoint{
			Epoch:             record.Epoch,
			CirculatingSupply: record.CirculatingSupply,
			TotalEmitted:      record.TotalEmitted,
			Timestamp:         record.EpochEndTimestamp,
			SupplySource:      record.SupplySource,
		})
	}

	// The most recent points are the ones kept when the range holds more than the limit.
	if len(points) > limit {
		points = points[len(points)-limit:]
	}

	return &protobuff.GetSupplyHistoryResponse{
		SupplyCap:    supplyCap,
		CurrentEpoch: s.cache.GetQubicData().Epoch,
		// Read from the same place GetLatestData reads it, so that the two endpoints cannot disagree.
		CurrentCirculatingSupply: s.cache.GetSpectrumData().CirculatingSupply,
		Points:                   points,
	}, nil
}

type Pageable struct {
	Page, Size uint32
}

const maxPageSize uint32 = 1000
const defaultPageSize uint32 = 100

var assetNameRegexp, _ = regexp.Compile("^[A-Z0-9]{1,7}$")

func (s *Server) GetAssetOwners(ctx context.Context, req *protobuff.GetAssetOwnershipRequest) (*protobuff.GetAssetOwnershipResponse, error) {

	var pageSize uint32
	if req.GetPageSize() > maxPageSize {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid page size (maximum is %d).", maxPageSize)
	} else if req.GetPageSize() == 0 {
		pageSize = defaultPageSize
	} else {
		pageSize = req.GetPageSize()
	}

	// validate issuer identity
	err := validateIdentity(req.IssuerIdentity)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid issuer: %s", err.Error())
	}

	// validate asset name
	if len(req.AssetName) == 0 || !assetNameRegexp.MatchString(req.AssetName) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid asset name: %s", req.AssetName)
	}

	pageNumber := max(0, int(req.Page)-1) // API index starts with '1', implementation index starts with '0'.

	ownerships, tick, totalCount, err := s.assetService.GetOwnedAssets(ctx, req.GetIssuerIdentity(), req.GetAssetName(),
		Pageable{uint32(pageNumber), pageSize})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "getting asset ownerships: %s", err.Error())
	}

	pagination, err := getPaginationInformation(totalCount, pageNumber+1, int(pageSize))
	if err != nil {
		log.Printf("Error creating pagination info: %s", err.Error())
		return nil, status.Error(codes.Internal, "creating pagination info")
	}

	return &protobuff.GetAssetOwnershipResponse{
		Pagination: pagination,
		Tick:       tick,
		Owners:     ownerships,
	}, nil

}

func validateIdentity(identityString string) error {
	identity := types.Identity(identityString)
	pubKey, err := identity.ToPubKey(false)
	if err != nil {
		return err
	}
	reverse, err := identity.FromPubKey(pubKey, false)
	if err != nil {
		return err
	}
	if reverse != identity { // checksum errors are not checked in conversion
		return errors.Errorf("invalid identity [%s]", identityString)
	}
	return nil
}

// ATTENTION: first page has pageNumber == 1 as API starts with index 1
func getPaginationInformation(totalRecords, pageNumber, pageSize int) (*protobuff.Pagination, error) {

	if pageNumber < 1 {
		return nil, errors.Errorf("invalid page number [%d]", pageNumber)
	}

	if pageSize < 1 {
		return nil, errors.Errorf("invalid page size [%d]", pageSize)
	}

	if totalRecords < 0 {
		return nil, errors.Errorf("invalid number of total records [%d]", totalRecords)
	}

	totalPages := totalRecords / pageSize // rounds down
	if totalRecords%pageSize != 0 {
		totalPages += 1
	}

	// next page starts at index 1. -1 if no next page.
	nextPage := pageNumber + 1
	if nextPage > totalPages {
		nextPage = -1
	}

	// previous page starts at index 1. -1 if no previous page
	previousPage := pageNumber - 1
	if previousPage == 0 {
		previousPage = -1
	}

	currentPage := pageNumber
	if totalRecords == 0 {
		currentPage = 0
	}

	pagination := protobuff.Pagination{
		TotalRecords: int32(totalRecords),
		CurrentPage:  int32(currentPage), // 0 if there are no records
		TotalPages:   int32(totalPages),  // 0 if there are no records
		PageSize:     int32(pageSize),
	}
	return &pagination, nil
}

func (s *Server) Start() error {

	println("Starting GRPC server...")
	srv := grpc.NewServer(
		grpc.MaxRecvMsgSize(600*1024*1024),
		grpc.MaxSendMsgSize(600*1024*1024),
	)
	protobuff.RegisterStatsServiceServer(srv, s)
	reflection.Register(srv)

	lis, err := net.Listen("tcp", s.grpcAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	go func() {
		if err := srv.Serve(lis); err != nil {
			panic(err)
		}
	}()

	if s.httpAddress != "" {
		go func() {
			mux := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{EmitDefaultValues: true, EmitUnpopulated: false},
			}))
			opts := []grpc.DialOption{
				grpc.WithTransportCredentials(insecure.NewCredentials()),
				grpc.WithDefaultCallOptions(
					grpc.MaxCallRecvMsgSize(600*1024*1024),
					grpc.MaxCallSendMsgSize(600*1024*1024),
				),
			}

			if err := protobuff.RegisterStatsServiceHandlerFromEndpoint(
				context.Background(),
				mux,
				s.grpcAddress,
				opts,
			); err != nil {
				panic(err)
			}

			if err := http.ListenAndServe(s.httpAddress, mux); err != nil {
				panic(err)
			}
		}()
	}

	return nil

}

func NewServer(httpAddress string, grpcAddress string, cache *cache.Cache, dbClient *mongo.Client, database string, assetService *AssetServiceImpl, richListCollection string, richListPageSize int32, richListLimit int) *Server {
	return &Server{
		httpAddress:             httpAddress,
		grpcAddress:             grpcAddress,
		cache:                   cache,
		dbClient:                dbClient,
		mongoDatabase:           database,
		mongoRichListCollection: richListCollection,
		richListPageSize:        richListPageSize,
		assetService:            assetService,
		richListLimit:           richListLimit,
	}
}
