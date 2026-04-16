package live

import (
	"context"
	"fmt"
	"log"

	qubic "github.com/qubic/go-node-connector"
	"github.com/qubic/go-node-connector/types"
)

type ClientPool interface {
	Get() (*qubic.Client, error)
	Close(*qubic.Client) error
	Put(*qubic.Client) error
}

type AssetFetcher interface {
	GetAssetOwnerships(ctx context.Context, identity, name string) (*types.AssetOwnerships, error)
}

type AssetClient struct {
	qPool ClientPool
}

func NewAssetClient(qPool ClientPool) *AssetClient {
	return &AssetClient{qPool: qPool}
}

func (f *AssetClient) GetAssetOwnerships(ctx context.Context, identity, name string) (*types.AssetOwnerships, error) {
	client, err := f.qPool.Get()
	if err != nil {
		return nil, fmt.Errorf("getting pool connection: %w", err)
	}
	assets, err := client.GetAssetOwnershipsByFilter(ctx, identity, name, "", 0)
	if err != nil {
		_ = f.qPool.Close(client)
		return nil, fmt.Errorf("getting asset ownerships: %w", err)
	}
	err = f.qPool.Put(client)
	if err != nil {
		log.Printf("WARN: error returning client to pool: %v", err)
	}
	return &assets, nil
}
