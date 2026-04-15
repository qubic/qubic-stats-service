package rpc

import (
	"testing"

	"github.com/qubic/go-node-connector/types"
	"github.com/stretchr/testify/assert"
)

func Test_AssetService_CacheKey(t *testing.T) {
	got := cacheKey("ISSUER", "ASSET")
	want := "owners:ISSUER:ASSET"
	if got != want {
		t.Errorf("cacheKey() = %q, want %q", got, want)
	}
}

func Test_AssetService_CombineOwnedAssets(t *testing.T) {

	ownerships := types.AssetOwnerships{
		{
			Asset: types.AssetOwnershipData{
				PublicKey:             [32]byte{7, 8, 9},
				Type:                  3,
				Padding:               [1]int8{},
				ManagingContractIndex: 0,
				IssuanceIndex:         0,
				NumberOfUnits:         1000,
			},
			Tick:          1,
			UniverseIndex: 1,
		},
		{
			Asset: types.AssetOwnershipData{
				PublicKey:             [32]byte{1, 2, 3},
				Type:                  3,
				Padding:               [1]int8{},
				ManagingContractIndex: 1,
				IssuanceIndex:         1,
				NumberOfUnits:         1000,
			},
			Tick:          1,
			UniverseIndex: 1,
		}, {
			Asset: types.AssetOwnershipData{
				PublicKey:             [32]byte{6, 6, 6},
				Type:                  3,
				Padding:               [1]int8{},
				ManagingContractIndex: 2,
				IssuanceIndex:         2,
				NumberOfUnits:         100,
			},
			Tick:          1,
			UniverseIndex: 2,
		}, {
			Asset: types.AssetOwnershipData{
				PublicKey:             [32]byte{1, 2, 3},
				Type:                  3,
				Padding:               [1]int8{},
				ManagingContractIndex: 3,
				IssuanceIndex:         3,
				NumberOfUnits:         10,
			},
			Tick:          1,
			UniverseIndex: 3,
		},
	}

	combined, err := combineEntriesForSameIdentity(ownerships)
	assert.NoError(t, err)

	expected := types.AssetOwnerships{
		{
			Asset: types.AssetOwnershipData{
				PublicKey:             [32]byte{1, 2, 3},
				Type:                  3,
				Padding:               [1]int8{},
				ManagingContractIndex: 1,
				IssuanceIndex:         1,
				NumberOfUnits:         1010,
			},
			Tick:          1,
			UniverseIndex: 1,
		}, {
			Asset: types.AssetOwnershipData{
				PublicKey:             [32]byte{7, 8, 9},
				Type:                  3,
				Padding:               [1]int8{},
				ManagingContractIndex: 0,
				IssuanceIndex:         0,
				NumberOfUnits:         1000,
			},
			Tick:          1,
			UniverseIndex: 1,
		}, {
			Asset: types.AssetOwnershipData{
				PublicKey:             [32]byte{6, 6, 6},
				Type:                  3,
				Padding:               [1]int8{},
				ManagingContractIndex: 2,
				IssuanceIndex:         2,
				NumberOfUnits:         100,
			},
			Tick:          1,
			UniverseIndex: 2,
		},
	}

	assert.Equal(t, expected, *combined)

}
