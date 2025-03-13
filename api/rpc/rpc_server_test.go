package rpc

import (
	"context"
	"github.com/qubic/qubic-stats-api/protobuff"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

type FakeAssetService struct {
	AssetOwnerships []*protobuff.AssetOwnership
	Tick            uint32
	Err             error
}

func (fas FakeAssetService) GetOwnedAssets(context.Context, string, string, Pageable) ([]*protobuff.AssetOwnership, uint32, int, error) {
	return fas.AssetOwnerships, fas.Tick, len(fas.AssetOwnerships), fas.Err
}

func Test_RpcServer_GetAssetOwners(t *testing.T) {

	assetOwnerships := []*protobuff.AssetOwnership{
		{
			Identity:       "A",
			NumberOfShares: 123,
		},
		{
			Identity:       "B",
			NumberOfShares: 12,
		},
		{
			Identity:       "C",
			NumberOfShares: 1,
		},
	}

	server := Server{
		assetService: FakeAssetService{
			AssetOwnerships: assetOwnerships,
			Tick:            12345,
			Err:             nil,
		},
	}

	request := protobuff.GetAssetOwnershipRequest{
		IssuerIdentity: "TESTQXXOYTGAFHJXAHLTGDXUIPQCKBYFUAAUXJEWCFIXBGPJBRCHCJNDMWYI",
		AssetName:      "TEST",
		Page:           0,
		PageSize:       0,
	}

	response, err := server.GetAssetOwners(context.Background(), &request)
	assert.NoError(t, err)

	assert.Equal(t, 12345, int(response.Tick))
	assert.Equal(t, &protobuff.Pagination{
		TotalRecords: 3,
		CurrentPage:  1,
		TotalPages:   1,
		PageSize:     100,
	}, response.Pagination)
	assert.Equal(t, assetOwnerships, response.Owners)

}

func Test_RpcServer_GetAssetOwners_InvalidAssetName_thenError(t *testing.T) {
	server := Server{}

	invalidAssetNames := []string{"NAMETOOLONG", "lower", "&*^&*^", "NO$Äö"}

	for _, assetName := range invalidAssetNames {
		request := protobuff.GetAssetOwnershipRequest{
			IssuerIdentity: "TESTQXXOYTGAFHJXAHLTGDXUIPQCKBYFUAAUXJEWCFIXBGPJBRCHCJNDMWYI",
			AssetName:      assetName,
			Page:           0,
			PageSize:       0,
		}
		_, err := server.GetAssetOwners(context.Background(), &request)
		assert.ErrorContains(t, err, "invalid asset name")
	}
}

func Test_RpcServer_GetAssetOwners_InvalidIdentity_thenError(t *testing.T) {
	server := Server{}

	invalidIdentities := []string{
		"TESTQXXWWWWRONGCHECKSUMMMMMEKBYFUAAUXJEWCFIXBGPJBRCHCJNDMWYI",
		"SHORT",
		"testqxxoytgafhjxahltgdxuipqckbyfuaauxjewcfixbgpjbrchcjndmwyi",
	}

	for _, identity := range invalidIdentities {
		log.Println(identity)
		request := protobuff.GetAssetOwnershipRequest{
			IssuerIdentity: identity,
			AssetName:      "TEST",
			Page:           0,
			PageSize:       0,
		}
		_, err := server.GetAssetOwners(context.Background(), &request)
		assert.ErrorContains(t, err, "invalid issuer")
	}
}

func Test_RpcServerPagination_createPaginationInfo(t *testing.T) {

	pagination, err := getPaginationInformation(0, 1, 100)
	assert.NoError(t, err)
	verify(t, pagination, 0, 0, 100, 0)

	pagination, err = getPaginationInformation(1, 1, 100)
	assert.NoError(t, err)
	verify(t, pagination, 1, 1, 100, 1)

	pagination, err = getPaginationInformation(12345, 1, 100)
	assert.NoError(t, err)
	verify(t, pagination, 12345, 1, 100, 124)

	pagination, err = getPaginationInformation(12345, 10, 100)
	assert.NoError(t, err)
	verify(t, pagination, 12345, 10, 100, 124)

	pagination, err = getPaginationInformation(12345, 124, 100)
	assert.NoError(t, err)
	verify(t, pagination, 12345, 124, 100, 124)

	pagination, err = getPaginationInformation(12345, 125, 100)
	assert.NoError(t, err)
	verify(t, pagination, 12345, 125, 100, 124)

	pagination, err = getPaginationInformation(12345, 999, 100)
	assert.NoError(t, err)
	verify(t, pagination, 12345, 999, 100, 124)

	pagination, err = getPaginationInformation(121, 122, 1)
	assert.NoError(t, err)
	verify(t, pagination, 121, 122, 1, 121)
}

func Test_RpcServerPagination_createPaginationInfo_givenInvalidArguments_thenError(t *testing.T) {
	_, err := getPaginationInformation(12345, 0, 100)
	assert.Error(t, err)

	_, err = getPaginationInformation(-1, 1, 100)
	assert.Error(t, err)

	_, err = getPaginationInformation(12345, 1, 0)
	assert.Error(t, err)
}

func verify(t *testing.T, pagination *protobuff.Pagination, totalRecords, pageNumber, pageSize, totalPages int) {
	assert.NotNil(t, pagination)
	assert.Equal(t, int32(totalRecords), pagination.TotalRecords, "unexpected number of total records")
	assert.Equal(t, int32(pageNumber), pagination.CurrentPage, "unexpected current page number")
	assert.Equal(t, int32(pageSize), pagination.PageSize, "unexpected page size")
	assert.Equal(t, int32(totalPages), pagination.TotalPages, "unexpected number of total pages")
}
