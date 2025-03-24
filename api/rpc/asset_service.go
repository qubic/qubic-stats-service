package rpc

import (
	"cmp"
	"context"
	"fmt"
	"github.com/jellydator/ttlcache/v3"
	"github.com/pkg/errors"
	qubic "github.com/qubic/go-node-connector"
	"github.com/qubic/go-node-connector/types"
	"github.com/qubic/qubic-stats-api/protobuff"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"slices"
)

type AssetService interface {
	GetOwnedAssets(ctx context.Context, issuerIdentity, assetName string, page Pageable) ([]*protobuff.AssetOwnership, uint32, int, error)
}

type AssetServiceImpl struct {
	qPool           ClientPool
	assetOwnerCache *ttlcache.Cache[string, *types.AssetOwnerships]
}

type ClientPool interface {
	Get() (*qubic.Client, error)
	Close(*qubic.Client) error
	Put(*qubic.Client) error
}

func NewAssetService(qPool ClientPool, assetOwnersCache *ttlcache.Cache[string, *types.AssetOwnerships]) *AssetServiceImpl {
	service := AssetServiceImpl{
		qPool:           qPool,
		assetOwnerCache: assetOwnersCache,
	}
	return &service
}

const ownersCacheKey string = "owners:%s:%s"

func (s *AssetServiceImpl) GetOwnedAssets(ctx context.Context, issuerIdentity, assetName string, page Pageable) ([]*protobuff.AssetOwnership, uint32, int, error) {

	retrievedAssets, err := s.getAssetOwners(ctx, issuerIdentity, assetName)
	if err != nil {
		return nil, 0, -1, status.Errorf(codes.Internal, "retrieving asset owners: %v", err)
	}
	assets := *retrievedAssets

	slices.SortFunc(assets, func(a, b types.AssetOwnership) int {
		return -cmp.Compare(a.Asset.NumberOfUnits, b.Asset.NumberOfUnits)
	})

	start := int(page.Page) * int(page.Size)
	end := start + int(page.Size)
	endIndex := min(end, len(assets))
	startIndex := min(endIndex, start)
	assetsSlice := assets[startIndex:endIndex]

	ownerships := make([]*protobuff.AssetOwnership, 0)
	var tick uint32
	for _, asset := range assetsSlice {

		var owner types.Identity
		owner, err := owner.FromPubKey(asset.Asset.PublicKey, false)
		if err != nil {
			return nil, 0, -1, errors.Wrap(err, "failed to get identity for public key")
		}

		assetOwnership := protobuff.AssetOwnership{
			Identity:       owner.String(),
			NumberOfShares: asset.Asset.NumberOfUnits,
		}

		tick = max(asset.Tick, tick)
		ownerships = append(ownerships, &assetOwnership)

	}

	return ownerships, tick, len(assets), nil
}

func (s *AssetServiceImpl) getAssetOwners(ctx context.Context, issuerIdentity, assetName string) (*types.AssetOwnerships, error) {
	key := cacheKey(issuerIdentity, assetName)
	var assets *types.AssetOwnerships
	if s.assetOwnerCache.Has(key) {
		assets = s.assetOwnerCache.Get(key).Value()
	}
	if assets == nil {
		queriedAssets, err := s.getAssetOwnersFromNode(ctx, issuerIdentity, assetName)
		if err != nil {
			return nil, err
		}
		if len(*queriedAssets) > 0 {
			// we only cache queries that return data
			s.assetOwnerCache.Set(key, queriedAssets, ttlcache.DefaultTTL)
		}
		assets = queriedAssets
	}
	return assets, nil
}

func (s *AssetServiceImpl) getAssetOwnersFromNode(ctx context.Context, identity string, name string) (*types.AssetOwnerships, error) {
	client, err := s.qPool.Get()
	if err != nil {
		return nil, errors.Wrap(err, "getting pool connection")
	}
	assets, err := client.GetAssetOwnershipsByFilter(ctx, identity, name, "", 0)
	if err != nil {
		_ = s.qPool.Close(client)
		return nil, errors.Wrap(err, "getting asset ownerships")
	}
	err = s.qPool.Put(client)
	if err != nil {
		log.Printf("WARN: error returning client to pool: %v", err)
	}
	return &assets, nil
}

func cacheKey(issuerIdentity, assetName string) string {
	return fmt.Sprintf(ownersCacheKey, issuerIdentity, assetName)
}
