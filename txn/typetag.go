package txn

import (
	"fmt"

	"github.com/inodrahq/go-bcs"
)

// TypeTag represents a Move type tag (BCS enum).
// bool=0, u8=1, u64=2, u128=3, address=4, signer=5, vector=6, struct=7, u16=8, u32=9, u256=10
type TypeTag struct {
	variant typeTagVariant
}

type typeTagVariant interface {
	bcs.Marshaler
	bcs.Unmarshaler
	typeTagTag() uint32
}

// Commonly used TypeTag constructors
func TypeTagBool() TypeTag     { return TypeTag{variant: &simpleTypeTag{tag: 0}} }
func TypeTagU8() TypeTag       { return TypeTag{variant: &simpleTypeTag{tag: 1}} }
func TypeTagU64() TypeTag      { return TypeTag{variant: &simpleTypeTag{tag: 2}} }
func TypeTagU128() TypeTag     { return TypeTag{variant: &simpleTypeTag{tag: 3}} }
func TypeTagAddress() TypeTag  { return TypeTag{variant: &simpleTypeTag{tag: 4}} }
func TypeTagSigner() TypeTag   { return TypeTag{variant: &simpleTypeTag{tag: 5}} }
func TypeTagU16() TypeTag      { return TypeTag{variant: &simpleTypeTag{tag: 8}} }
func TypeTagU32() TypeTag      { return TypeTag{variant: &simpleTypeTag{tag: 9}} }
func TypeTagU256() TypeTag     { return TypeTag{variant: &simpleTypeTag{tag: 10}} }

func TypeTagVector(inner TypeTag) TypeTag {
	return TypeTag{variant: &vectorTypeTag{Inner: inner}}
}

func TypeTagStruct(st StructTag) TypeTag {
	return TypeTag{variant: &structTypeTag{Tag: st}}
}

func (t TypeTag) MarshalBCS(e *bcs.Encoder) error {
	if err := e.WriteULEB128(t.variant.typeTagTag()); err != nil {
		return err
	}
	return t.variant.MarshalBCS(e)
}

func (t *TypeTag) UnmarshalBCS(d *bcs.Decoder) error {
	tag, err := d.ReadULEB128()
	if err != nil {
		return err
	}
	switch tag {
	case 0, 1, 2, 3, 4, 5, 8, 9, 10:
		t.variant = &simpleTypeTag{tag: tag}
	case 6:
		v := &vectorTypeTag{}
		if err := v.UnmarshalBCS(d); err != nil {
			return err
		}
		t.variant = v
	case 7:
		v := &structTypeTag{}
		if err := v.UnmarshalBCS(d); err != nil {
			return err
		}
		t.variant = v
	default:
		return fmt.Errorf("unknown TypeTag tag: %d", tag)
	}
	return nil
}

// simpleTypeTag for types with no data (bool, u8, etc.)
type simpleTypeTag struct{ tag uint32 }

func (s *simpleTypeTag) typeTagTag() uint32          { return s.tag }
func (s *simpleTypeTag) MarshalBCS(e *bcs.Encoder) error  { return nil }
func (s *simpleTypeTag) UnmarshalBCS(d *bcs.Decoder) error { return nil }

// vectorTypeTag wraps another TypeTag
type vectorTypeTag struct{ Inner TypeTag }

func (*vectorTypeTag) typeTagTag() uint32            { return 6 }
func (v *vectorTypeTag) MarshalBCS(e *bcs.Encoder) error  { return v.Inner.MarshalBCS(e) }
func (v *vectorTypeTag) UnmarshalBCS(d *bcs.Decoder) error { return v.Inner.UnmarshalBCS(d) }

// structTypeTag wraps a StructTag
type structTypeTag struct{ Tag StructTag }

func (*structTypeTag) typeTagTag() uint32            { return 7 }
func (v *structTypeTag) MarshalBCS(e *bcs.Encoder) error  { return v.Tag.MarshalBCS(e) }
func (v *structTypeTag) UnmarshalBCS(d *bcs.Decoder) error { return v.Tag.UnmarshalBCS(d) }

// StructTag identifies a Move struct type.
type StructTag struct {
	Address    SuiAddress
	Module     bcs.String
	Name       bcs.String
	TypeParams []TypeTag
}

func (s StructTag) MarshalBCS(e *bcs.Encoder) error {
	if err := s.Address.MarshalBCS(e); err != nil {
		return err
	}
	if err := s.Module.MarshalBCS(e); err != nil {
		return err
	}
	if err := s.Name.MarshalBCS(e); err != nil {
		return err
	}
	if err := e.WriteULEB128(uint32(len(s.TypeParams))); err != nil {
		return err
	}
	for i := range s.TypeParams {
		if err := s.TypeParams[i].MarshalBCS(e); err != nil {
			return err
		}
	}
	return nil
}

func (s *StructTag) UnmarshalBCS(d *bcs.Decoder) error {
	if err := s.Address.UnmarshalBCS(d); err != nil {
		return err
	}
	if err := s.Module.UnmarshalBCS(d); err != nil {
		return err
	}
	if err := s.Name.UnmarshalBCS(d); err != nil {
		return err
	}
	n, err := d.ReadULEB128()
	if err != nil {
		return err
	}
	s.TypeParams = make([]TypeTag, n)
	for i := uint32(0); i < n; i++ {
		if err := s.TypeParams[i].UnmarshalBCS(d); err != nil {
			return err
		}
	}
	return nil
}
