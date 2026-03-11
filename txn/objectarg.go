package txn

import (
	"fmt"

	"github.com/inodrahq/go-bcs"
)

// ObjectArg represents an object argument (BCS enum).
// Variants: ImmOrOwnedObject(0), SharedObject(1), Receiving(2)
type ObjectArg struct {
	variant objectArgVariant
}

type objectArgVariant interface {
	bcs.Marshaler
	bcs.Unmarshaler
	objectArgTag() uint32
}

// ImmOrOwnedObject creates an ObjectArg for an immutable or owned object.
func ImmOrOwnedObject(ref ObjectRef) ObjectArg {
	return ObjectArg{variant: &immOrOwnedVariant{Ref: ref}}
}

// SharedObject creates an ObjectArg for a shared object.
func SharedObject(ref SharedObjectRef) ObjectArg {
	return ObjectArg{variant: &sharedVariant{Ref: ref}}
}

// Receiving creates an ObjectArg for a receiving object.
func Receiving(ref ObjectRef) ObjectArg {
	return ObjectArg{variant: &receivingVariant{Ref: ref}}
}

func (o ObjectArg) MarshalBCS(e *bcs.Encoder) error {
	if err := e.WriteULEB128(o.variant.objectArgTag()); err != nil {
		return err
	}
	return o.variant.MarshalBCS(e)
}

func (o *ObjectArg) UnmarshalBCS(d *bcs.Decoder) error {
	tag, err := d.ReadULEB128()
	if err != nil {
		return err
	}
	switch tag {
	case 0:
		v := &immOrOwnedVariant{}
		if err := v.UnmarshalBCS(d); err != nil {
			return err
		}
		o.variant = v
	case 1:
		v := &sharedVariant{}
		if err := v.UnmarshalBCS(d); err != nil {
			return err
		}
		o.variant = v
	case 2:
		v := &receivingVariant{}
		if err := v.UnmarshalBCS(d); err != nil {
			return err
		}
		o.variant = v
	default:
		return fmt.Errorf("unknown ObjectArg tag: %d", tag)
	}
	return nil
}

type immOrOwnedVariant struct{ Ref ObjectRef }

func (*immOrOwnedVariant) objectArgTag() uint32            { return 0 }
func (v *immOrOwnedVariant) MarshalBCS(e *bcs.Encoder) error  { return v.Ref.MarshalBCS(e) }
func (v *immOrOwnedVariant) UnmarshalBCS(d *bcs.Decoder) error { return v.Ref.UnmarshalBCS(d) }

type sharedVariant struct{ Ref SharedObjectRef }

func (*sharedVariant) objectArgTag() uint32            { return 1 }
func (v *sharedVariant) MarshalBCS(e *bcs.Encoder) error  { return v.Ref.MarshalBCS(e) }
func (v *sharedVariant) UnmarshalBCS(d *bcs.Decoder) error { return v.Ref.UnmarshalBCS(d) }

type receivingVariant struct{ Ref ObjectRef }

func (*receivingVariant) objectArgTag() uint32             { return 2 }
func (v *receivingVariant) MarshalBCS(e *bcs.Encoder) error  { return v.Ref.MarshalBCS(e) }
func (v *receivingVariant) UnmarshalBCS(d *bcs.Decoder) error { return v.Ref.UnmarshalBCS(d) }
