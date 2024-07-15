package rpc

import (
	"context"
	"github.com/qubic/qubic-stats-api/protobuff"
	"google.golang.org/protobuf/types/known/emptypb"
)

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
