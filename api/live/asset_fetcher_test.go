package live

import (
	"context"
	"encoding/hex"
	"log"
	"net"
	"testing"
	"time"

	qubic "github.com/qubic/go-node-connector"
	"github.com/stretchr/testify/assert"
)

var tcpServer net.Listener
var assetService AssetFetcher
var clientPoolCount = 0

type FakePool struct {
}

func (f FakePool) Get() (*qubic.Client, error) {

	conn, err := net.DialTimeout("tcp", net.JoinHostPort("localhost", "12345"), time.Second)
	if err != nil {
		log.Fatal("creating connection")
	}
	client, err := qubic.NewClientWithConn(context.Background(), conn)
	if err != nil {
		log.Fatal("creating client")
	}

	clientPoolCount++
	return client, nil
}

func (f FakePool) Put(*qubic.Client) error {
	return nil
}

func (f FakePool) Close(*qubic.Client) error {
	return nil
}

func Test_AssetService_GetOwnedAssets_ReturnAll(t *testing.T) {

	setup(t)

	assets, err := assetService.GetAssetOwnerships(context.Background(), "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAFXIB", "TEST")
	assert.NoError(t, err)
	assert.Len(t, *assets, 7)

	tearDown(t)

}

func setup(t *testing.T) {
	startTcpServer(t) // this is a bit hacky. only serves one response.
	assetService = NewAssetClient(FakePool{})
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
		"feb0fb0e023c5f98ae9549112117ef3bf80608fcd252abc5772a07efd3f88b10020001000400000001000000000000005b1c3401feb0fb00" + // 1 share / tick number 20192347 (test data)
		"4000003511dcb7b9" + // RESPOND_ASSETS
		"7b5efffa039860590ecc801ab2f9a95da0b97592398d3414db1d3e44cac79d9a020001000400000004000000000000005b1c34017b5eff00" + // 4 shares
		"4000003511dcb7b9" + // RESPOND_ASSETS
		"2fc8a29a7a4a6969cd3a57244c48c5027b5b6940ed11f739d052b40e9dd357fa020001000830bb00aa7747660700000077e84f012fc8a200" + // 31780730794 shares / tick number 22014071
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
