package txn

import (
	"fmt"

	"github.com/inodrahq/go-bcs"
)

// Argument represents a transaction argument (BCS enum).
// Variants: GasCoin(0), Input(1), Result(2), NestedResult(3)
type Argument struct {
	variant argumentVariant
}

type argumentVariant interface {
	bcs.Marshaler
	bcs.Unmarshaler
	argumentTag() uint32
}

// GasCoin returns an Argument referencing the gas coin.
func GasCoin() Argument {
	return Argument{variant: &gasCoinVariant{}}
}

// Input returns an Argument referencing a call input by index.
func Input(index uint16) Argument {
	return Argument{variant: &inputVariant{Index: bcs.U16(index)}}
}

// Result returns an Argument referencing a command result by index.
func Result(index uint16) Argument {
	return Argument{variant: &resultVariant{Index: bcs.U16(index)}}
}

// NestedResult returns an Argument referencing a nested result.
func NestedResult(cmdIndex, resultIndex uint16) Argument {
	return Argument{variant: &nestedResultVariant{
		CmdIndex:    bcs.U16(cmdIndex),
		ResultIndex: bcs.U16(resultIndex),
	}}
}

func (a Argument) MarshalBCS(e *bcs.Encoder) error {
	if err := e.WriteULEB128(a.variant.argumentTag()); err != nil {
		return err
	}
	return a.variant.MarshalBCS(e)
}

func (a *Argument) UnmarshalBCS(d *bcs.Decoder) error {
	tag, err := d.ReadULEB128()
	if err != nil {
		return err
	}
	switch tag {
	case 0:
		a.variant = &gasCoinVariant{}
	case 1:
		v := &inputVariant{}
		if err := v.UnmarshalBCS(d); err != nil {
			return err
		}
		a.variant = v
	case 2:
		v := &resultVariant{}
		if err := v.UnmarshalBCS(d); err != nil {
			return err
		}
		a.variant = v
	case 3:
		v := &nestedResultVariant{}
		if err := v.UnmarshalBCS(d); err != nil {
			return err
		}
		a.variant = v
	default:
		return fmt.Errorf("unknown Argument tag: %d", tag)
	}
	return nil
}

// --- Variants ---

type gasCoinVariant struct{}

func (*gasCoinVariant) argumentTag() uint32          { return 0 }
func (*gasCoinVariant) MarshalBCS(e *bcs.Encoder) error  { return nil }
func (*gasCoinVariant) UnmarshalBCS(d *bcs.Decoder) error { return nil }

type inputVariant struct{ Index bcs.U16 }

func (*inputVariant) argumentTag() uint32            { return 1 }
func (v *inputVariant) MarshalBCS(e *bcs.Encoder) error  { return v.Index.MarshalBCS(e) }
func (v *inputVariant) UnmarshalBCS(d *bcs.Decoder) error { return v.Index.UnmarshalBCS(d) }

type resultVariant struct{ Index bcs.U16 }

func (*resultVariant) argumentTag() uint32           { return 2 }
func (v *resultVariant) MarshalBCS(e *bcs.Encoder) error  { return v.Index.MarshalBCS(e) }
func (v *resultVariant) UnmarshalBCS(d *bcs.Decoder) error { return v.Index.UnmarshalBCS(d) }

type nestedResultVariant struct {
	CmdIndex    bcs.U16
	ResultIndex bcs.U16
}

func (*nestedResultVariant) argumentTag() uint32 { return 3 }
func (v *nestedResultVariant) MarshalBCS(e *bcs.Encoder) error {
	if err := v.CmdIndex.MarshalBCS(e); err != nil {
		return err
	}
	return v.ResultIndex.MarshalBCS(e)
}
func (v *nestedResultVariant) UnmarshalBCS(d *bcs.Decoder) error {
	if err := v.CmdIndex.UnmarshalBCS(d); err != nil {
		return err
	}
	return v.ResultIndex.UnmarshalBCS(d)
}
