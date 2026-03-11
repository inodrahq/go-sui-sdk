package txn_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/inodrahq/go-bcs"
	"github.com/inodrahq/go-sui-sdk/txn"
)

// helper: encode a value, decode into dst, re-encode and compare bytes.
func roundTrip(t *testing.T, name string, encode func(*bcs.Encoder) error, decode func(*bcs.Decoder) error, reEncode func(*bcs.Encoder) error) {
	t.Helper()

	e1 := bcs.NewEncoder()
	if err := encode(e1); err != nil {
		t.Fatalf("%s: encode error: %v", name, err)
	}
	b1 := e1.Bytes()

	d := bcs.NewDecoder(b1)
	if err := decode(d); err != nil {
		t.Fatalf("%s: decode error: %v", name, err)
	}
	if d.Remaining() != 0 {
		t.Fatalf("%s: %d trailing bytes after decode", name, d.Remaining())
	}

	e2 := bcs.NewEncoder()
	if err := reEncode(e2); err != nil {
		t.Fatalf("%s: re-encode error: %v", name, err)
	}
	b2 := e2.Bytes()

	if !bytes.Equal(b1, b2) {
		t.Errorf("%s: round-trip mismatch\n  original:  %s\n  re-encoded:%s", name, hex.EncodeToString(b1), hex.EncodeToString(b2))
	}
}

// --- 1. SuiAddress ---

func TestRoundTrip_SuiAddress(t *testing.T) {
	original, err := txn.ParseAddress("0x90f3e6d73b5730f16974f4df1d3441394ebae62186baf83608599f226455afa7")
	if err != nil {
		t.Fatal(err)
	}

	var decoded txn.SuiAddress
	roundTrip(t, "SuiAddress",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)

	if decoded.Hex() != original.Hex() {
		t.Errorf("SuiAddress Hex mismatch: got %s, want %s", decoded.Hex(), original.Hex())
	}
}

func TestRoundTrip_SuiAddress_Zero(t *testing.T) {
	original, err := txn.ParseAddress("0x0")
	if err != nil {
		t.Fatal(err)
	}

	var decoded txn.SuiAddress
	roundTrip(t, "SuiAddress_Zero",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)
}

func TestRoundTrip_SuiAddress_Short(t *testing.T) {
	// Short address with zero-padding
	original, err := txn.ParseAddress("0x1")
	if err != nil {
		t.Fatal(err)
	}

	var decoded txn.SuiAddress
	roundTrip(t, "SuiAddress_Short",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)

	expected := "0x0000000000000000000000000000000000000000000000000000000000000001"
	if decoded.Hex() != expected {
		t.Errorf("SuiAddress_Short Hex: got %s, want %s", decoded.Hex(), expected)
	}
}

// --- 2. ObjectDigest ---

func TestRoundTrip_ObjectDigest(t *testing.T) {
	b, _ := hex.DecodeString("aabbccddaabbccddaabbccddaabbccddaabbccddaabbccddaabbccddaabbccdd")
	var original txn.ObjectDigest
	copy(original[:], b)

	var decoded txn.ObjectDigest
	roundTrip(t, "ObjectDigest",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)

	if decoded != original {
		t.Errorf("ObjectDigest mismatch")
	}
}

func TestRoundTrip_ObjectDigest_Zero(t *testing.T) {
	var original txn.ObjectDigest // all zeros

	var decoded txn.ObjectDigest
	roundTrip(t, "ObjectDigest_Zero",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)

	if decoded != original {
		t.Errorf("ObjectDigest_Zero mismatch")
	}
}

// --- 3. ObjectRef ---

func TestRoundTrip_ObjectRef(t *testing.T) {
	original := txn.ObjectRef{
		ObjectID: mustAddr(t, "1111111111111111111111111111111111111111111111111111111111111111"),
		Version:  42,
		Digest:   mustDigest(t, "2222222222222222222222222222222222222222222222222222222222222222"),
	}

	var decoded txn.ObjectRef
	roundTrip(t, "ObjectRef",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)

	if decoded.ObjectID != original.ObjectID {
		t.Errorf("ObjectRef ObjectID mismatch")
	}
	if decoded.Version != original.Version {
		t.Errorf("ObjectRef Version: got %d, want %d", decoded.Version, original.Version)
	}
	if decoded.Digest != original.Digest {
		t.Errorf("ObjectRef Digest mismatch")
	}
}

// --- 4. SharedObjectRef ---

func TestRoundTrip_SharedObjectRef(t *testing.T) {
	original := txn.SharedObjectRef{
		ObjectID:             mustAddr(t, "7777777777777777777777777777777777777777777777777777777777777777"),
		InitialSharedVersion: 100,
		Mutable:              true,
	}

	var decoded txn.SharedObjectRef
	roundTrip(t, "SharedObjectRef",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)

	if decoded.ObjectID != original.ObjectID {
		t.Errorf("SharedObjectRef ObjectID mismatch")
	}
	if decoded.InitialSharedVersion != original.InitialSharedVersion {
		t.Errorf("SharedObjectRef InitialSharedVersion: got %d, want %d", decoded.InitialSharedVersion, original.InitialSharedVersion)
	}
	if decoded.Mutable != original.Mutable {
		t.Errorf("SharedObjectRef Mutable: got %v, want %v", decoded.Mutable, original.Mutable)
	}
}

func TestRoundTrip_SharedObjectRef_Immutable(t *testing.T) {
	original := txn.SharedObjectRef{
		ObjectID:             mustAddr(t, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
		InitialSharedVersion: 1,
		Mutable:              false,
	}

	var decoded txn.SharedObjectRef
	roundTrip(t, "SharedObjectRef_Immutable",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)
}

// --- 5. Argument variants ---

func TestRoundTrip_Argument_GasCoin(t *testing.T) {
	original := txn.GasCoin()

	var decoded txn.Argument
	roundTrip(t, "Argument_GasCoin",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)
}

func TestRoundTrip_Argument_Input(t *testing.T) {
	original := txn.Input(7)

	var decoded txn.Argument
	roundTrip(t, "Argument_Input",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)
}

func TestRoundTrip_Argument_Result(t *testing.T) {
	original := txn.Result(3)

	var decoded txn.Argument
	roundTrip(t, "Argument_Result",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)
}

func TestRoundTrip_Argument_NestedResult(t *testing.T) {
	original := txn.NestedResult(5, 2)

	var decoded txn.Argument
	roundTrip(t, "Argument_NestedResult",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)
}

// --- 6. CallArg variants ---

func TestRoundTrip_CallArg_Pure(t *testing.T) {
	original := txn.PureCallArg([]byte{0x01, 0x02, 0x03, 0x04})

	var decoded txn.CallArg
	roundTrip(t, "CallArg_Pure",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)
}

func TestRoundTrip_CallArg_Pure_Empty(t *testing.T) {
	original := txn.PureCallArg([]byte{})

	var decoded txn.CallArg
	roundTrip(t, "CallArg_Pure_Empty",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)
}

func TestRoundTrip_CallArg_Object(t *testing.T) {
	ref := txn.ObjectRef{
		ObjectID: mustAddr(t, "1111111111111111111111111111111111111111111111111111111111111111"),
		Version:  10,
		Digest:   mustDigest(t, "2222222222222222222222222222222222222222222222222222222222222222"),
	}
	original := txn.ObjectCallArg(txn.ImmOrOwnedObject(ref))

	var decoded txn.CallArg
	roundTrip(t, "CallArg_Object",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)
}

// --- 7. ObjectArg variants ---

func TestRoundTrip_ObjectArg_ImmOrOwnedObject(t *testing.T) {
	ref := txn.ObjectRef{
		ObjectID: mustAddr(t, "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"),
		Version:  99,
		Digest:   mustDigest(t, "fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210"),
	}
	original := txn.ImmOrOwnedObject(ref)

	var decoded txn.ObjectArg
	roundTrip(t, "ObjectArg_ImmOrOwnedObject",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)
}

func TestRoundTrip_ObjectArg_SharedObject(t *testing.T) {
	ref := txn.SharedObjectRef{
		ObjectID:             mustAddr(t, "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"),
		InitialSharedVersion: 50,
		Mutable:              true,
	}
	original := txn.SharedObject(ref)

	var decoded txn.ObjectArg
	roundTrip(t, "ObjectArg_SharedObject",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)
}

func TestRoundTrip_ObjectArg_Receiving(t *testing.T) {
	ref := txn.ObjectRef{
		ObjectID: mustAddr(t, "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"),
		Version:  77,
		Digest:   mustDigest(t, "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"),
	}
	original := txn.Receiving(ref)

	var decoded txn.ObjectArg
	roundTrip(t, "ObjectArg_Receiving",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)
}

// --- 8. Command variants ---

func TestRoundTrip_Command_MoveCall(t *testing.T) {
	original := txn.MoveCallCommand(txn.MoveCall{
		Package:  mustAddr(t, "0000000000000000000000000000000000000000000000000000000000000002"),
		Module:   "pay",
		Function: "split",
		TypeArgs: []txn.TypeTag{
			txn.TypeTagStruct(txn.StructTag{
				Address:    mustAddr(t, "0000000000000000000000000000000000000000000000000000000000000002"),
				Module:     "sui",
				Name:       "SUI",
				TypeParams: []txn.TypeTag{},
			}),
		},
		Args: []txn.Argument{txn.Input(0), txn.Input(1)},
	})

	var decoded txn.Command
	roundTrip(t, "Command_MoveCall",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)
}

func TestRoundTrip_Command_MoveCall_NoTypeArgs(t *testing.T) {
	original := txn.MoveCallCommand(txn.MoveCall{
		Package:  mustAddr(t, "8888888888888888888888888888888888888888888888888888888888888888"),
		Module:   "counter",
		Function: "increment",
		TypeArgs: []txn.TypeTag{},
		Args:     []txn.Argument{txn.Input(0)},
	})

	var decoded txn.Command
	roundTrip(t, "Command_MoveCall_NoTypeArgs",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)
}

func TestRoundTrip_Command_TransferObjects(t *testing.T) {
	original := txn.TransferObjectsCommand(
		[]txn.Argument{txn.Result(0), txn.Result(1)},
		txn.Input(0),
	)

	var decoded txn.Command
	roundTrip(t, "Command_TransferObjects",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)
}

func TestRoundTrip_Command_SplitCoins(t *testing.T) {
	original := txn.SplitCoinsCommand(
		txn.GasCoin(),
		[]txn.Argument{txn.Input(0), txn.Input(1)},
	)

	var decoded txn.Command
	roundTrip(t, "Command_SplitCoins",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)
}

func TestRoundTrip_Command_MergeCoins(t *testing.T) {
	original := txn.MergeCoinsCommand(
		txn.Input(0),
		[]txn.Argument{txn.Input(1), txn.Input(2)},
	)

	var decoded txn.Command
	roundTrip(t, "Command_MergeCoins",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)
}

// --- 9. TypeTag variants ---

func TestRoundTrip_TypeTag_Bool(t *testing.T) {
	original := txn.TypeTagBool()

	var decoded txn.TypeTag
	roundTrip(t, "TypeTag_Bool",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)
}

func TestRoundTrip_TypeTag_U8(t *testing.T) {
	original := txn.TypeTagU8()

	var decoded txn.TypeTag
	roundTrip(t, "TypeTag_U8",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)
}

func TestRoundTrip_TypeTag_U64(t *testing.T) {
	original := txn.TypeTagU64()

	var decoded txn.TypeTag
	roundTrip(t, "TypeTag_U64",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)
}

func TestRoundTrip_TypeTag_U128(t *testing.T) {
	original := txn.TypeTagU128()

	var decoded txn.TypeTag
	roundTrip(t, "TypeTag_U128",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)
}

func TestRoundTrip_TypeTag_Address(t *testing.T) {
	original := txn.TypeTagAddress()

	var decoded txn.TypeTag
	roundTrip(t, "TypeTag_Address",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)
}

func TestRoundTrip_TypeTag_Signer(t *testing.T) {
	original := txn.TypeTagSigner()

	var decoded txn.TypeTag
	roundTrip(t, "TypeTag_Signer",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)
}

func TestRoundTrip_TypeTag_U16(t *testing.T) {
	original := txn.TypeTagU16()

	var decoded txn.TypeTag
	roundTrip(t, "TypeTag_U16",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)
}

func TestRoundTrip_TypeTag_U32(t *testing.T) {
	original := txn.TypeTagU32()

	var decoded txn.TypeTag
	roundTrip(t, "TypeTag_U32",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)
}

func TestRoundTrip_TypeTag_U256(t *testing.T) {
	original := txn.TypeTagU256()

	var decoded txn.TypeTag
	roundTrip(t, "TypeTag_U256",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)
}

func TestRoundTrip_TypeTag_Vector(t *testing.T) {
	original := txn.TypeTagVector(txn.TypeTagU8())

	var decoded txn.TypeTag
	roundTrip(t, "TypeTag_Vector",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)
}

func TestRoundTrip_TypeTag_VectorNested(t *testing.T) {
	original := txn.TypeTagVector(txn.TypeTagVector(txn.TypeTagU64()))

	var decoded txn.TypeTag
	roundTrip(t, "TypeTag_VectorNested",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)
}

func TestRoundTrip_TypeTag_Struct(t *testing.T) {
	original := txn.TypeTagStruct(txn.StructTag{
		Address:    mustAddr(t, "0000000000000000000000000000000000000000000000000000000000000002"),
		Module:     "sui",
		Name:       "SUI",
		TypeParams: []txn.TypeTag{},
	})

	var decoded txn.TypeTag
	roundTrip(t, "TypeTag_Struct",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)
}

func TestRoundTrip_TypeTag_StructWithTypeParams(t *testing.T) {
	original := txn.TypeTagStruct(txn.StructTag{
		Address: mustAddr(t, "0000000000000000000000000000000000000000000000000000000000000002"),
		Module:  "coin",
		Name:    "Coin",
		TypeParams: []txn.TypeTag{
			txn.TypeTagStruct(txn.StructTag{
				Address:    mustAddr(t, "0000000000000000000000000000000000000000000000000000000000000002"),
				Module:     "sui",
				Name:       "SUI",
				TypeParams: []txn.TypeTag{},
			}),
		},
	})

	var decoded txn.TypeTag
	roundTrip(t, "TypeTag_StructWithTypeParams",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)
}

// --- 10. GasData ---

func TestRoundTrip_GasData(t *testing.T) {
	original := txn.GasData{
		Payment: []txn.ObjectRef{
			{
				ObjectID: mustAddr(t, "0000000000000000000000000000000000000000000000000000000000000005"),
				Version:  1,
				Digest:   mustDigest(t, "0000000000000000000000000000000000000000000000000000000000000001"),
			},
		},
		Owner:  mustAddr(t, "90f3e6d73b5730f16974f4df1d3441394ebae62186baf83608599f226455afa7"),
		Price:  1000,
		Budget: 10000000,
	}

	var decoded txn.GasData
	roundTrip(t, "GasData",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)
}

func TestRoundTrip_GasData_MultiplePayments(t *testing.T) {
	original := txn.GasData{
		Payment: []txn.ObjectRef{
			{
				ObjectID: mustAddr(t, "0000000000000000000000000000000000000000000000000000000000000005"),
				Version:  1,
				Digest:   mustDigest(t, "0000000000000000000000000000000000000000000000000000000000000001"),
			},
			{
				ObjectID: mustAddr(t, "0000000000000000000000000000000000000000000000000000000000000006"),
				Version:  2,
				Digest:   mustDigest(t, "0000000000000000000000000000000000000000000000000000000000000002"),
			},
		},
		Owner:  mustAddr(t, "90f3e6d73b5730f16974f4df1d3441394ebae62186baf83608599f226455afa7"),
		Price:  750,
		Budget: 5000000,
	}

	var decoded txn.GasData
	roundTrip(t, "GasData_MultiplePayments",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)
}

// --- 11. TransactionExpiration ---

func TestRoundTrip_TransactionExpiration_None(t *testing.T) {
	original := txn.TransactionExpiration{}

	var decoded txn.TransactionExpiration
	roundTrip(t, "TransactionExpiration_None",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)

	if decoded.Epoch != nil {
		t.Errorf("TransactionExpiration_None: Epoch should be nil, got %v", decoded.Epoch)
	}
}

func TestRoundTrip_TransactionExpiration_Epoch(t *testing.T) {
	epoch := bcs.U64(12345)
	original := txn.TransactionExpiration{Epoch: &epoch}

	var decoded txn.TransactionExpiration
	roundTrip(t, "TransactionExpiration_Epoch",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)

	if decoded.Epoch == nil {
		t.Fatal("TransactionExpiration_Epoch: Epoch should not be nil")
	}
	if *decoded.Epoch != epoch {
		t.Errorf("TransactionExpiration_Epoch: got %d, want %d", *decoded.Epoch, epoch)
	}
}

// --- 12. TransactionKind ---

func TestRoundTrip_TransactionKind(t *testing.T) {
	pt := &txn.ProgrammableTransaction{
		Inputs: []txn.CallArg{
			txn.PureCallArg([]byte{0xCA, 0xFE}),
		},
		Commands: []txn.Command{
			txn.SplitCoinsCommand(txn.GasCoin(), []txn.Argument{txn.Input(0)}),
		},
	}
	original := txn.TransactionKind{ProgrammableTx: pt}

	var decoded txn.TransactionKind
	roundTrip(t, "TransactionKind",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)

	if decoded.ProgrammableTx == nil {
		t.Fatal("TransactionKind: ProgrammableTx should not be nil")
	}
}

// --- 13. TransactionDataV1 ---

func TestRoundTrip_TransactionDataV1(t *testing.T) {
	amountEnc := bcs.NewEncoder()
	amt := bcs.U64(1000000000)
	amt.MarshalBCS(amountEnc)

	recipientAddr := mustAddr(t, "fd233cd9a5dd7e577f16fa523427c75fbc382af1583c39fdf1c6747d2ed807a3")
	recipientEnc := bcs.NewEncoder()
	recipientAddr.MarshalBCS(recipientEnc)

	original := txn.TransactionDataV1{
		Kind: txn.TransactionKind{
			ProgrammableTx: &txn.ProgrammableTransaction{
				Inputs: []txn.CallArg{
					txn.PureCallArg(amountEnc.Bytes()),
					txn.PureCallArg(recipientEnc.Bytes()),
				},
				Commands: []txn.Command{
					txn.SplitCoinsCommand(txn.GasCoin(), []txn.Argument{txn.Input(0)}),
					txn.TransferObjectsCommand([]txn.Argument{txn.Result(0)}, txn.Input(1)),
				},
			},
		},
		Sender: mustAddr(t, "90f3e6d73b5730f16974f4df1d3441394ebae62186baf83608599f226455afa7"),
		GasData: txn.GasData{
			Payment: []txn.ObjectRef{{
				ObjectID: mustAddr(t, "0000000000000000000000000000000000000000000000000000000000000005"),
				Version:  1,
				Digest:   mustDigest(t, "0000000000000000000000000000000000000000000000000000000000000001"),
			}},
			Owner:  mustAddr(t, "90f3e6d73b5730f16974f4df1d3441394ebae62186baf83608599f226455afa7"),
			Price:  1000,
			Budget: 10000000,
		},
		Expiration: txn.TransactionExpiration{},
	}

	var decoded txn.TransactionDataV1
	roundTrip(t, "TransactionDataV1",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)
}

// --- 14. TransactionData (top level) ---

func TestRoundTrip_TransactionData(t *testing.T) {
	amountEnc := bcs.NewEncoder()
	amt := bcs.U64(1000000000)
	amt.MarshalBCS(amountEnc)

	recipientAddr := mustAddr(t, "fd233cd9a5dd7e577f16fa523427c75fbc382af1583c39fdf1c6747d2ed807a3")
	recipientEnc := bcs.NewEncoder()
	recipientAddr.MarshalBCS(recipientEnc)

	original := txn.TransactionData{
		V1: &txn.TransactionDataV1{
			Kind: txn.TransactionKind{
				ProgrammableTx: &txn.ProgrammableTransaction{
					Inputs: []txn.CallArg{
						txn.PureCallArg(amountEnc.Bytes()),
						txn.PureCallArg(recipientEnc.Bytes()),
					},
					Commands: []txn.Command{
						txn.SplitCoinsCommand(txn.GasCoin(), []txn.Argument{txn.Input(0)}),
						txn.TransferObjectsCommand([]txn.Argument{txn.Result(0)}, txn.Input(1)),
					},
				},
			},
			Sender: mustAddr(t, "90f3e6d73b5730f16974f4df1d3441394ebae62186baf83608599f226455afa7"),
			GasData: txn.GasData{
				Payment: []txn.ObjectRef{{
					ObjectID: mustAddr(t, "0000000000000000000000000000000000000000000000000000000000000005"),
					Version:  1,
					Digest:   mustDigest(t, "0000000000000000000000000000000000000000000000000000000000000001"),
				}},
				Owner:  mustAddr(t, "90f3e6d73b5730f16974f4df1d3441394ebae62186baf83608599f226455afa7"),
				Price:  1000,
				Budget: 10000000,
			},
			Expiration: txn.TransactionExpiration{},
		},
	}

	var decoded txn.TransactionData
	roundTrip(t, "TransactionData",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)
}

func TestRoundTrip_TransactionData_WithExpiration(t *testing.T) {
	epoch := bcs.U64(9999)
	original := txn.TransactionData{
		V1: &txn.TransactionDataV1{
			Kind: txn.TransactionKind{
				ProgrammableTx: &txn.ProgrammableTransaction{
					Inputs:   []txn.CallArg{txn.PureCallArg([]byte{0x01})},
					Commands: []txn.Command{txn.SplitCoinsCommand(txn.GasCoin(), []txn.Argument{txn.Input(0)})},
				},
			},
			Sender: mustAddr(t, "90f3e6d73b5730f16974f4df1d3441394ebae62186baf83608599f226455afa7"),
			GasData: txn.GasData{
				Payment: []txn.ObjectRef{{
					ObjectID: mustAddr(t, "0000000000000000000000000000000000000000000000000000000000000005"),
					Version:  1,
					Digest:   mustDigest(t, "0000000000000000000000000000000000000000000000000000000000000001"),
				}},
				Owner:  mustAddr(t, "90f3e6d73b5730f16974f4df1d3441394ebae62186baf83608599f226455afa7"),
				Price:  1000,
				Budget: 5000000,
			},
			Expiration: txn.TransactionExpiration{Epoch: &epoch},
		},
	}

	var decoded txn.TransactionData
	roundTrip(t, "TransactionData_WithExpiration",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)

	if decoded.V1.Expiration.Epoch == nil {
		t.Fatal("expected Epoch to be non-nil")
	}
	if *decoded.V1.Expiration.Epoch != epoch {
		t.Errorf("Epoch: got %d, want %d", *decoded.V1.Expiration.Epoch, epoch)
	}
}

// --- 15. ProgrammableTransaction ---

func TestRoundTrip_ProgrammableTransaction(t *testing.T) {
	amountEnc := bcs.NewEncoder()
	amt := bcs.U64(500000000)
	amt.MarshalBCS(amountEnc)

	coinRef := txn.ObjectRef{
		ObjectID: mustAddr(t, "1111111111111111111111111111111111111111111111111111111111111111"),
		Version:  10,
		Digest:   mustDigest(t, "2222222222222222222222222222222222222222222222222222222222222222"),
	}

	original := txn.ProgrammableTransaction{
		Inputs: []txn.CallArg{
			txn.ObjectCallArg(txn.ImmOrOwnedObject(coinRef)),
			txn.PureCallArg(amountEnc.Bytes()),
		},
		Commands: []txn.Command{
			txn.MoveCallCommand(txn.MoveCall{
				Package:  mustAddr(t, "0000000000000000000000000000000000000000000000000000000000000002"),
				Module:   "pay",
				Function: "split",
				TypeArgs: []txn.TypeTag{
					txn.TypeTagStruct(txn.StructTag{
						Address:    mustAddr(t, "0000000000000000000000000000000000000000000000000000000000000002"),
						Module:     "sui",
						Name:       "SUI",
						TypeParams: []txn.TypeTag{},
					}),
				},
				Args: []txn.Argument{txn.Input(0), txn.Input(1)},
			}),
		},
	}

	var decoded txn.ProgrammableTransaction
	roundTrip(t, "ProgrammableTransaction",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)
}

func TestRoundTrip_ProgrammableTransaction_Empty(t *testing.T) {
	original := txn.ProgrammableTransaction{
		Inputs:   []txn.CallArg{},
		Commands: []txn.Command{},
	}

	var decoded txn.ProgrammableTransaction
	roundTrip(t, "ProgrammableTransaction_Empty",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)
}

func TestRoundTrip_ProgrammableTransaction_MultipleCommands(t *testing.T) {
	amountEnc := bcs.NewEncoder()
	amt := bcs.U64(1000000000)
	amt.MarshalBCS(amountEnc)

	recipientAddr := mustAddr(t, "fd233cd9a5dd7e577f16fa523427c75fbc382af1583c39fdf1c6747d2ed807a3")
	recipientEnc := bcs.NewEncoder()
	recipientAddr.MarshalBCS(recipientEnc)

	coin1 := txn.ObjectRef{
		ObjectID: mustAddr(t, "3333333333333333333333333333333333333333333333333333333333333333"),
		Version:  5,
		Digest:   mustDigest(t, "4444444444444444444444444444444444444444444444444444444444444444"),
	}

	original := txn.ProgrammableTransaction{
		Inputs: []txn.CallArg{
			txn.PureCallArg(amountEnc.Bytes()),
			txn.PureCallArg(recipientEnc.Bytes()),
			txn.ObjectCallArg(txn.ImmOrOwnedObject(coin1)),
		},
		Commands: []txn.Command{
			txn.SplitCoinsCommand(txn.GasCoin(), []txn.Argument{txn.Input(0)}),
			txn.TransferObjectsCommand([]txn.Argument{txn.Result(0)}, txn.Input(1)),
			txn.MergeCoinsCommand(txn.Input(2), []txn.Argument{txn.GasCoin()}),
		},
	}

	var decoded txn.ProgrammableTransaction
	roundTrip(t, "ProgrammableTransaction_MultipleCommands",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)
}

// --- Complex end-to-end: shared object + MoveCall round-trip ---

func TestRoundTrip_TransactionData_SharedObjectMoveCall(t *testing.T) {
	amountEnc := bcs.NewEncoder()
	amt := bcs.U64(42)
	amt.MarshalBCS(amountEnc)

	original := txn.TransactionData{
		V1: &txn.TransactionDataV1{
			Kind: txn.TransactionKind{
				ProgrammableTx: &txn.ProgrammableTransaction{
					Inputs: []txn.CallArg{
						txn.ObjectCallArg(txn.SharedObject(txn.SharedObjectRef{
							ObjectID:             mustAddr(t, "7777777777777777777777777777777777777777777777777777777777777777"),
							InitialSharedVersion: 1,
							Mutable:              true,
						})),
						txn.PureCallArg(amountEnc.Bytes()),
					},
					Commands: []txn.Command{
						txn.MoveCallCommand(txn.MoveCall{
							Package:  mustAddr(t, "8888888888888888888888888888888888888888888888888888888888888888"),
							Module:   "counter",
							Function: "increment",
							TypeArgs: []txn.TypeTag{},
							Args:     []txn.Argument{txn.Input(0), txn.Input(1)},
						}),
					},
				},
			},
			Sender: mustAddr(t, "90f3e6d73b5730f16974f4df1d3441394ebae62186baf83608599f226455afa7"),
			GasData: txn.GasData{
				Payment: []txn.ObjectRef{{
					ObjectID: mustAddr(t, "0000000000000000000000000000000000000000000000000000000000000005"),
					Version:  1,
					Digest:   mustDigest(t, "0000000000000000000000000000000000000000000000000000000000000001"),
				}},
				Owner:  mustAddr(t, "90f3e6d73b5730f16974f4df1d3441394ebae62186baf83608599f226455afa7"),
				Price:  1000,
				Budget: 5000000,
			},
			Expiration: txn.TransactionExpiration{},
		},
	}

	var decoded txn.TransactionData
	roundTrip(t, "TransactionData_SharedObjectMoveCall",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)
}

// --- CallArg with SharedObject via ObjectCallArg ---

func TestRoundTrip_CallArg_ObjectShared(t *testing.T) {
	original := txn.ObjectCallArg(txn.SharedObject(txn.SharedObjectRef{
		ObjectID:             mustAddr(t, "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"),
		InitialSharedVersion: 25,
		Mutable:              false,
	}))

	var decoded txn.CallArg
	roundTrip(t, "CallArg_ObjectShared",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)
}

func TestRoundTrip_CallArg_ObjectReceiving(t *testing.T) {
	ref := txn.ObjectRef{
		ObjectID: mustAddr(t, "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"),
		Version:  33,
		Digest:   mustDigest(t, "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
	}
	original := txn.ObjectCallArg(txn.Receiving(ref))

	var decoded txn.CallArg
	roundTrip(t, "CallArg_ObjectReceiving",
		func(e *bcs.Encoder) error { return original.MarshalBCS(e) },
		func(d *bcs.Decoder) error { return decoded.UnmarshalBCS(d) },
		func(e *bcs.Encoder) error { return decoded.MarshalBCS(e) },
	)
}
