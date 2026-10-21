package spectrum

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
)

type Entity struct {
	publicKey [32]byte

	incomingAmount int64
	outgoingAmount int64

	incomingTransfers uint32
	outgoingTransfers uint32

	latestIncomingTick uint32
	latestOutgoingTick uint32
}

const entitySize = 64 // 32 + 8 + 8 + 4 + 4 + 4 + 4

func (e *Entity) GetPublicKeyString() string {
	return hex.EncodeToString(e.publicKey[:])
}

func (e *Entity) GetBalance() (int64, error) {
	balance := e.incomingAmount - e.outgoingAmount
	if balance < 0 {
		return 0, fmt.Errorf("Negative balance (%d) for %s", balance, e.GetPublicKeyString())
	}

	return balance, nil
}

func (e *Entity) UnmarshallFromBinary(r io.Reader) error {

	byteOrder := binary.LittleEndian

	err := binary.Read(r, byteOrder, &e.publicKey)
	if err != nil {
		return fmt.Errorf("reading public key: %w", err)
	}

	err = binary.Read(r, byteOrder, &e.incomingAmount)
	if err != nil {
		return fmt.Errorf("reading incoming amount: %w", err)
	}
	err = binary.Read(r, byteOrder, &e.outgoingAmount)
	if err != nil {
		return fmt.Errorf("reading outgoing amount: %w", err)
	}

	err = binary.Read(r, byteOrder, &e.incomingTransfers)
	if err != nil {
		return fmt.Errorf("reading incoming transfers: %w", err)
	}
	err = binary.Read(r, byteOrder, &e.outgoingTransfers)
	if err != nil {
		return fmt.Errorf("reading outgoing transfers: %w", err)
	}

	err = binary.Read(r, byteOrder, &e.latestIncomingTick)
	if err != nil {
		return fmt.Errorf("reading latest incoming tick: %w", err)
	}
	err = binary.Read(r, byteOrder, &e.latestOutgoingTick)
	if err != nil {
		return fmt.Errorf("reading latest outgoing tick: %w", err)
	}

	return nil
}

type RichListEntity struct {
	// Rank is the zero based position in the rich list, highest balance first. It is persisted so
	// that a page can be served with an indexed range query instead of sorting the whole collection.
	Rank     int32  `bson:"rank"`
	Identity string `bson:"identity"`
	Balance  int64  `bson:"balance"`
}

type RichList []RichListEntity
