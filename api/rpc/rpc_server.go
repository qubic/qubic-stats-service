package rpc

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/qubic/qubic-stats-api/cache"
	"github.com/qubic/qubic-stats-api/protobuff"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"net"
	"net/http"
)

type Server struct {
	protobuff.UnimplementedStatsServiceServer
	httpAddress string
	grpcAddress string

	cache *cache.Cache

	dbClient                *mongo.Client
	mongoDatabase           string
	mongoRichListCollection string
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
		},
	}, nil

}

func (s *Server) GetRichListLength(ctx context.Context, request *protobuff.GetRichListLengthRequest) (*protobuff.GetRichListLengthResponse, error) {
	collection := s.dbClient.Database(s.mongoDatabase).Collection(s.mongoRichListCollection + "_" + request.Epoch)

	count, err := collection.CountDocuments(ctx, bson.D{}, options.Count().SetHint("_id_"))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to obtain length of rich list")
	}

	return &protobuff.GetRichListLengthResponse{
		Length: uint32(count),
	}, err

}

func (s *Server) GetRichListSlice(ctx context.Context, request *protobuff.GetRichListSliceRequest) (*protobuff.GetRichListSliceResponse, error) {

	collection := s.dbClient.Database(s.mongoDatabase).Collection(s.mongoRichListCollection + "_" + request.Epoch)

	skip := int64(request.Start)
	limit := int64(request.End - request.Start)

	findOptions := options.Find().SetSkip(skip).SetLimit(limit).SetSort(bson.D{{"balance", -1}})

	cursor, err := collection.Find(ctx, bson.D{{}}, findOptions)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot get rich list slice from the database. error: %v", err)
	}

	var results cache.RichList
	err = cursor.All(ctx, &results)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to unmarshal database response. error: %v", err)
	}

	var list []*protobuff.RichListEntity

	for _, entity := range results {
		list = append(list, &protobuff.RichListEntity{
			Identity: entity.Identity,
			Balance:  entity.Balance,
		})
	}

	return &protobuff.GetRichListSliceResponse{
		RichList: &protobuff.RichList{
			Entities: list,
		},
	}, nil

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

func NewServer(httpAddress string, grpcAddress string, cache *cache.Cache, dbClient *mongo.Client, database string, richListCollection string) *Server {
	return &Server{
		httpAddress:             httpAddress,
		grpcAddress:             grpcAddress,
		cache:                   cache,
		dbClient:                dbClient,
		mongoDatabase:           database,
		mongoRichListCollection: richListCollection,
	}
}
