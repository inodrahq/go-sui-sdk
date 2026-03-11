package tx

import (
	"github.com/inodrahq/go-bcs"
	"github.com/inodrahq/go-sui-sdk/txn"
)

// mustEncode encodes a BCS value into an encoder, panicking on error.
// The encoder is backed by bytes.Buffer which never returns write errors,
// so this is safe to use for in-memory encoding.
func mustEncode(e *bcs.Encoder, m bcs.Marshaler) {
	if err := m.MarshalBCS(e); err != nil {
		panic("tx: bcs encode: " + err.Error())
	}
}

// PureU8 creates a Pure CallArg for a u8 value.
func PureU8(v uint8) txn.CallArg {
	e := bcs.NewEncoder()
	mustEncode(e, bcs.U8(v))
	return txn.PureCallArg(e.Bytes())
}

// PureU16 creates a Pure CallArg for a u16 value.
func PureU16(v uint16) txn.CallArg {
	e := bcs.NewEncoder()
	mustEncode(e, bcs.U16(v))
	return txn.PureCallArg(e.Bytes())
}

// PureU32 creates a Pure CallArg for a u32 value.
func PureU32(v uint32) txn.CallArg {
	e := bcs.NewEncoder()
	mustEncode(e, bcs.U32(v))
	return txn.PureCallArg(e.Bytes())
}

// PureU64 creates a Pure CallArg for a u64 value.
func PureU64(v uint64) txn.CallArg {
	e := bcs.NewEncoder()
	mustEncode(e, bcs.U64(v))
	return txn.PureCallArg(e.Bytes())
}

// PureBool creates a Pure CallArg for a bool value.
func PureBool(v bool) txn.CallArg {
	e := bcs.NewEncoder()
	mustEncode(e, bcs.Bool(v))
	return txn.PureCallArg(e.Bytes())
}

// PureString creates a Pure CallArg for a BCS string value.
func PureString(v string) txn.CallArg {
	e := bcs.NewEncoder()
	mustEncode(e, bcs.String(v))
	return txn.PureCallArg(e.Bytes())
}

// PureAddress creates a Pure CallArg for a Sui address.
func PureAddress(addr txn.SuiAddress) txn.CallArg {
	e := bcs.NewEncoder()
	mustEncode(e, &addr)
	return txn.PureCallArg(e.Bytes())
}

// PureBytes creates a Pure CallArg for raw bytes (BCS Vec<u8>).
func PureBytes(b []byte) txn.CallArg {
	e := bcs.NewEncoder()
	if err := e.WriteULEB128(uint32(len(b))); err != nil {
		panic("tx: bcs encode: " + err.Error())
	}
	if err := e.WriteBytes(b); err != nil {
		panic("tx: bcs encode: " + err.Error())
	}
	return txn.PureCallArg(e.Bytes())
}
