package txn

import "github.com/inodrahq/go-bcs"

// ObjectDigest is a 32-byte object digest, BCS-encoded with ULEB128 length prefix
// (matching Sui's Digest<32> with serde_bytes).
type ObjectDigest [32]byte

func (d ObjectDigest) MarshalBCS(e *bcs.Encoder) error {
	if err := e.WriteULEB128(32); err != nil {
		return err
	}
	return e.WriteBytes(d[:])
}

func (d *ObjectDigest) UnmarshalBCS(dec *bcs.Decoder) error {
	length, err := dec.ReadULEB128()
	if err != nil {
		return err
	}
	if length != 32 {
		return bcs.ErrLengthMismatch
	}
	b, err := dec.ReadBytes(32)
	if err != nil {
		return err
	}
	copy(d[:], b)
	return nil
}

// ObjectRef is a reference to a Sui object (ID, version, digest).
type ObjectRef struct {
	ObjectID SuiAddress
	Version  bcs.U64
	Digest   ObjectDigest
}

func (r ObjectRef) MarshalBCS(e *bcs.Encoder) error {
	return bcs.EncodeStruct(e, r)
}

func (r *ObjectRef) UnmarshalBCS(d *bcs.Decoder) error {
	return bcs.DecodeStruct(d, r)
}

// SharedObjectRef references a shared object by ID and initial shared version.
type SharedObjectRef struct {
	ObjectID                SuiAddress
	InitialSharedVersion   bcs.U64
	Mutable                bcs.Bool
}

func (r SharedObjectRef) MarshalBCS(e *bcs.Encoder) error {
	return bcs.EncodeStruct(e, r)
}

func (r *SharedObjectRef) UnmarshalBCS(d *bcs.Decoder) error {
	return bcs.DecodeStruct(d, r)
}
