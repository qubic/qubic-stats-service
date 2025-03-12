package rpc

import (
	"github.com/qubic/qubic-stats-api/protobuff"
	"github.com/stretchr/testify/assert"
	"testing"
)

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
