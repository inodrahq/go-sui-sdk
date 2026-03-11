package txn

import (
	"fmt"

	"github.com/inodrahq/go-bcs"
)

// CallArg represents a transaction call argument (BCS enum).
// Variants: Pure(0), Object(1)
type CallArg struct {
	variant callArgVariant
}

type callArgVariant interface {
	bcs.Marshaler
	bcs.Unmarshaler
	callArgTag() uint32
}

// PureCallArg creates a CallArg with pure BCS-encoded bytes.
func PureCallArg(data []byte) CallArg {
	return CallArg{variant: &pureCallArgVariant{Data: bcs.ByteVector(data)}}
}

// ObjectCallArg creates a CallArg referencing an object.
func ObjectCallArg(obj ObjectArg) CallArg {
	return CallArg{variant: &objectCallArgVariant{Arg: obj}}
}

func (c CallArg) MarshalBCS(e *bcs.Encoder) error {
	if err := e.WriteULEB128(c.variant.callArgTag()); err != nil {
		return err
	}
	return c.variant.MarshalBCS(e)
}

func (c *CallArg) UnmarshalBCS(d *bcs.Decoder) error {
	tag, err := d.ReadULEB128()
	if err != nil {
		return err
	}
	switch tag {
	case 0:
		v := &pureCallArgVariant{}
		if err := v.UnmarshalBCS(d); err != nil {
			return err
		}
		c.variant = v
	case 1:
		v := &objectCallArgVariant{}
		if err := v.UnmarshalBCS(d); err != nil {
			return err
		}
		c.variant = v
	default:
		return fmt.Errorf("unknown CallArg tag: %d", tag)
	}
	return nil
}

type pureCallArgVariant struct{ Data bcs.ByteVector }

func (*pureCallArgVariant) callArgTag() uint32            { return 0 }
func (v *pureCallArgVariant) MarshalBCS(e *bcs.Encoder) error  { return v.Data.MarshalBCS(e) }
func (v *pureCallArgVariant) UnmarshalBCS(d *bcs.Decoder) error { return v.Data.UnmarshalBCS(d) }

type objectCallArgVariant struct{ Arg ObjectArg }

func (*objectCallArgVariant) callArgTag() uint32 { return 1 }
func (v *objectCallArgVariant) MarshalBCS(e *bcs.Encoder) error {
	return v.Arg.MarshalBCS(e)
}
func (v *objectCallArgVariant) UnmarshalBCS(d *bcs.Decoder) error {
	return v.Arg.UnmarshalBCS(d)
}
