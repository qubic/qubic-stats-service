package rpc

import (
	"cmp"
	"context"
	"github.com/pkg/errors"
	"github.com/qubic/go-node-connector/types"
	"github.com/qubic/qubic-stats-api/protobuff"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"slices"
)

func (s *Server) GetOwnedAssets(ctx context.Context, issuerIdentity, assetName string, page Pageable) ([]*protobuff.AssetOwnership, uint32, int, error) {

	start := int(page.Page) * int(page.Size)
	end := start + int(page.Size)

	client, err := s.qPool.Get()
	if err != nil {
		return nil, 0, -1, status.Errorf(codes.Internal, "getting pool connection: %v", err)
	}

	assets, err := client.GetAssetOwnershipsByFilter(ctx, issuerIdentity, assetName, "", 0)
	if err != nil {
		s.qPool.Close(client)
		return nil, 0, -1, status.Errorf(codes.Internal, "getting asset ownerships: %v", err)
	}
	s.qPool.Put(client)

	ownerships := make([]*protobuff.AssetOwnership, 0)

	var tick uint32

	slices.SortFunc(assets, func(a, b types.AssetOwnership) int {
		return -cmp.Compare(a.Asset.NumberOfUnits, b.Asset.NumberOfUnits)
	})

	endIndex := min(end, len(assets))
	startIndex := min(endIndex, start)
	assetsSlice := assets[startIndex:endIndex]

	for _, asset := range assetsSlice {

		var owner types.Identity
		owner, err := owner.FromPubKey(asset.Asset.PublicKey, false)
		if err != nil {
			return nil, 0, -1, errors.Wrap(err, "failed to get identity for public key")
		}

		assetOwnership := protobuff.AssetOwnership{
			Identity:       owner.String(),
			NumberOfShares: uint32(asset.Asset.NumberOfUnits),
		}

		tick = max(asset.Tick, tick)
		ownerships = append(ownerships, &assetOwnership)

	}

	return ownerships, tick, len(assets), nil
}
