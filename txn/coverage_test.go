package txn_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/inodrahq/go-bcs"
	"github.com/inodrahq/go-sui-sdk/txn"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// encode encodes a value and returns the BCS bytes.
func encode(t *testing.T, m bcs.Marshaler) []byte {
	t.Helper()
	e := bcs.NewEncoder()
	if err := m.MarshalBCS(e); err != nil {
		t.Fatalf("encode: %v", err)
	}
	return e.Bytes()
}

// encodeDecode encodes then decodes, asserting no error and no trailing bytes,
// and returns the decoded bytes for optional further assertions.
func encodeDecode(t *testing.T, m bcs.Marshaler, u bcs.Unmarshaler) []byte {
	t.Helper()
	b := encode(t, m)
	d := bcs.NewDecoder(b)
	if err := u.UnmarshalBCS(d); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if d.Remaining() != 0 {
		t.Fatalf("%d trailing bytes", d.Remaining())
	}
	return b
}

// mustDecode decodes raw bytes into u and asserts success.
func mustDecode(t *testing.T, raw []byte, u bcs.Unmarshaler) {
	t.Helper()
	d := bcs.NewDecoder(raw)
	if err := u.UnmarshalBCS(d); err != nil {
		t.Fatalf("decode: %v", err)
	}
}

// expectDecodeErr decodes raw bytes and expects an error.
func expectDecodeErr(t *testing.T, raw []byte, u bcs.Unmarshaler) {
	t.Helper()
	d := bcs.NewDecoder(raw)
	if err := u.UnmarshalBCS(d); err == nil {
		t.Fatal("expected decode error, got nil")
	}
}

func addr(t *testing.T, s string) txn.SuiAddress {
	t.Helper()
	a, err := txn.ParseAddress(s)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func digest(t *testing.T, s string) txn.ObjectDigest {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatal(err)
	}
	var d txn.ObjectDigest
	copy(d[:], b)
	return d
}

func sampleObjectRef(t *testing.T) txn.ObjectRef {
	t.Helper()
	return txn.ObjectRef{
		ObjectID: addr(t, "1111111111111111111111111111111111111111111111111111111111111111"),
		Version:  42,
		Digest:   digest(t, "2222222222222222222222222222222222222222222222222222222222222222"),
	}
}

func sampleGasData(t *testing.T) txn.GasData {
	t.Helper()
	return txn.GasData{
		Payment: []txn.ObjectRef{
			{
				ObjectID: addr(t, "0000000000000000000000000000000000000000000000000000000000000005"),
				Version:  1,
				Digest:   digest(t, "0000000000000000000000000000000000000000000000000000000000000001"),
			},
		},
		Owner:  addr(t, "90f3e6d73b5730f16974f4df1d3441394ebae62186baf83608599f226455afa7"),
		Price:  1000,
		Budget: 10000000,
	}
}

// ---------------------------------------------------------------------------
// SuiAddress
// ---------------------------------------------------------------------------

func TestAddress_ParseTooLong(t *testing.T) {
	// 66 hex chars = 33 bytes, should fail
	_, err := txn.ParseAddress("0x" + "aa" + "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	if err == nil {
		t.Fatal("expected error for 33-byte address")
	}
}

func TestAddress_UnmarshalTruncated(t *testing.T) {
	var a txn.SuiAddress
	expectDecodeErr(t, []byte{0x01, 0x02}, &a)
}

// ---------------------------------------------------------------------------
// ObjectDigest
// ---------------------------------------------------------------------------

func TestObjectDigest_UnmarshalBadLength(t *testing.T) {
	// Encode length prefix of 16 instead of 32 — should error
	e := bcs.NewEncoder()
	e.WriteULEB128(16)
	e.WriteBytes(make([]byte, 16))
	var d txn.ObjectDigest
	expectDecodeErr(t, e.Bytes(), &d)
}

func TestObjectDigest_UnmarshalTruncated(t *testing.T) {
	var d txn.ObjectDigest
	expectDecodeErr(t, []byte{}, &d)
}

func TestObjectDigest_UnmarshalTruncatedData(t *testing.T) {
	e := bcs.NewEncoder()
	e.WriteULEB128(32)
	e.WriteBytes(make([]byte, 10)) // only 10 bytes instead of 32
	var d txn.ObjectDigest
	expectDecodeErr(t, e.Bytes(), &d)
}

// ---------------------------------------------------------------------------
// ObjectRef
// ---------------------------------------------------------------------------

func TestObjectRef_EncodeDecode(t *testing.T) {
	original := sampleObjectRef(t)
	var decoded txn.ObjectRef
	b := encodeDecode(t, original, &decoded)

	// re-encode and compare
	b2 := encode(t, decoded)
	if !bytes.Equal(b, b2) {
		t.Error("ObjectRef re-encode mismatch")
	}
}

func TestObjectRef_UnmarshalTruncated(t *testing.T) {
	var r txn.ObjectRef
	expectDecodeErr(t, []byte{0x01, 0x02}, &r)
}

// ---------------------------------------------------------------------------
// SharedObjectRef
// ---------------------------------------------------------------------------

func TestSharedObjectRef_EncodeDecode(t *testing.T) {
	original := txn.SharedObjectRef{
		ObjectID:             addr(t, "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"),
		InitialSharedVersion: 50,
		Mutable:              true,
	}
	var decoded txn.SharedObjectRef
	b := encodeDecode(t, original, &decoded)
	b2 := encode(t, decoded)
	if !bytes.Equal(b, b2) {
		t.Error("SharedObjectRef re-encode mismatch")
	}
}

func TestSharedObjectRef_UnmarshalTruncated(t *testing.T) {
	var r txn.SharedObjectRef
	expectDecodeErr(t, []byte{0x01}, &r)
}

// ---------------------------------------------------------------------------
// Argument — error paths
// ---------------------------------------------------------------------------

func TestArgument_UnmarshalBadTag(t *testing.T) {
	e := bcs.NewEncoder()
	e.WriteULEB128(99) // invalid tag
	var a txn.Argument
	expectDecodeErr(t, e.Bytes(), &a)
}

func TestArgument_UnmarshalEmpty(t *testing.T) {
	var a txn.Argument
	expectDecodeErr(t, []byte{}, &a)
}

func TestArgument_UnmarshalInputTruncated(t *testing.T) {
	e := bcs.NewEncoder()
	e.WriteULEB128(1) // Input tag, but no index follows
	var a txn.Argument
	expectDecodeErr(t, e.Bytes(), &a)
}

func TestArgument_UnmarshalResultTruncated(t *testing.T) {
	e := bcs.NewEncoder()
	e.WriteULEB128(2) // Result tag, but no index follows
	var a txn.Argument
	expectDecodeErr(t, e.Bytes(), &a)
}

func TestArgument_UnmarshalNestedResultTruncated(t *testing.T) {
	e := bcs.NewEncoder()
	e.WriteULEB128(3) // NestedResult tag, but no data follows
	var a txn.Argument
	expectDecodeErr(t, e.Bytes(), &a)
}

func TestArgument_UnmarshalNestedResultPartial(t *testing.T) {
	e := bcs.NewEncoder()
	e.WriteULEB128(3) // NestedResult tag
	u := bcs.U16(5)
	u.MarshalBCS(e) // CmdIndex only, missing ResultIndex
	var a txn.Argument
	expectDecodeErr(t, e.Bytes(), &a)
}

// Direct encode/decode for each variant to cover MarshalBCS
func TestArgument_GasCoin_EncodeDecode(t *testing.T) {
	original := txn.GasCoin()
	var decoded txn.Argument
	encodeDecode(t, original, &decoded)
	// re-encode should match
	b1 := encode(t, original)
	b2 := encode(t, decoded)
	if !bytes.Equal(b1, b2) {
		t.Error("GasCoin re-encode mismatch")
	}
}

func TestArgument_Input_EncodeDecode(t *testing.T) {
	original := txn.Input(255)
	var decoded txn.Argument
	b := encodeDecode(t, original, &decoded)
	b2 := encode(t, decoded)
	if !bytes.Equal(b, b2) {
		t.Error("Input re-encode mismatch")
	}
}

func TestArgument_Result_EncodeDecode(t *testing.T) {
	original := txn.Result(128)
	var decoded txn.Argument
	b := encodeDecode(t, original, &decoded)
	b2 := encode(t, decoded)
	if !bytes.Equal(b, b2) {
		t.Error("Result re-encode mismatch")
	}
}

func TestArgument_NestedResult_EncodeDecode(t *testing.T) {
	original := txn.NestedResult(10, 20)
	var decoded txn.Argument
	b := encodeDecode(t, original, &decoded)
	b2 := encode(t, decoded)
	if !bytes.Equal(b, b2) {
		t.Error("NestedResult re-encode mismatch")
	}
}

// ---------------------------------------------------------------------------
// CallArg — error paths
// ---------------------------------------------------------------------------

func TestCallArg_UnmarshalBadTag(t *testing.T) {
	e := bcs.NewEncoder()
	e.WriteULEB128(99)
	var c txn.CallArg
	expectDecodeErr(t, e.Bytes(), &c)
}

func TestCallArg_UnmarshalEmpty(t *testing.T) {
	var c txn.CallArg
	expectDecodeErr(t, []byte{}, &c)
}

func TestCallArg_UnmarshalPureTruncated(t *testing.T) {
	e := bcs.NewEncoder()
	e.WriteULEB128(0) // Pure tag, but no length/data follows
	var c txn.CallArg
	// This might or might not error depending on ByteVector handling of empty.
	// We just ensure it doesn't panic.
	d := bcs.NewDecoder(e.Bytes())
	_ = c.UnmarshalBCS(d)
}

func TestCallArg_UnmarshalObjectTruncated(t *testing.T) {
	e := bcs.NewEncoder()
	e.WriteULEB128(1) // Object tag, but no ObjectArg data follows
	var c txn.CallArg
	expectDecodeErr(t, e.Bytes(), &c)
}

func TestCallArg_Pure_EncodeDecode(t *testing.T) {
	original := txn.PureCallArg([]byte{0xDE, 0xAD, 0xBE, 0xEF})
	var decoded txn.CallArg
	b := encodeDecode(t, original, &decoded)
	b2 := encode(t, decoded)
	if !bytes.Equal(b, b2) {
		t.Error("Pure CallArg re-encode mismatch")
	}
}

func TestCallArg_ObjectImmOrOwned_EncodeDecode(t *testing.T) {
	ref := sampleObjectRef(t)
	original := txn.ObjectCallArg(txn.ImmOrOwnedObject(ref))
	var decoded txn.CallArg
	b := encodeDecode(t, original, &decoded)
	b2 := encode(t, decoded)
	if !bytes.Equal(b, b2) {
		t.Error("Object CallArg re-encode mismatch")
	}
}

// ---------------------------------------------------------------------------
// ObjectArg — error paths
// ---------------------------------------------------------------------------

func TestObjectArg_UnmarshalBadTag(t *testing.T) {
	e := bcs.NewEncoder()
	e.WriteULEB128(99)
	var o txn.ObjectArg
	expectDecodeErr(t, e.Bytes(), &o)
}

func TestObjectArg_UnmarshalEmpty(t *testing.T) {
	var o txn.ObjectArg
	expectDecodeErr(t, []byte{}, &o)
}

func TestObjectArg_UnmarshalImmOrOwnedTruncated(t *testing.T) {
	e := bcs.NewEncoder()
	e.WriteULEB128(0) // ImmOrOwned tag, but no ObjectRef data follows
	var o txn.ObjectArg
	expectDecodeErr(t, e.Bytes(), &o)
}

func TestObjectArg_UnmarshalSharedTruncated(t *testing.T) {
	e := bcs.NewEncoder()
	e.WriteULEB128(1) // Shared tag, but no SharedObjectRef data follows
	var o txn.ObjectArg
	expectDecodeErr(t, e.Bytes(), &o)
}

func TestObjectArg_UnmarshalReceivingTruncated(t *testing.T) {
	e := bcs.NewEncoder()
	e.WriteULEB128(2) // Receiving tag, but no ObjectRef data follows
	var o txn.ObjectArg
	expectDecodeErr(t, e.Bytes(), &o)
}

func TestObjectArg_ImmOrOwned_EncodeDecode(t *testing.T) {
	ref := sampleObjectRef(t)
	original := txn.ImmOrOwnedObject(ref)
	var decoded txn.ObjectArg
	b := encodeDecode(t, original, &decoded)
	b2 := encode(t, decoded)
	if !bytes.Equal(b, b2) {
		t.Error("ImmOrOwned ObjectArg re-encode mismatch")
	}
}

func TestObjectArg_Shared_EncodeDecode(t *testing.T) {
	original := txn.SharedObject(txn.SharedObjectRef{
		ObjectID:             addr(t, "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"),
		InitialSharedVersion: 999,
		Mutable:              false,
	})
	var decoded txn.ObjectArg
	b := encodeDecode(t, original, &decoded)
	b2 := encode(t, decoded)
	if !bytes.Equal(b, b2) {
		t.Error("Shared ObjectArg re-encode mismatch")
	}
}

func TestObjectArg_Receiving_EncodeDecode(t *testing.T) {
	ref := sampleObjectRef(t)
	original := txn.Receiving(ref)
	var decoded txn.ObjectArg
	b := encodeDecode(t, original, &decoded)
	b2 := encode(t, decoded)
	if !bytes.Equal(b, b2) {
		t.Error("Receiving ObjectArg re-encode mismatch")
	}
}

// ---------------------------------------------------------------------------
// Command — error paths and encode/decode
// ---------------------------------------------------------------------------

func TestCommand_UnmarshalBadTag(t *testing.T) {
	e := bcs.NewEncoder()
	e.WriteULEB128(99)
	var c txn.Command
	expectDecodeErr(t, e.Bytes(), &c)
}

func TestCommand_UnmarshalEmpty(t *testing.T) {
	var c txn.Command
	expectDecodeErr(t, []byte{}, &c)
}

func TestCommand_MoveCall_UnmarshalTruncated(t *testing.T) {
	e := bcs.NewEncoder()
	e.WriteULEB128(0) // MoveCall tag, but no data
	var c txn.Command
	expectDecodeErr(t, e.Bytes(), &c)
}

func TestCommand_TransferObjects_UnmarshalTruncated(t *testing.T) {
	e := bcs.NewEncoder()
	e.WriteULEB128(1) // TransferObjects tag, but no data
	var c txn.Command
	expectDecodeErr(t, e.Bytes(), &c)
}

func TestCommand_SplitCoins_UnmarshalTruncated(t *testing.T) {
	e := bcs.NewEncoder()
	e.WriteULEB128(2) // SplitCoins tag, but no data
	var c txn.Command
	expectDecodeErr(t, e.Bytes(), &c)
}

func TestCommand_MergeCoins_UnmarshalTruncated(t *testing.T) {
	e := bcs.NewEncoder()
	e.WriteULEB128(3) // MergeCoins tag, but no data
	var c txn.Command
	expectDecodeErr(t, e.Bytes(), &c)
}

func TestCommand_MoveCall_EncodeDecode(t *testing.T) {
	original := txn.MoveCallCommand(txn.MoveCall{
		Package:  addr(t, "0000000000000000000000000000000000000000000000000000000000000002"),
		Module:   "pay",
		Function: "split",
		TypeArgs: []txn.TypeTag{
			txn.TypeTagStruct(txn.StructTag{
				Address:    addr(t, "0000000000000000000000000000000000000000000000000000000000000002"),
				Module:     "sui",
				Name:       "SUI",
				TypeParams: []txn.TypeTag{},
			}),
		},
		Args: []txn.Argument{txn.Input(0), txn.Input(1)},
	})
	var decoded txn.Command
	b := encodeDecode(t, original, &decoded)
	b2 := encode(t, decoded)
	if !bytes.Equal(b, b2) {
		t.Error("MoveCall Command re-encode mismatch")
	}
}

func TestCommand_TransferObjects_EncodeDecode(t *testing.T) {
	original := txn.TransferObjectsCommand(
		[]txn.Argument{txn.Result(0), txn.Result(1)},
		txn.Input(0),
	)
	var decoded txn.Command
	b := encodeDecode(t, original, &decoded)
	b2 := encode(t, decoded)
	if !bytes.Equal(b, b2) {
		t.Error("TransferObjects Command re-encode mismatch")
	}
}

func TestCommand_SplitCoins_EncodeDecode(t *testing.T) {
	original := txn.SplitCoinsCommand(
		txn.GasCoin(),
		[]txn.Argument{txn.Input(0), txn.Input(1)},
	)
	var decoded txn.Command
	b := encodeDecode(t, original, &decoded)
	b2 := encode(t, decoded)
	if !bytes.Equal(b, b2) {
		t.Error("SplitCoins Command re-encode mismatch")
	}
}

func TestCommand_MergeCoins_EncodeDecode(t *testing.T) {
	original := txn.MergeCoinsCommand(
		txn.Input(0),
		[]txn.Argument{txn.Input(1), txn.Input(2)},
	)
	var decoded txn.Command
	b := encodeDecode(t, original, &decoded)
	b2 := encode(t, decoded)
	if !bytes.Equal(b, b2) {
		t.Error("MergeCoins Command re-encode mismatch")
	}
}

// ---------------------------------------------------------------------------
// TypeTag — error paths and encode/decode
// ---------------------------------------------------------------------------

func TestTypeTag_UnmarshalBadTag(t *testing.T) {
	e := bcs.NewEncoder()
	e.WriteULEB128(99) // invalid tag
	var tt txn.TypeTag
	expectDecodeErr(t, e.Bytes(), &tt)
}

func TestTypeTag_UnmarshalEmpty(t *testing.T) {
	var tt txn.TypeTag
	expectDecodeErr(t, []byte{}, &tt)
}

func TestTypeTag_Vector_EncodeDecode(t *testing.T) {
	original := txn.TypeTagVector(txn.TypeTagU8())
	var decoded txn.TypeTag
	b := encodeDecode(t, original, &decoded)
	b2 := encode(t, decoded)
	if !bytes.Equal(b, b2) {
		t.Error("Vector TypeTag re-encode mismatch")
	}
}

func TestTypeTag_VectorOfVector_EncodeDecode(t *testing.T) {
	original := txn.TypeTagVector(txn.TypeTagVector(txn.TypeTagU64()))
	var decoded txn.TypeTag
	b := encodeDecode(t, original, &decoded)
	b2 := encode(t, decoded)
	if !bytes.Equal(b, b2) {
		t.Error("VectorOfVector TypeTag re-encode mismatch")
	}
}

func TestTypeTag_VectorTruncated(t *testing.T) {
	e := bcs.NewEncoder()
	e.WriteULEB128(6) // Vector tag, but no inner type
	var tt txn.TypeTag
	expectDecodeErr(t, e.Bytes(), &tt)
}

func TestTypeTag_StructTruncated(t *testing.T) {
	e := bcs.NewEncoder()
	e.WriteULEB128(7) // Struct tag, but no struct data
	var tt txn.TypeTag
	expectDecodeErr(t, e.Bytes(), &tt)
}

func TestTypeTag_AllSimpleVariants(t *testing.T) {
	tags := []struct {
		name string
		tag  txn.TypeTag
	}{
		{"bool", txn.TypeTagBool()},
		{"u8", txn.TypeTagU8()},
		{"u64", txn.TypeTagU64()},
		{"u128", txn.TypeTagU128()},
		{"address", txn.TypeTagAddress()},
		{"signer", txn.TypeTagSigner()},
		{"u16", txn.TypeTagU16()},
		{"u32", txn.TypeTagU32()},
		{"u256", txn.TypeTagU256()},
	}
	for _, tc := range tags {
		t.Run(tc.name, func(t *testing.T) {
			var decoded txn.TypeTag
			b := encodeDecode(t, tc.tag, &decoded)
			b2 := encode(t, decoded)
			if !bytes.Equal(b, b2) {
				t.Errorf("%s TypeTag re-encode mismatch", tc.name)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// StructTag — encode/decode with type params
// ---------------------------------------------------------------------------

func TestStructTag_NoTypeParams_EncodeDecode(t *testing.T) {
	original := txn.StructTag{
		Address:    addr(t, "0000000000000000000000000000000000000000000000000000000000000002"),
		Module:     "sui",
		Name:       "SUI",
		TypeParams: []txn.TypeTag{},
	}
	var decoded txn.StructTag
	b := encodeDecode(t, original, &decoded)
	b2 := encode(t, decoded)
	if !bytes.Equal(b, b2) {
		t.Error("StructTag re-encode mismatch")
	}
}

func TestStructTag_WithTypeParams_EncodeDecode(t *testing.T) {
	original := txn.StructTag{
		Address: addr(t, "0000000000000000000000000000000000000000000000000000000000000002"),
		Module:  "coin",
		Name:    "Coin",
		TypeParams: []txn.TypeTag{
			txn.TypeTagStruct(txn.StructTag{
				Address:    addr(t, "0000000000000000000000000000000000000000000000000000000000000002"),
				Module:     "sui",
				Name:       "SUI",
				TypeParams: []txn.TypeTag{},
			}),
			txn.TypeTagU64(),
		},
	}
	var decoded txn.StructTag
	b := encodeDecode(t, original, &decoded)
	b2 := encode(t, decoded)
	if !bytes.Equal(b, b2) {
		t.Error("StructTag with TypeParams re-encode mismatch")
	}
}

func TestStructTag_UnmarshalTruncated(t *testing.T) {
	var s txn.StructTag
	expectDecodeErr(t, []byte{0x01}, &s)
}

// ---------------------------------------------------------------------------
// GasData — encode/decode
// ---------------------------------------------------------------------------

func TestGasData_SinglePayment_EncodeDecode(t *testing.T) {
	original := sampleGasData(t)
	var decoded txn.GasData
	b := encodeDecode(t, original, &decoded)
	b2 := encode(t, decoded)
	if !bytes.Equal(b, b2) {
		t.Error("GasData re-encode mismatch")
	}
	if len(decoded.Payment) != 1 {
		t.Errorf("expected 1 payment, got %d", len(decoded.Payment))
	}
	if decoded.Price != 1000 {
		t.Errorf("expected price 1000, got %d", decoded.Price)
	}
	if decoded.Budget != 10000000 {
		t.Errorf("expected budget 10000000, got %d", decoded.Budget)
	}
}

func TestGasData_MultiplePayments_EncodeDecode(t *testing.T) {
	original := txn.GasData{
		Payment: []txn.ObjectRef{
			{
				ObjectID: addr(t, "0000000000000000000000000000000000000000000000000000000000000005"),
				Version:  1,
				Digest:   digest(t, "0000000000000000000000000000000000000000000000000000000000000001"),
			},
			{
				ObjectID: addr(t, "0000000000000000000000000000000000000000000000000000000000000006"),
				Version:  2,
				Digest:   digest(t, "0000000000000000000000000000000000000000000000000000000000000002"),
			},
			{
				ObjectID: addr(t, "0000000000000000000000000000000000000000000000000000000000000007"),
				Version:  3,
				Digest:   digest(t, "0000000000000000000000000000000000000000000000000000000000000003"),
			},
		},
		Owner:  addr(t, "90f3e6d73b5730f16974f4df1d3441394ebae62186baf83608599f226455afa7"),
		Price:  750,
		Budget: 5000000,
	}
	var decoded txn.GasData
	b := encodeDecode(t, original, &decoded)
	b2 := encode(t, decoded)
	if !bytes.Equal(b, b2) {
		t.Error("GasData multiple payments re-encode mismatch")
	}
	if len(decoded.Payment) != 3 {
		t.Errorf("expected 3 payments, got %d", len(decoded.Payment))
	}
}

func TestGasData_EmptyPayment_EncodeDecode(t *testing.T) {
	original := txn.GasData{
		Payment: []txn.ObjectRef{},
		Owner:   addr(t, "90f3e6d73b5730f16974f4df1d3441394ebae62186baf83608599f226455afa7"),
		Price:   500,
		Budget:  1000000,
	}
	var decoded txn.GasData
	b := encodeDecode(t, original, &decoded)
	b2 := encode(t, decoded)
	if !bytes.Equal(b, b2) {
		t.Error("GasData empty payment re-encode mismatch")
	}
}

func TestGasData_UnmarshalTruncated(t *testing.T) {
	var g txn.GasData
	expectDecodeErr(t, []byte{}, &g)
}

func TestGasData_UnmarshalTruncatedAfterPayment(t *testing.T) {
	// Encode only the payment vec, then cut short
	e := bcs.NewEncoder()
	e.WriteULEB128(0) // 0 payments
	// Missing owner, price, budget
	var g txn.GasData
	expectDecodeErr(t, e.Bytes(), &g)
}

// ---------------------------------------------------------------------------
// TransactionExpiration — error paths
// ---------------------------------------------------------------------------

func TestTransactionExpiration_None_EncodeDecode(t *testing.T) {
	original := txn.TransactionExpiration{}
	var decoded txn.TransactionExpiration
	b := encodeDecode(t, original, &decoded)
	b2 := encode(t, decoded)
	if !bytes.Equal(b, b2) {
		t.Error("TransactionExpiration None re-encode mismatch")
	}
	if decoded.Epoch != nil {
		t.Error("expected nil Epoch")
	}
}

func TestTransactionExpiration_Epoch_EncodeDecode(t *testing.T) {
	epoch := bcs.U64(99999)
	original := txn.TransactionExpiration{Epoch: &epoch}
	var decoded txn.TransactionExpiration
	b := encodeDecode(t, original, &decoded)
	b2 := encode(t, decoded)
	if !bytes.Equal(b, b2) {
		t.Error("TransactionExpiration Epoch re-encode mismatch")
	}
	if decoded.Epoch == nil {
		t.Fatal("expected non-nil Epoch")
	}
	if *decoded.Epoch != 99999 {
		t.Errorf("expected 99999, got %d", *decoded.Epoch)
	}
}

func TestTransactionExpiration_UnmarshalBadTag(t *testing.T) {
	e := bcs.NewEncoder()
	e.WriteULEB128(99)
	var te txn.TransactionExpiration
	expectDecodeErr(t, e.Bytes(), &te)
}

func TestTransactionExpiration_UnmarshalEmpty(t *testing.T) {
	var te txn.TransactionExpiration
	expectDecodeErr(t, []byte{}, &te)
}

func TestTransactionExpiration_UnmarshalEpochTruncated(t *testing.T) {
	e := bcs.NewEncoder()
	e.WriteULEB128(1) // Epoch tag, but no epoch value
	var te txn.TransactionExpiration
	expectDecodeErr(t, e.Bytes(), &te)
}

// ---------------------------------------------------------------------------
// TransactionKind — error paths
// ---------------------------------------------------------------------------

func TestTransactionKind_UnmarshalBadTag(t *testing.T) {
	e := bcs.NewEncoder()
	e.WriteULEB128(99)
	var k txn.TransactionKind
	expectDecodeErr(t, e.Bytes(), &k)
}

func TestTransactionKind_UnmarshalEmpty(t *testing.T) {
	var k txn.TransactionKind
	expectDecodeErr(t, []byte{}, &k)
}

func TestTransactionKind_EncodeDecode(t *testing.T) {
	pt := &txn.ProgrammableTransaction{
		Inputs:   []txn.CallArg{txn.PureCallArg([]byte{0x42})},
		Commands: []txn.Command{txn.SplitCoinsCommand(txn.GasCoin(), []txn.Argument{txn.Input(0)})},
	}
	original := txn.TransactionKind{ProgrammableTx: pt}
	var decoded txn.TransactionKind
	b := encodeDecode(t, original, &decoded)
	b2 := encode(t, decoded)
	if !bytes.Equal(b, b2) {
		t.Error("TransactionKind re-encode mismatch")
	}
	if decoded.ProgrammableTx == nil {
		t.Fatal("expected non-nil ProgrammableTx")
	}
}

// ---------------------------------------------------------------------------
// TransactionDataV1 — encode/decode
// ---------------------------------------------------------------------------

func TestTransactionDataV1_EncodeDecode(t *testing.T) {
	amountEnc := bcs.NewEncoder()
	amt := bcs.U64(1000000000)
	amt.MarshalBCS(amountEnc)

	recipientAddr := addr(t, "fd233cd9a5dd7e577f16fa523427c75fbc382af1583c39fdf1c6747d2ed807a3")
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
		Sender:     addr(t, "90f3e6d73b5730f16974f4df1d3441394ebae62186baf83608599f226455afa7"),
		GasData:    sampleGasData(t),
		Expiration: txn.TransactionExpiration{},
	}
	var decoded txn.TransactionDataV1
	b := encodeDecode(t, original, &decoded)
	b2 := encode(t, decoded)
	if !bytes.Equal(b, b2) {
		t.Error("TransactionDataV1 re-encode mismatch")
	}
}

func TestTransactionDataV1_WithExpiration_EncodeDecode(t *testing.T) {
	epoch := bcs.U64(12345)
	original := txn.TransactionDataV1{
		Kind: txn.TransactionKind{
			ProgrammableTx: &txn.ProgrammableTransaction{
				Inputs:   []txn.CallArg{txn.PureCallArg([]byte{0x01})},
				Commands: []txn.Command{txn.SplitCoinsCommand(txn.GasCoin(), []txn.Argument{txn.Input(0)})},
			},
		},
		Sender:     addr(t, "90f3e6d73b5730f16974f4df1d3441394ebae62186baf83608599f226455afa7"),
		GasData:    sampleGasData(t),
		Expiration: txn.TransactionExpiration{Epoch: &epoch},
	}
	var decoded txn.TransactionDataV1
	b := encodeDecode(t, original, &decoded)
	b2 := encode(t, decoded)
	if !bytes.Equal(b, b2) {
		t.Error("TransactionDataV1 with expiration re-encode mismatch")
	}
	if decoded.Expiration.Epoch == nil || *decoded.Expiration.Epoch != 12345 {
		t.Error("expiration epoch mismatch")
	}
}

func TestTransactionDataV1_UnmarshalTruncated(t *testing.T) {
	var v txn.TransactionDataV1
	expectDecodeErr(t, []byte{}, &v)
}

// ---------------------------------------------------------------------------
// TransactionData — error paths and encode/decode
// ---------------------------------------------------------------------------

func TestTransactionData_UnmarshalBadVersion(t *testing.T) {
	e := bcs.NewEncoder()
	e.WriteULEB128(5) // Unknown version
	var td txn.TransactionData
	expectDecodeErr(t, e.Bytes(), &td)
}

func TestTransactionData_UnmarshalEmpty(t *testing.T) {
	var td txn.TransactionData
	expectDecodeErr(t, []byte{}, &td)
}

func TestTransactionData_EncodeDecode(t *testing.T) {
	epoch := bcs.U64(42)
	original := txn.TransactionData{
		V1: &txn.TransactionDataV1{
			Kind: txn.TransactionKind{
				ProgrammableTx: &txn.ProgrammableTransaction{
					Inputs:   []txn.CallArg{txn.PureCallArg([]byte{0xAB, 0xCD})},
					Commands: []txn.Command{txn.SplitCoinsCommand(txn.GasCoin(), []txn.Argument{txn.Input(0)})},
				},
			},
			Sender:     addr(t, "90f3e6d73b5730f16974f4df1d3441394ebae62186baf83608599f226455afa7"),
			GasData:    sampleGasData(t),
			Expiration: txn.TransactionExpiration{Epoch: &epoch},
		},
	}
	var decoded txn.TransactionData
	b := encodeDecode(t, original, &decoded)
	b2 := encode(t, decoded)
	if !bytes.Equal(b, b2) {
		t.Error("TransactionData re-encode mismatch")
	}
	if decoded.V1 == nil {
		t.Fatal("expected non-nil V1")
	}
	if decoded.V1.Expiration.Epoch == nil || *decoded.V1.Expiration.Epoch != 42 {
		t.Error("expiration epoch mismatch")
	}
}

// ---------------------------------------------------------------------------
// ProgrammableTransaction — encode/decode
// ---------------------------------------------------------------------------

func TestProgrammableTransaction_Empty_EncodeDecode(t *testing.T) {
	original := txn.ProgrammableTransaction{
		Inputs:   []txn.CallArg{},
		Commands: []txn.Command{},
	}
	var decoded txn.ProgrammableTransaction
	b := encodeDecode(t, original, &decoded)
	b2 := encode(t, decoded)
	if !bytes.Equal(b, b2) {
		t.Error("ProgrammableTransaction empty re-encode mismatch")
	}
}

func TestProgrammableTransaction_Complex_EncodeDecode(t *testing.T) {
	coin1 := sampleObjectRef(t)
	amountEnc := bcs.NewEncoder()
	amt := bcs.U64(500_000_000)
	amt.MarshalBCS(amountEnc)

	recipientAddr := addr(t, "fd233cd9a5dd7e577f16fa523427c75fbc382af1583c39fdf1c6747d2ed807a3")
	recipientEnc := bcs.NewEncoder()
	recipientAddr.MarshalBCS(recipientEnc)

	original := txn.ProgrammableTransaction{
		Inputs: []txn.CallArg{
			txn.PureCallArg(amountEnc.Bytes()),
			txn.PureCallArg(recipientEnc.Bytes()),
			txn.ObjectCallArg(txn.ImmOrOwnedObject(coin1)),
			txn.ObjectCallArg(txn.SharedObject(txn.SharedObjectRef{
				ObjectID:             addr(t, "7777777777777777777777777777777777777777777777777777777777777777"),
				InitialSharedVersion: 1,
				Mutable:              true,
			})),
		},
		Commands: []txn.Command{
			txn.SplitCoinsCommand(txn.GasCoin(), []txn.Argument{txn.Input(0)}),
			txn.TransferObjectsCommand([]txn.Argument{txn.Result(0)}, txn.Input(1)),
			txn.MergeCoinsCommand(txn.Input(2), []txn.Argument{txn.GasCoin()}),
			txn.MoveCallCommand(txn.MoveCall{
				Package:  addr(t, "0000000000000000000000000000000000000000000000000000000000000002"),
				Module:   "pay",
				Function: "split",
				TypeArgs: []txn.TypeTag{
					txn.TypeTagStruct(txn.StructTag{
						Address:    addr(t, "0000000000000000000000000000000000000000000000000000000000000002"),
						Module:     "sui",
						Name:       "SUI",
						TypeParams: []txn.TypeTag{},
					}),
				},
				Args: []txn.Argument{txn.Input(3), txn.NestedResult(0, 0)},
			}),
		},
	}
	var decoded txn.ProgrammableTransaction
	b := encodeDecode(t, original, &decoded)
	b2 := encode(t, decoded)
	if !bytes.Equal(b, b2) {
		t.Error("ProgrammableTransaction complex re-encode mismatch")
	}
	if len(decoded.Inputs) != 4 {
		t.Errorf("expected 4 inputs, got %d", len(decoded.Inputs))
	}
	if len(decoded.Commands) != 4 {
		t.Errorf("expected 4 commands, got %d", len(decoded.Commands))
	}
}

func TestProgrammableTransaction_UnmarshalTruncated(t *testing.T) {
	var pt txn.ProgrammableTransaction
	expectDecodeErr(t, []byte{}, &pt)
}

func TestProgrammableTransaction_UnmarshalTruncatedCommands(t *testing.T) {
	// Encode 0 inputs, then truncate before commands
	e := bcs.NewEncoder()
	e.WriteULEB128(0) // 0 inputs
	// Missing commands vec
	var pt txn.ProgrammableTransaction
	expectDecodeErr(t, e.Bytes(), &pt)
}

// ---------------------------------------------------------------------------
// Full TransactionData roundtrip — complex scenario
// ---------------------------------------------------------------------------

func TestTransactionData_FullComplex_EncodeDecode(t *testing.T) {
	amountEnc := bcs.NewEncoder()
	amt := bcs.U64(42)
	amt.MarshalBCS(amountEnc)

	epoch := bcs.U64(999)

	original := txn.TransactionData{
		V1: &txn.TransactionDataV1{
			Kind: txn.TransactionKind{
				ProgrammableTx: &txn.ProgrammableTransaction{
					Inputs: []txn.CallArg{
						txn.ObjectCallArg(txn.SharedObject(txn.SharedObjectRef{
							ObjectID:             addr(t, "7777777777777777777777777777777777777777777777777777777777777777"),
							InitialSharedVersion: 1,
							Mutable:              true,
						})),
						txn.PureCallArg(amountEnc.Bytes()),
						txn.ObjectCallArg(txn.Receiving(txn.ObjectRef{
							ObjectID: addr(t, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
							Version:  55,
							Digest:   digest(t, "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"),
						})),
					},
					Commands: []txn.Command{
						txn.MoveCallCommand(txn.MoveCall{
							Package:  addr(t, "8888888888888888888888888888888888888888888888888888888888888888"),
							Module:   "counter",
							Function: "increment",
							TypeArgs: []txn.TypeTag{txn.TypeTagVector(txn.TypeTagU8())},
							Args:     []txn.Argument{txn.Input(0), txn.Input(1), txn.Input(2)},
						}),
						txn.SplitCoinsCommand(txn.GasCoin(), []txn.Argument{txn.Input(1)}),
						txn.TransferObjectsCommand([]txn.Argument{txn.Result(1)}, txn.Input(0)),
					},
				},
			},
			Sender: addr(t, "90f3e6d73b5730f16974f4df1d3441394ebae62186baf83608599f226455afa7"),
			GasData: txn.GasData{
				Payment: []txn.ObjectRef{
					{
						ObjectID: addr(t, "0000000000000000000000000000000000000000000000000000000000000005"),
						Version:  1,
						Digest:   digest(t, "0000000000000000000000000000000000000000000000000000000000000001"),
					},
					{
						ObjectID: addr(t, "0000000000000000000000000000000000000000000000000000000000000006"),
						Version:  2,
						Digest:   digest(t, "0000000000000000000000000000000000000000000000000000000000000002"),
					},
				},
				Owner:  addr(t, "90f3e6d73b5730f16974f4df1d3441394ebae62186baf83608599f226455afa7"),
				Price:  1000,
				Budget: 50000000,
			},
			Expiration: txn.TransactionExpiration{Epoch: &epoch},
		},
	}

	var decoded txn.TransactionData
	b := encodeDecode(t, original, &decoded)
	b2 := encode(t, decoded)
	if !bytes.Equal(b, b2) {
		t.Error("FullComplex TransactionData re-encode mismatch")
	}

	// Validate decoded fields
	if decoded.V1 == nil {
		t.Fatal("expected non-nil V1")
	}
	v1 := decoded.V1
	if v1.Kind.ProgrammableTx == nil {
		t.Fatal("expected non-nil ProgrammableTx")
	}
	if len(v1.Kind.ProgrammableTx.Inputs) != 3 {
		t.Errorf("expected 3 inputs, got %d", len(v1.Kind.ProgrammableTx.Inputs))
	}
	if len(v1.Kind.ProgrammableTx.Commands) != 3 {
		t.Errorf("expected 3 commands, got %d", len(v1.Kind.ProgrammableTx.Commands))
	}
	if len(v1.GasData.Payment) != 2 {
		t.Errorf("expected 2 gas payments, got %d", len(v1.GasData.Payment))
	}
	if v1.GasData.Price != 1000 {
		t.Errorf("expected price 1000, got %d", v1.GasData.Price)
	}
	if v1.GasData.Budget != 50000000 {
		t.Errorf("expected budget 50000000, got %d", v1.GasData.Budget)
	}
	if v1.Expiration.Epoch == nil || *v1.Expiration.Epoch != 999 {
		t.Error("expiration epoch mismatch")
	}
}

// ---------------------------------------------------------------------------
// MoveCall — partial decode errors for inner fields
// ---------------------------------------------------------------------------

func TestMoveCall_UnmarshalTruncatedAfterPackage(t *testing.T) {
	// Encode only the package address
	e := bcs.NewEncoder()
	a := addr(t, "0000000000000000000000000000000000000000000000000000000000000002")
	a.MarshalBCS(e)
	// Missing module, function, typeargs, args
	var mc txn.MoveCall
	expectDecodeErr(t, e.Bytes(), &mc)
}

func TestMoveCall_UnmarshalTruncatedAfterModule(t *testing.T) {
	e := bcs.NewEncoder()
	a := addr(t, "0000000000000000000000000000000000000000000000000000000000000002")
	a.MarshalBCS(e)
	mod := bcs.String("pay")
	mod.MarshalBCS(e)
	// Missing function, typeargs, args
	var mc txn.MoveCall
	expectDecodeErr(t, e.Bytes(), &mc)
}

func TestMoveCall_UnmarshalTruncatedAfterFunction(t *testing.T) {
	e := bcs.NewEncoder()
	a := addr(t, "0000000000000000000000000000000000000000000000000000000000000002")
	a.MarshalBCS(e)
	mod := bcs.String("pay")
	mod.MarshalBCS(e)
	fn := bcs.String("split")
	fn.MarshalBCS(e)
	// Missing typeargs, args
	var mc txn.MoveCall
	expectDecodeErr(t, e.Bytes(), &mc)
}

func TestMoveCall_UnmarshalTruncatedAfterTypeArgs(t *testing.T) {
	e := bcs.NewEncoder()
	a := addr(t, "0000000000000000000000000000000000000000000000000000000000000002")
	a.MarshalBCS(e)
	mod := bcs.String("pay")
	mod.MarshalBCS(e)
	fn := bcs.String("split")
	fn.MarshalBCS(e)
	e.WriteULEB128(0) // 0 type args
	// Missing args
	var mc txn.MoveCall
	expectDecodeErr(t, e.Bytes(), &mc)
}

// ---------------------------------------------------------------------------
// TransferObjects — partial decode errors
// ---------------------------------------------------------------------------

func TestTransferObjects_UnmarshalTruncatedAfterObjects(t *testing.T) {
	e := bcs.NewEncoder()
	e.WriteULEB128(1)               // 1 object
	txn.Result(0).MarshalBCS(e)     // the object argument
	// Missing destination
	var to txn.TransferObjects
	expectDecodeErr(t, e.Bytes(), &to)
}

// ---------------------------------------------------------------------------
// SplitCoins — partial decode errors
// ---------------------------------------------------------------------------

func TestSplitCoins_UnmarshalTruncatedAfterCoin(t *testing.T) {
	e := bcs.NewEncoder()
	txn.GasCoin().MarshalBCS(e) // coin
	// Missing amounts vec
	var sc txn.SplitCoins
	expectDecodeErr(t, e.Bytes(), &sc)
}

// ---------------------------------------------------------------------------
// MergeCoins — partial decode errors
// ---------------------------------------------------------------------------

func TestMergeCoins_UnmarshalTruncatedAfterDest(t *testing.T) {
	e := bcs.NewEncoder()
	txn.Input(0).MarshalBCS(e) // destination
	// Missing sources vec
	var mc txn.MergeCoins
	expectDecodeErr(t, e.Bytes(), &mc)
}

// ---------------------------------------------------------------------------
// Decode from known BCS bytes — verifies specific byte patterns
// ---------------------------------------------------------------------------

func TestDecode_GasCoin_FromBytes(t *testing.T) {
	// GasCoin = ULEB128(0) = byte 0x00
	var a txn.Argument
	mustDecode(t, []byte{0x00}, &a)
	b := encode(t, a)
	if !bytes.Equal(b, []byte{0x00}) {
		t.Errorf("GasCoin bytes: got %x, want 00", b)
	}
}

func TestDecode_Input0_FromBytes(t *testing.T) {
	// Input(0) = ULEB128(1) + U16(0) = 0x01 0x00 0x00
	var a txn.Argument
	mustDecode(t, []byte{0x01, 0x00, 0x00}, &a)
	b := encode(t, a)
	if !bytes.Equal(b, []byte{0x01, 0x00, 0x00}) {
		t.Errorf("Input(0) bytes: got %x, want 010000", b)
	}
}

func TestDecode_Result1_FromBytes(t *testing.T) {
	// Result(1) = ULEB128(2) + U16(1) = 0x02 0x01 0x00
	var a txn.Argument
	mustDecode(t, []byte{0x02, 0x01, 0x00}, &a)
	b := encode(t, a)
	if !bytes.Equal(b, []byte{0x02, 0x01, 0x00}) {
		t.Errorf("Result(1) bytes: got %x, want 020100", b)
	}
}

func TestDecode_NestedResult_FromBytes(t *testing.T) {
	// NestedResult(0, 1) = ULEB128(3) + U16(0) + U16(1) = 0x03 0x00 0x00 0x01 0x00
	var a txn.Argument
	mustDecode(t, []byte{0x03, 0x00, 0x00, 0x01, 0x00}, &a)
	b := encode(t, a)
	if !bytes.Equal(b, []byte{0x03, 0x00, 0x00, 0x01, 0x00}) {
		t.Errorf("NestedResult(0,1) bytes: got %x, want 030000010000", b)
	}
}

func TestDecode_TypeTagBool_FromBytes(t *testing.T) {
	// TypeTag::Bool = ULEB128(0) = 0x00
	var tt txn.TypeTag
	mustDecode(t, []byte{0x00}, &tt)
	b := encode(t, tt)
	if !bytes.Equal(b, []byte{0x00}) {
		t.Errorf("TypeTag::Bool bytes: got %x, want 00", b)
	}
}

func TestDecode_TypeTagU256_FromBytes(t *testing.T) {
	// TypeTag::U256 = ULEB128(10) = 0x0a
	var tt txn.TypeTag
	mustDecode(t, []byte{0x0a}, &tt)
	b := encode(t, tt)
	if !bytes.Equal(b, []byte{0x0a}) {
		t.Errorf("TypeTag::U256 bytes: got %x, want 0a", b)
	}
}

func TestDecode_TypeTagVectorU8_FromBytes(t *testing.T) {
	// TypeTag::Vector(U8) = ULEB128(6) + ULEB128(1) = 0x06 0x01
	var tt txn.TypeTag
	mustDecode(t, []byte{0x06, 0x01}, &tt)
	b := encode(t, tt)
	if !bytes.Equal(b, []byte{0x06, 0x01}) {
		t.Errorf("TypeTag::Vector(U8) bytes: got %x, want 0601", b)
	}
}

func TestDecode_TransactionExpiration_None_FromBytes(t *testing.T) {
	// None = ULEB128(0) = 0x00
	var te txn.TransactionExpiration
	mustDecode(t, []byte{0x00}, &te)
	if te.Epoch != nil {
		t.Error("expected nil Epoch")
	}
}

func TestDecode_TransactionExpiration_Epoch_FromBytes(t *testing.T) {
	// Epoch(42) = ULEB128(1) + U64(42) = 0x01 + 2a00000000000000
	raw := []byte{0x01, 0x2a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	var te txn.TransactionExpiration
	mustDecode(t, raw, &te)
	if te.Epoch == nil {
		t.Fatal("expected non-nil Epoch")
	}
	if *te.Epoch != 42 {
		t.Errorf("expected 42, got %d", *te.Epoch)
	}
}

// ---------------------------------------------------------------------------
// Inner-field decode error paths — truncated mid-stream
// ---------------------------------------------------------------------------

// MoveCall: truncated in type_args element
func TestMoveCall_UnmarshalTruncatedTypeArgElement(t *testing.T) {
	e := bcs.NewEncoder()
	a := addr(t, "0000000000000000000000000000000000000000000000000000000000000002")
	a.MarshalBCS(e)
	mod := bcs.String("pay")
	mod.MarshalBCS(e)
	fn := bcs.String("split")
	fn.MarshalBCS(e)
	e.WriteULEB128(1) // 1 type arg, but the type arg data is missing
	var mc txn.MoveCall
	expectDecodeErr(t, e.Bytes(), &mc)
}

// MoveCall: truncated in args element
func TestMoveCall_UnmarshalTruncatedArgElement(t *testing.T) {
	e := bcs.NewEncoder()
	a := addr(t, "0000000000000000000000000000000000000000000000000000000000000002")
	a.MarshalBCS(e)
	mod := bcs.String("pay")
	mod.MarshalBCS(e)
	fn := bcs.String("split")
	fn.MarshalBCS(e)
	e.WriteULEB128(0)  // 0 type args
	e.WriteULEB128(1)  // 1 arg, but arg data is missing
	var mc txn.MoveCall
	expectDecodeErr(t, e.Bytes(), &mc)
}

// TransferObjects: truncated in objects element
func TestTransferObjects_UnmarshalTruncatedObjectElement(t *testing.T) {
	e := bcs.NewEncoder()
	e.WriteULEB128(1) // 1 object, but object data is missing
	var to txn.TransferObjects
	expectDecodeErr(t, e.Bytes(), &to)
}

// SplitCoins: truncated in amounts element
func TestSplitCoins_UnmarshalTruncatedAmountElement(t *testing.T) {
	e := bcs.NewEncoder()
	txn.GasCoin().MarshalBCS(e) // coin
	e.WriteULEB128(1)            // 1 amount, but amount data is missing
	var sc txn.SplitCoins
	expectDecodeErr(t, e.Bytes(), &sc)
}

// MergeCoins: truncated in sources element
func TestMergeCoins_UnmarshalTruncatedSourceElement(t *testing.T) {
	e := bcs.NewEncoder()
	txn.Input(0).MarshalBCS(e) // destination
	e.WriteULEB128(1)           // 1 source, but source data is missing
	var mc txn.MergeCoins
	expectDecodeErr(t, e.Bytes(), &mc)
}

// GasData: truncated after owner (missing price)
func TestGasData_UnmarshalTruncatedAfterOwner(t *testing.T) {
	e := bcs.NewEncoder()
	e.WriteULEB128(0) // 0 payments
	a := addr(t, "90f3e6d73b5730f16974f4df1d3441394ebae62186baf83608599f226455afa7")
	a.MarshalBCS(e) // owner
	// Missing price, budget
	var g txn.GasData
	expectDecodeErr(t, e.Bytes(), &g)
}

// GasData: truncated after price (missing budget)
func TestGasData_UnmarshalTruncatedAfterPrice(t *testing.T) {
	e := bcs.NewEncoder()
	e.WriteULEB128(0) // 0 payments
	a := addr(t, "90f3e6d73b5730f16974f4df1d3441394ebae62186baf83608599f226455afa7")
	a.MarshalBCS(e) // owner
	p := bcs.U64(1000)
	p.MarshalBCS(e) // price
	// Missing budget
	var g txn.GasData
	expectDecodeErr(t, e.Bytes(), &g)
}

// GasData: truncated inside payment element
func TestGasData_UnmarshalTruncatedPaymentElement(t *testing.T) {
	e := bcs.NewEncoder()
	e.WriteULEB128(1) // 1 payment, but payment data is truncated
	var g txn.GasData
	expectDecodeErr(t, e.Bytes(), &g)
}

// ProgrammableTransaction: truncated inside input element
func TestProgrammableTransaction_UnmarshalTruncatedInputElement(t *testing.T) {
	e := bcs.NewEncoder()
	e.WriteULEB128(1) // 1 input, but input data is missing
	var pt txn.ProgrammableTransaction
	expectDecodeErr(t, e.Bytes(), &pt)
}

// ProgrammableTransaction: truncated inside command element
func TestProgrammableTransaction_UnmarshalTruncatedCommandElement(t *testing.T) {
	e := bcs.NewEncoder()
	e.WriteULEB128(0) // 0 inputs
	e.WriteULEB128(1) // 1 command, but command data is missing
	var pt txn.ProgrammableTransaction
	expectDecodeErr(t, e.Bytes(), &pt)
}

// TransactionDataV1: truncated after kind (missing sender)
func TestTransactionDataV1_UnmarshalTruncatedAfterKind(t *testing.T) {
	// Encode a valid TransactionKind, then stop
	e := bcs.NewEncoder()
	kind := txn.TransactionKind{
		ProgrammableTx: &txn.ProgrammableTransaction{
			Inputs:   []txn.CallArg{},
			Commands: []txn.Command{},
		},
	}
	kind.MarshalBCS(e) // valid kind
	// Missing sender, gas data, expiration
	var v1 txn.TransactionDataV1
	expectDecodeErr(t, e.Bytes(), &v1)
}

// TransactionDataV1: truncated after sender (missing gas data)
func TestTransactionDataV1_UnmarshalTruncatedAfterSender(t *testing.T) {
	e := bcs.NewEncoder()
	kind := txn.TransactionKind{
		ProgrammableTx: &txn.ProgrammableTransaction{
			Inputs:   []txn.CallArg{},
			Commands: []txn.Command{},
		},
	}
	kind.MarshalBCS(e)
	a := addr(t, "90f3e6d73b5730f16974f4df1d3441394ebae62186baf83608599f226455afa7")
	a.MarshalBCS(e) // sender
	// Missing gas data, expiration
	var v1 txn.TransactionDataV1
	expectDecodeErr(t, e.Bytes(), &v1)
}

// TransactionDataV1: truncated after gas data (missing expiration)
func TestTransactionDataV1_UnmarshalTruncatedAfterGasData(t *testing.T) {
	e := bcs.NewEncoder()
	kind := txn.TransactionKind{
		ProgrammableTx: &txn.ProgrammableTransaction{
			Inputs:   []txn.CallArg{},
			Commands: []txn.Command{},
		},
	}
	kind.MarshalBCS(e)
	a := addr(t, "90f3e6d73b5730f16974f4df1d3441394ebae62186baf83608599f226455afa7")
	a.MarshalBCS(e) // sender
	gd := txn.GasData{
		Payment: []txn.ObjectRef{},
		Owner:   a,
		Price:   1000,
		Budget:  5000000,
	}
	gd.MarshalBCS(e) // gas data
	// Missing expiration
	var v1 txn.TransactionDataV1
	expectDecodeErr(t, e.Bytes(), &v1)
}

// TransactionData: truncated V1 data
func TestTransactionData_UnmarshalTruncatedV1(t *testing.T) {
	e := bcs.NewEncoder()
	e.WriteULEB128(0) // V1 tag
	// But no V1 data
	var td txn.TransactionData
	expectDecodeErr(t, e.Bytes(), &td)
}

// TransactionKind: truncated ProgrammableTx
func TestTransactionKind_UnmarshalTruncatedPT(t *testing.T) {
	e := bcs.NewEncoder()
	e.WriteULEB128(0) // ProgrammableTransaction tag
	// But no PT data
	var k txn.TransactionKind
	expectDecodeErr(t, e.Bytes(), &k)
}

// StructTag: truncated after address (missing module)
func TestStructTag_UnmarshalTruncatedAfterAddress(t *testing.T) {
	e := bcs.NewEncoder()
	a := addr(t, "0000000000000000000000000000000000000000000000000000000000000002")
	a.MarshalBCS(e)
	// Missing module, name, type_params
	var s txn.StructTag
	expectDecodeErr(t, e.Bytes(), &s)
}

// StructTag: truncated after module (missing name)
func TestStructTag_UnmarshalTruncatedAfterModule(t *testing.T) {
	e := bcs.NewEncoder()
	a := addr(t, "0000000000000000000000000000000000000000000000000000000000000002")
	a.MarshalBCS(e)
	mod := bcs.String("sui")
	mod.MarshalBCS(e)
	// Missing name, type_params
	var s txn.StructTag
	expectDecodeErr(t, e.Bytes(), &s)
}

// StructTag: truncated after name (missing type_params)
func TestStructTag_UnmarshalTruncatedAfterName(t *testing.T) {
	e := bcs.NewEncoder()
	a := addr(t, "0000000000000000000000000000000000000000000000000000000000000002")
	a.MarshalBCS(e)
	mod := bcs.String("sui")
	mod.MarshalBCS(e)
	name := bcs.String("SUI")
	name.MarshalBCS(e)
	// Missing type_params
	var s txn.StructTag
	expectDecodeErr(t, e.Bytes(), &s)
}

// StructTag: truncated inside type_params element
func TestStructTag_UnmarshalTruncatedTypeParamElement(t *testing.T) {
	e := bcs.NewEncoder()
	a := addr(t, "0000000000000000000000000000000000000000000000000000000000000002")
	a.MarshalBCS(e)
	mod := bcs.String("coin")
	mod.MarshalBCS(e)
	name := bcs.String("Coin")
	name.MarshalBCS(e)
	e.WriteULEB128(1) // 1 type param, but type param data is missing
	var s txn.StructTag
	expectDecodeErr(t, e.Bytes(), &s)
}
