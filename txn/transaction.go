package txn

import (
	"fmt"

	"github.com/inodrahq/go-bcs"
)

// TransactionData is the top-level BCS enum: V1(0)
type TransactionData struct {
	V1 *TransactionDataV1
}

func (t TransactionData) MarshalBCS(e *bcs.Encoder) error {
	// Always V1 (tag 0)
	if err := e.WriteULEB128(0); err != nil {
		return err
	}
	return t.V1.MarshalBCS(e)
}

func (t *TransactionData) UnmarshalBCS(d *bcs.Decoder) error {
	tag, err := d.ReadULEB128()
	if err != nil {
		return err
	}
	if tag != 0 {
		return fmt.Errorf("unknown TransactionData version: %d", tag)
	}
	t.V1 = &TransactionDataV1{}
	return t.V1.UnmarshalBCS(d)
}

// TransactionDataV1 contains the transaction kind, sender, gas data, and expiration.
type TransactionDataV1 struct {
	Kind       TransactionKind
	Sender     SuiAddress
	GasData    GasData
	Expiration TransactionExpiration
}

func (t TransactionDataV1) MarshalBCS(e *bcs.Encoder) error {
	if err := t.Kind.MarshalBCS(e); err != nil {
		return err
	}
	if err := t.Sender.MarshalBCS(e); err != nil {
		return err
	}
	if err := t.GasData.MarshalBCS(e); err != nil {
		return err
	}
	return t.Expiration.MarshalBCS(e)
}

func (t *TransactionDataV1) UnmarshalBCS(d *bcs.Decoder) error {
	if err := t.Kind.UnmarshalBCS(d); err != nil {
		return err
	}
	if err := t.Sender.UnmarshalBCS(d); err != nil {
		return err
	}
	if err := t.GasData.UnmarshalBCS(d); err != nil {
		return err
	}
	return t.Expiration.UnmarshalBCS(d)
}

// TransactionKind is a BCS enum: ProgrammableTransaction(0)
type TransactionKind struct {
	ProgrammableTx *ProgrammableTransaction
}

func (k TransactionKind) MarshalBCS(e *bcs.Encoder) error {
	if err := e.WriteULEB128(0); err != nil {
		return err
	}
	return k.ProgrammableTx.MarshalBCS(e)
}

func (k *TransactionKind) UnmarshalBCS(d *bcs.Decoder) error {
	tag, err := d.ReadULEB128()
	if err != nil {
		return err
	}
	if tag != 0 {
		return fmt.Errorf("unknown TransactionKind: %d", tag)
	}
	k.ProgrammableTx = &ProgrammableTransaction{}
	return k.ProgrammableTx.UnmarshalBCS(d)
}

// TransactionExpiration is a BCS enum: None(0), Epoch(1)
type TransactionExpiration struct {
	Epoch *bcs.U64 // nil = None
}

func (te TransactionExpiration) MarshalBCS(e *bcs.Encoder) error {
	if te.Epoch == nil {
		return e.WriteULEB128(0)
	}
	if err := e.WriteULEB128(1); err != nil {
		return err
	}
	return te.Epoch.MarshalBCS(e)
}

func (te *TransactionExpiration) UnmarshalBCS(d *bcs.Decoder) error {
	tag, err := d.ReadULEB128()
	if err != nil {
		return err
	}
	switch tag {
	case 0:
		te.Epoch = nil
	case 1:
		te.Epoch = new(bcs.U64)
		return te.Epoch.UnmarshalBCS(d)
	default:
		return fmt.Errorf("unknown TransactionExpiration: %d", tag)
	}
	return nil
}
