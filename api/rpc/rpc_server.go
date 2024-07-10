package rpc

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"github.com/qubic/qubic-stats-api/db"
	"github.com/qubic/qubic-stats-api/protobuff"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
	"log"
	"net"
	"net/http"
	"time"
)

type ServerConfiguration struct {
	HttpAddress string
	GrpcAddress string

	Connection         *db.Connection
	Database           string
	SpectrumCollection string
	DataCollection     string

	CacheStrategy              string
	CacheValidityDuration      time.Duration
	SpectrumDataUpdateInterval time.Duration
}

type Server struct {
	protobuff.UnimplementedStatsServiceServer

	httpAddress string
	grpcAddress string

	connection         *db.Connection
	database           string
	spectrumCollection string
	dataCollection     string

	dbClient *mongo.Client

	dataCache                  DataCache
	cachingStrategy            string
	cacheValidityDuration      time.Duration
	spectrumDataUpdateInterval time.Duration
}

func (s *Server) Start() error {

	println("Starting Web Service...")

	//fmt.Printf("Caching strategy: %s\n", s.cachingStrategy)
	fmt.Printf("Cache validity duration: %v\n", s.cacheValidityDuration)

	println("Connecting to database...")
	err := s.createDatabaseClient()
	if err != nil {
		return errors.Wrap(err, "creating database client")
	}
	defer func() {
		if err = s.dbClient.Disconnect(context.Background()); err != nil {
			log.Fatalf("service: exited with error: %s\n", err.Error())
		}
	}()

	println("Connected to database.")

	println("Fetching initial data...")
	err = s.updateCache(true, true)
	if err != nil {
		return errors.Wrap(err, "initializing cache")
	}

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

	exit := s.startUpdateService()
	<-exit

	return nil

}

func NewServer(configuration *ServerConfiguration) *Server {
	return &Server{
		httpAddress:                configuration.HttpAddress,
		grpcAddress:                configuration.GrpcAddress,
		connection:                 configuration.Connection,
		database:                   configuration.Database,
		spectrumCollection:         configuration.SpectrumCollection,
		dataCollection:             configuration.DataCollection,
		cachingStrategy:            configuration.CacheStrategy,
		cacheValidityDuration:      configuration.CacheValidityDuration,
		spectrumDataUpdateInterval: configuration.SpectrumDataUpdateInterval,
	}
}
