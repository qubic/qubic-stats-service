package rpc

import (
	"bytes"
	"cmp"
	"context"
	"fmt"
	"slices"

	"github.com/jellydator/ttlcache/v3"
	"github.com/qubic/go-node-connector/types"
	"github.com/qubic/qubic-stats-api/live"
	"github.com/qubic/qubic-stats-api/protobuff"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AssetService interface {
	GetOwnedAssets(ctx context.Context, issuerIdentity, assetName string, page Pageable) ([]*protobuff.AssetOwnership, uint32, int, error)
}

type AssetServiceImpl struct {
	fetcher         live.AssetFetcher
	assetOwnerCache *ttlcache.Cache[string, *types.AssetOwnerships]
}

func NewAssetService(fetcher live.AssetFetcher, assetOwnersCache *ttlcache.Cache[string, *types.AssetOwnerships]) *AssetServiceImpl {
	return &AssetServiceImpl{
		fetcher:         fetcher,
		assetOwnerCache: assetOwnersCache,
	}
}

const ownersCacheKey string = "owners:%s:%s"

func (s *AssetServiceImpl) GetOwnedAssets(ctx context.Context, issuerIdentity, assetName string, page Pageable) ([]*protobuff.AssetOwnership, uint32, int, error) {

	retrievedAssets, err := s.getAssetOwners(ctx, issuerIdentity, assetName)
	if err != nil {
		return nil, 0, -1, status.Errorf(codes.Internal, "retrieving asset owners: %v", err)
	}
	assets := *retrievedAssets

	slices.SortFunc(assets, func(a, b types.AssetOwnership) int {
		if c := -cmp.Compare(a.Asset.NumberOfUnits, b.Asset.NumberOfUnits); c != 0 {
			return c
		}
		return bytes.Compare(a.Asset.PublicKey[:], b.Asset.PublicKey[:])
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
			return nil, 0, -1, fmt.Errorf("failed to get identity for public key: %w", err)
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
		queriedAssets, err := s.fetcher.GetAssetOwnerships(ctx, issuerIdentity, assetName)
		if err != nil {
			return nil, err
		}
		queriedAssets, err = combineEntriesForSameIdentity(*queriedAssets)
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

// combineEntriesForSameIdentity there might be several entries for one identity and different managing contracts
func combineEntriesForSameIdentity(ownerships []types.AssetOwnership) (*types.AssetOwnerships, error) {
	var identityMap = make(map[[32]byte]*types.AssetOwnership)

	// combine multiple ownerships for the same identity into one
	for _, ownership := range ownerships {
		val, found := identityMap[ownership.Asset.PublicKey]
		if !found {
			identityMap[ownership.Asset.PublicKey] = &ownership
		} else {
			val.Asset.NumberOfUnits += ownership.Asset.NumberOfUnits
		}
	}

	// create combined ownership list
	var combined = make(types.AssetOwnerships, 0, len(identityMap))
	for _, v := range identityMap {
		combined = append(combined, *v)
	}

	return &combined, nil
}

func cacheKey(issuerIdentity, assetName string) string {
	return fmt.Sprintf(ownersCacheKey, issuerIdentity, assetName)
}
