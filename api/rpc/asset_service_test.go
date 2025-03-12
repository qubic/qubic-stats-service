package rpc

import (
	"context"
	"encoding/hex"
	"github.com/jellydator/ttlcache/v3"
	qubic "github.com/qubic/go-node-connector"
	"github.com/qubic/go-node-connector/types"
	"github.com/stretchr/testify/assert"
	"log"
	"net"
	"testing"
	"time"
)

var tcpServer net.Listener
var assetService *AssetService
var assetOwnersCache *ttlcache.Cache[string, *types.AssetOwnerships]

var clientPoolCount = 0

type FakePool struct {
	client *qubic.Client
}

func (f FakePool) Get() (*qubic.Client, error) {
	clientPoolCount++
	return f.client, nil
}

func (f FakePool) Put(*qubic.Client) error {
	return nil
}

func (f FakePool) Close(*qubic.Client) error {
	return nil
}

func Test_AssetService_GetOwnedAssets_ReturnAll(t *testing.T) {

	setup(t)

	assets, tick, total, err := assetService.GetOwnedAssets(context.Background(), "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAFXIB", "TEST",
		Pageable{Page: 0, Size: 10})
	assert.NoError(t, err)
	assert.Len(t, assets, 7)
	assert.Equal(t, int(tick), 20192347)
	assert.Equal(t, total, 7)

	tearDown(t)

}

func Test_AssetService_GetOwnedAssets_CacheResponse(t *testing.T) {

	setup(t)

	clientPoolCount = 0          // clear count
	assetOwnersCache.DeleteAll() // clear cache

	for i := 0; i < 2; i++ {
		assets, _, _, err := assetService.GetOwnedAssets(context.Background(), "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAFXIB", "TEST",
			Pageable{Page: 0, Size: 10})
		assert.NoError(t, err)
		assert.Len(t, assets, 7)

		assert.Equal(t, 1, clientPoolCount) // client used only once
		assert.True(t, assetOwnersCache.Has("owners:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAFXIB:TEST"))
	}

	tearDown(t)

}

func Test_AssetService_GetOwnedAssets_ReturnPaginated(t *testing.T) {

	setup(t)

	assets, tick, total, err := assetService.GetOwnedAssets(context.Background(), "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAFXIB", "TEST",
		Pageable{Page: 0, Size: 5})
	assert.NoError(t, err)
	assert.Len(t, assets, 5)
	assert.Equal(t, int(tick), 20192347)
	assert.Equal(t, total, 7)

	assets, tick, total, err = assetService.GetOwnedAssets(context.Background(), "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAFXIB", "TEST",
		Pageable{Page: 1, Size: 5})
	assert.NoError(t, err)
	assert.Len(t, assets, 2)
	assert.Equal(t, int(tick), 20192347)
	assert.Equal(t, total, 7)

	assets, tick, total, err = assetService.GetOwnedAssets(context.Background(), "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAFXIB", "TEST",
		Pageable{Page: 6, Size: 1})
	assert.NoError(t, err)
	assert.Len(t, assets, 1)
	assert.Equal(t, int(tick), 20192347)
	assert.Equal(t, total, 7)

	tearDown(t)

}

func setup(t *testing.T) {
	assetOwnersCache = ttlcache.New[string, *types.AssetOwnerships](
		ttlcache.WithTTL[string, *types.AssetOwnerships](10*time.Second),
		ttlcache.WithDisableTouchOnHit[string, *types.AssetOwnerships](),
		ttlcache.WithCapacity[string, *types.AssetOwnerships](uint64(1024*1024)),
	)
	go assetOwnersCache.Start()

	startTcpServer(t)

	conn, err := net.DialTimeout("tcp", net.JoinHostPort("localhost", "12345"), time.Second)
	assert.NoError(t, err)

	//client, err := qubic.NewClient(context.Background(), "localhost", "12345")
	client, err := qubic.NewClientWithConn(context.Background(), conn)
	assert.NoError(t, err)

	pool := FakePool{
		client: client,
	}

	assetService = NewAssetService(pool, assetOwnersCache)
}

func tearDown(t *testing.T) {
	stopTcpServer(t)
}

func startTcpServer(t *testing.T) {
	log.Print("Starting tcp server")
	var err error
	tcpServer, err = net.Listen("tcp", "localhost:12345")
	if err != nil {
		log.Fatal(err)
	}

	go tcpAccept(t)
}

func tcpAccept(t *testing.T) {
	conn, err := tcpServer.Accept()
	assert.NoError(t, err)
	handleRequest(conn, t)
}

func handleRequest(conn net.Conn, t *testing.T) {
	log.Print("Received request")
	hexStr := "4000003511dcb7b9" + // RESPOND_ASSETS
		"7b5efffa039860590ecc801ab2f9a95da0b97592398d3414db1d3e44cac79d9a020001000400000004000000000000005b1c34017b5eff00" +
		"4000003511dcb7b9" + // RESPOND_ASSETS
		"feb0fb0e023c5f98ae9549112117ef3bf80608fcd252abc5772a07efd3f88b10020001000400000001000000000000005b1c3401feb0fb00" +
		"4000003511dcb7b9" + // RESPOND_ASSETS
		"feb0fb0e023c5f98ae9549112117ef3bf80608fcd252abc5772a07efd3f88b10020001000400000001000000000000005b1c3401feb0fb00" +
		"4000003511dcb7b9" + // RESPOND_ASSETS
		"feb0fb0e023c5f98ae9549112117ef3bf80608fcd252abc5772a07efd3f88b10020001000400000001000000000000005b1c3401feb0fb00" +
		"4000003511dcb7b9" + // RESPOND_ASSETS
		"feb0fb0e023c5f98ae9549112117ef3bf80608fcd252abc5772a07efd3f88b10020001000400000001000000000000005b1c3401feb0fb00" +
		"4000003511dcb7b9" + // RESPOND_ASSETS
		"feb0fb0e023c5f98ae9549112117ef3bf80608fcd252abc5772a07efd3f88b10020001000400000001000000000000005b1c3401feb0fb00" +
		"4000003511dcb7b9" + // RESPOND_ASSETS
		"feb0fb0e023c5f98ae9549112117ef3bf80608fcd252abc5772a07efd3f88b10020001000400000001000000000000005b1c3401feb0fb00" +
		"0800002328af10a4" // END_RESPONSE
	ownedAssetsBin, err := hex.DecodeString(hexStr)
	assert.NoError(t, err)
	_, err = conn.Write(ownedAssetsBin)
	assert.NoError(t, err)
	err = conn.Close()
	assert.NoError(t, err)
}

func stopTcpServer(t *testing.T) {
	log.Print("Stopping tcp server")
	err := tcpServer.Close()
	assert.NoError(t, err)
}
