package rpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jellydator/ttlcache/v3"
	"github.com/qubic/go-node-connector/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// stubFetcher implements live.AssetFetcher for testing.
type stubFetcher struct {
	ownerships *types.AssetOwnerships
	err        error
	callCount  int
}

func (s *stubFetcher) GetAssetOwnerships(_ context.Context, _, _ string) (*types.AssetOwnerships, error) {
	s.callCount++
	return s.ownerships, s.err
}

func newCache() *ttlcache.Cache[string, *types.AssetOwnerships] {
	return ttlcache.New[string, *types.AssetOwnerships](
		ttlcache.WithTTL[string, *types.AssetOwnerships](time.Minute),
	)
}

func pubKey(b byte) [32]byte {
	var key [32]byte
	key[0] = b
	return key
}

func makeOwnership(pk [32]byte, units int64, tick uint32) types.AssetOwnership {
	return types.AssetOwnership{
		Asset: types.AssetOwnershipData{PublicKey: pk, NumberOfUnits: units},
		Tick:  tick,
	}
}

func TestGetOwnedAssets_ReturnsSortedBySharesOwnerships(t *testing.T) {
	ownerships := types.AssetOwnerships{
		makeOwnership(pubKey(1), 100, 10),
		makeOwnership(pubKey(2), 500, 20),
		makeOwnership(pubKey(3), 200, 15),
	}
	svc := NewAssetService(&stubFetcher{ownerships: &ownerships}, newCache())

	got, tick, total, err := svc.GetOwnedAssets(context.Background(), "ISSUER", "ASSET", Pageable{Page: 0, Size: 10})
	require.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Equal(t, uint32(20), tick)
	require.Len(t, got, 3)
	assert.Equal(t, int64(500), got[0].NumberOfShares)
	assert.Equal(t, int64(200), got[1].NumberOfShares)
	assert.Equal(t, int64(100), got[2].NumberOfShares)
}

func TestGetOwnedAssets_ReturnsSortedByIdentity(t *testing.T) {
	ownerships := types.AssetOwnerships{
		makeOwnership(pubKey(0), 99, 15),
		makeOwnership(pubKey(1), 99, 15),
		makeOwnership(pubKey(2), 100, 15),
		makeOwnership(pubKey(3), 100, 15),
		makeOwnership(pubKey(4), 100, 15),
		makeOwnership(pubKey(5), 100, 15),
		makeOwnership(pubKey(6), 123, 15),
	}
	svc := NewAssetService(&stubFetcher{ownerships: &ownerships}, newCache())

	got, _, total, err := svc.GetOwnedAssets(context.Background(), "ISSUER", "ASSET", Pageable{Page: 0, Size: 10})
	require.NoError(t, err)
	assert.Equal(t, 7, total)
	require.Len(t, got, 7)
	// sorted by shares (desc) and then by identity (asc)
	assert.Equal(t, int64(123), got[0].NumberOfShares)
	assert.Equal(t, "GAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQGNM", got[0].GetIdentity())
	assert.Equal(t, int64(100), got[1].NumberOfShares)
	assert.Equal(t, "CAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACNKL", got[1].GetIdentity())
	assert.Equal(t, int64(100), got[2].NumberOfShares)
	assert.Equal(t, "DAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAANMIG", got[2].GetIdentity())
	assert.Equal(t, int64(100), got[3].NumberOfShares)
	assert.Equal(t, "EAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAVWRF", got[3].GetIdentity())
	assert.Equal(t, int64(100), got[4].NumberOfShares)
	assert.Equal(t, "FAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAYWJB", got[4].GetIdentity())
	assert.Equal(t, int64(99), got[5].NumberOfShares)
	assert.Equal(t, "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAFXIB", got[5].GetIdentity())
	assert.Equal(t, int64(99), got[6].NumberOfShares)
	assert.Equal(t, "BAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAARMID", got[6].GetIdentity())
}

func TestGetOwnedAssets_ReturnsCombinedSortedByIdentity(t *testing.T) {
	// entries for same pub key get combined
	ownerships := types.AssetOwnerships{
		makeOwnership(pubKey(0), 60, 11),
		makeOwnership(pubKey(0), 40, 11),
		makeOwnership(pubKey(1), 50, 11),
		makeOwnership(pubKey(1), 50, 11),
		makeOwnership(pubKey(2), 99, 11),
		makeOwnership(pubKey(2), 1, 11),
		makeOwnership(pubKey(3), 100, 11),
		makeOwnership(pubKey(4), 70, 11),
		makeOwnership(pubKey(4), 30, 11),
	}
	svc := NewAssetService(&stubFetcher{ownerships: &ownerships}, newCache())

	got, _, total, err := svc.GetOwnedAssets(context.Background(), "ISSUER", "ASSET", Pageable{Page: 0, Size: 10})
	require.NoError(t, err)
	assert.Equal(t, 5, total)
	require.Len(t, got, 5)
	assert.Equal(t, int64(100), got[0].NumberOfShares)
	assert.Equal(t, "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAFXIB", got[0].GetIdentity())
	assert.Equal(t, int64(100), got[1].NumberOfShares)
	assert.Equal(t, "BAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAARMID", got[1].GetIdentity())
	assert.Equal(t, int64(100), got[1].NumberOfShares)
	assert.Equal(t, "CAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACNKL", got[2].GetIdentity())
	assert.Equal(t, int64(100), got[2].NumberOfShares)
	assert.Equal(t, "DAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAANMIG", got[3].GetIdentity())
	assert.Equal(t, int64(100), got[3].NumberOfShares)
	assert.Equal(t, "EAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAVWRF", got[4].GetIdentity())
	assert.Equal(t, int64(100), got[4].NumberOfShares)
}

func TestGetOwnedAssets_IdentityStrings(t *testing.T) {
	ownerships := types.AssetOwnerships{
		makeOwnership(pubKey(1), 100, 5),
		makeOwnership(pubKey(2), 200, 5),
	}
	svc := NewAssetService(&stubFetcher{ownerships: &ownerships}, newCache())

	got, _, _, err := svc.GetOwnedAssets(context.Background(), "ISSUER", "ASSET", Pageable{Page: 0, Size: 10})
	require.NoError(t, err)

	// sorted descending, so pubKey(2)=200 comes first
	assert.Equal(t, "CAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACNKL", got[0].Identity)
	assert.Equal(t, "BAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAARMID", got[1].Identity)
}

func TestGetOwnedAssets_Pagination(t *testing.T) {
	ownerships := types.AssetOwnerships{
		makeOwnership(pubKey(1), 500, 1),
		makeOwnership(pubKey(2), 400, 2),
		makeOwnership(pubKey(3), 300, 3),
		makeOwnership(pubKey(4), 200, 4),
		makeOwnership(pubKey(5), 100, 5),
	}
	svc := NewAssetService(&stubFetcher{ownerships: &ownerships}, newCache())
	ctx := context.Background()

	page0, _, total, err := svc.GetOwnedAssets(ctx, "ISSUER", "ASSET", Pageable{Page: 0, Size: 2})
	require.NoError(t, err)
	assert.Equal(t, 5, total)
	require.Len(t, page0, 2)
	assert.Equal(t, int64(500), page0[0].NumberOfShares)
	assert.Equal(t, int64(400), page0[1].NumberOfShares)

	page1, _, _, err := svc.GetOwnedAssets(ctx, "ISSUER", "ASSET", Pageable{Page: 1, Size: 2})
	require.NoError(t, err)
	require.Len(t, page1, 2)
	assert.Equal(t, int64(300), page1[0].NumberOfShares)
	assert.Equal(t, int64(200), page1[1].NumberOfShares)
}

func TestGetOwnedAssets_OutOfBoundsPageReturnsEmpty(t *testing.T) {
	ownerships := types.AssetOwnerships{makeOwnership(pubKey(1), 100, 1)}
	svc := NewAssetService(&stubFetcher{ownerships: &ownerships}, newCache())

	got, _, total, err := svc.GetOwnedAssets(context.Background(), "ISSUER", "ASSET", Pageable{Page: 5, Size: 10})
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Empty(t, got)
}

func TestGetOwnedAssets_FetcherErrorReturnsGRPCInternal(t *testing.T) {
	svc := NewAssetService(&stubFetcher{err: errors.New("node unreachable")}, newCache())

	_, _, _, err := svc.GetOwnedAssets(context.Background(), "ISSUER", "ASSET", Pageable{Page: 0, Size: 10})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok, "expected gRPC status error, got: %v", err)
	assert.Equal(t, codes.Internal, st.Code())
}

func TestGetOwnedAssets_EmptyResultNotCached(t *testing.T) {
	empty := types.AssetOwnerships{}
	stub := &stubFetcher{ownerships: &empty}
	svc := NewAssetService(stub, newCache())
	ctx := context.Background()

	_, _, _, _ = svc.GetOwnedAssets(ctx, "ISSUER", "ASSET", Pageable{Page: 0, Size: 10})
	_, _, _, _ = svc.GetOwnedAssets(ctx, "ISSUER", "ASSET", Pageable{Page: 0, Size: 10})

	assert.Equal(t, 2, stub.callCount, "empty results must not be cached")
}

func TestGetOwnedAssets_NonEmptyResultIsCached(t *testing.T) {
	ownerships := types.AssetOwnerships{makeOwnership(pubKey(1), 100, 1)}
	stub := &stubFetcher{ownerships: &ownerships}
	svc := NewAssetService(stub, newCache())
	ctx := context.Background()

	_, _, _, _ = svc.GetOwnedAssets(ctx, "ISSUER", "ASSET", Pageable{Page: 0, Size: 10})
	_, _, _, _ = svc.GetOwnedAssets(ctx, "ISSUER", "ASSET", Pageable{Page: 0, Size: 10})

	assert.Equal(t, 1, stub.callCount, "non-empty result should be cached")
}

func TestCombineEntriesForSameIdentity_SumsDuplicates(t *testing.T) {
	pk := pubKey(1)
	result, err := combineEntriesForSameIdentity([]types.AssetOwnership{
		makeOwnership(pk, 100, 1),
		makeOwnership(pk, 200, 2),
	})
	require.NoError(t, err)
	require.Len(t, *result, 1)
	assert.Equal(t, int64(300), (*result)[0].Asset.NumberOfUnits)
}

func TestCombineEntriesForSameIdentity_KeepsDistinctIdentities(t *testing.T) {
	result, err := combineEntriesForSameIdentity([]types.AssetOwnership{
		makeOwnership(pubKey(1), 100, 1),
		makeOwnership(pubKey(2), 200, 2),
	})
	require.NoError(t, err)
	assert.Len(t, *result, 2)
}

func TestCacheKey(t *testing.T) {
	assert.Equal(t, "owners:ISSUER:ASSET", cacheKey("ISSUER", "ASSET"))
}
