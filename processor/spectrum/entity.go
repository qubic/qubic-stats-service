package spectrum

import (
	"encoding/binary"
	"encoding/hex"
	"github.com/pkg/errors"
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

func (e *Entity) GetPublicKeyString() string {
	return hex.EncodeToString(e.publicKey[:])
}

func (e *Entity) GetBalance() (int64, error) {
	balance := e.incomingAmount - e.outgoingAmount
	if balance < 0 {
		return 0, errors.Errorf("Negative balance (%d) for %s", balance, e.GetPublicKeyString())
	}

	return balance, nil
}

func (e *Entity) UnmarshallFromBinary(r io.Reader) error {

	byteOrder := binary.LittleEndian

	err := binary.Read(r, byteOrder, &e.publicKey)
	if err != nil {
		return errors.Wrap(err, "reading public key")
	}

	err = binary.Read(r, byteOrder, &e.incomingAmount)
	if err != nil {
		return errors.Wrap(err, "reading incoming amount")
	}
	err = binary.Read(r, byteOrder, &e.outgoingAmount)
	if err != nil {
		return errors.Wrap(err, "reading outgoing amount")
	}

	err = binary.Read(r, byteOrder, &e.incomingTransfers)
	if err != nil {
		return errors.Wrap(err, "reading incoming transfers")
	}
	err = binary.Read(r, byteOrder, &e.outgoingTransfers)
	if err != nil {
		return errors.Wrap(err, "reading outgoing transfers")
	}

	err = binary.Read(r, byteOrder, &e.latestIncomingTick)
	if err != nil {
		return errors.Wrap(err, "reading latest incoming tick")
	}
	err = binary.Read(r, byteOrder, &e.latestOutgoingTick)
	if err != nil {
		return errors.Wrap(err, "reading latest outgoing tick")
	}

	return nil
}

type RichListEntity struct {
	Identity string `bson:"identity"`
	Balance  int64  `bson:"balance"`
}

type RichList []RichListEntity
