package tx_test

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"

	"github.com/inodrahq/go-bcs"
	"github.com/inodrahq/go-sui-sdk/crypto/ed25519"
	"github.com/inodrahq/go-sui-sdk/tx"
	"github.com/inodrahq/go-sui-sdk/txn"
)

type txVector struct {
	TxHex      string `json:"tx_hex"`
	Sender     string `json:"sender"`
	Recipient  string `json:"recipient,omitempty"`
	PrivateKey string `json:"private_key"`
	Signature  string `json:"signature"`
}

func loadTxVectors(t *testing.T) map[string]txVector {
	t.Helper()
	data, err := os.ReadFile("../txn/testdata/tx_vectors.json")
	if err != nil {
		t.Fatalf("failed to load tx vectors: %v", err)
	}
	var vectors map[string]txVector
	if err := json.Unmarshal(data, &vectors); err != nil {
		t.Fatalf("failed to parse tx vectors: %v", err)
	}
	return vectors
}

func mustAddr(t *testing.T, s string) txn.SuiAddress {
	t.Helper()
	addr, err := txn.ParseAddress(s)
	if err != nil {
		t.Fatal(err)
	}
	return addr
}

func mustDigest(t *testing.T, s string) txn.ObjectDigest {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatal(err)
	}
	var d txn.ObjectDigest
	copy(d[:], b)
	return d
}

func TestBuilderSimpleTransfer(t *testing.T) {
	vectors := loadTxVectors(t)
	v := vectors["simple_transfer"]

	senderHex := "90f3e6d73b5730f16974f4df1d3441394ebae62186baf83608599f226455afa7"
	recipientHex := "fd233cd9a5dd7e577f16fa523427c75fbc382af1583c39fdf1c6747d2ed807a3"
	gasObjectID := "0000000000000000000000000000000000000000000000000000000000000005"
	gasDigest := "0000000000000000000000000000000000000000000000000000000000000001"

	builder := tx.New()
	builder.SetSender(mustAddr(t, senderHex))
	builder.SetGasData(txn.GasData{
		Payment: []txn.ObjectRef{{
			ObjectID: mustAddr(t, gasObjectID),
			Version:  1,
			Digest:   mustDigest(t, gasDigest),
		}},
		Owner:  mustAddr(t, senderHex),
		Price:  1000,
		Budget: 10000000,
	})

	amountInput := builder.AddInput(tx.PureU64(1000000000))
	recipientInput := builder.AddInput(tx.PureAddress(mustAddr(t, recipientHex)))

	coin := builder.SplitCoins(builder.Gas(), []txn.Argument{amountInput})
	builder.TransferObjects([]txn.Argument{coin}, recipientInput)

	txBytes, err := builder.Build()
	if err != nil {
		t.Fatal(err)
	}

	gotHex := hex.EncodeToString(txBytes)
	if gotHex != v.TxHex {
		t.Errorf("simple_transfer hex mismatch:\n  got: %s\n  exp: %s", gotHex, v.TxHex)
	}
}

func TestBuilderSign(t *testing.T) {
	vectors := loadTxVectors(t)
	v := vectors["simple_transfer"]

	senderHex := "90f3e6d73b5730f16974f4df1d3441394ebae62186baf83608599f226455afa7"
	recipientHex := "fd233cd9a5dd7e577f16fa523427c75fbc382af1583c39fdf1c6747d2ed807a3"
	gasObjectID := "0000000000000000000000000000000000000000000000000000000000000005"
	gasDigest := "0000000000000000000000000000000000000000000000000000000000000001"

	builder := tx.New()
	builder.SetSender(mustAddr(t, senderHex))
	builder.SetGasData(txn.GasData{
		Payment: []txn.ObjectRef{{
			ObjectID: mustAddr(t, gasObjectID),
			Version:  1,
			Digest:   mustDigest(t, gasDigest),
		}},
		Owner:  mustAddr(t, senderHex),
		Price:  1000,
		Budget: 10000000,
	})

	amountInput := builder.AddInput(tx.PureU64(1000000000))
	recipientInput := builder.AddInput(tx.PureAddress(mustAddr(t, recipientHex)))

	coin := builder.SplitCoins(builder.Gas(), []txn.Argument{amountInput})
	builder.TransferObjects([]txn.Argument{coin}, recipientInput)

	privKeyBytes, _ := hex.DecodeString(v.PrivateKey)
	kp, err := ed25519.FromSeed(privKeyBytes)
	if err != nil {
		t.Fatal(err)
	}

	_, sig, err := builder.Sign(kp)
	if err != nil {
		t.Fatal(err)
	}

	if sig != v.Signature {
		t.Errorf("signature mismatch:\n  got: %s\n  exp: %s", sig, v.Signature)
	}
}

func TestBuilderMergeCoins(t *testing.T) {
	vectors := loadTxVectors(t)
	v := vectors["merge_coins"]

	senderHex := "90f3e6d73b5730f16974f4df1d3441394ebae62186baf83608599f226455afa7"
	gasObjectID := "0000000000000000000000000000000000000000000000000000000000000005"
	gasDigest := "0000000000000000000000000000000000000000000000000000000000000001"

	coin1 := txn.ObjectRef{
		ObjectID: mustAddr(t, "3333333333333333333333333333333333333333333333333333333333333333"),
		Version:  5,
		Digest:   mustDigest(t, "4444444444444444444444444444444444444444444444444444444444444444"),
	}
	coin2 := txn.ObjectRef{
		ObjectID: mustAddr(t, "5555555555555555555555555555555555555555555555555555555555555555"),
		Version:  3,
		Digest:   mustDigest(t, "6666666666666666666666666666666666666666666666666666666666666666"),
	}

	builder := tx.New()
	builder.SetSender(mustAddr(t, senderHex))
	builder.SetGasData(txn.GasData{
		Payment: []txn.ObjectRef{{
			ObjectID: mustAddr(t, gasObjectID),
			Version:  1,
			Digest:   mustDigest(t, gasDigest),
		}},
		Owner:  mustAddr(t, senderHex),
		Price:  1000,
		Budget: 5000000,
	})

	coin1Input := builder.AddInput(tx.ImmOrOwned(coin1))
	coin2Input := builder.AddInput(tx.ImmOrOwned(coin2))
	builder.MergeCoins(coin1Input, []txn.Argument{coin2Input})

	txBytes, err := builder.Build()
	if err != nil {
		t.Fatal(err)
	}

	gotHex := hex.EncodeToString(txBytes)
	if gotHex != v.TxHex {
		t.Errorf("merge_coins hex mismatch:\n  got: %s\n  exp: %s", gotHex, v.TxHex)
	}
}

func TestBuilderMoveCall(t *testing.T) {
	vectors := loadTxVectors(t)
	v := vectors["move_call_with_type_args"]

	senderHex := "90f3e6d73b5730f16974f4df1d3441394ebae62186baf83608599f226455afa7"
	suiFramework := "0000000000000000000000000000000000000000000000000000000000000002"
	gasObjectID := "0000000000000000000000000000000000000000000000000000000000000005"
	gasDigest := "0000000000000000000000000000000000000000000000000000000000000001"

	coinRef := txn.ObjectRef{
		ObjectID: mustAddr(t, "1111111111111111111111111111111111111111111111111111111111111111"),
		Version:  10,
		Digest:   mustDigest(t, "2222222222222222222222222222222222222222222222222222222222222222"),
	}

	amountEnc := bcs.NewEncoder()
	amt := bcs.U64(500000000)
	amt.MarshalBCS(amountEnc)

	builder := tx.New()
	builder.SetSender(mustAddr(t, senderHex))
	builder.SetGasData(txn.GasData{
		Payment: []txn.ObjectRef{{
			ObjectID: mustAddr(t, gasObjectID),
			Version:  1,
			Digest:   mustDigest(t, gasDigest),
		}},
		Owner:  mustAddr(t, senderHex),
		Price:  1000,
		Budget: 10000000,
	})

	coinInput := builder.AddInput(tx.ImmOrOwned(coinRef))
	amountInput := builder.AddInput(tx.PureU64(500000000))

	builder.MoveCall(txn.MoveCall{
		Package:  mustAddr(t, suiFramework),
		Module:   "pay",
		Function: "split",
		TypeArgs: []txn.TypeTag{
			txn.TypeTagStruct(txn.StructTag{
				Address:    mustAddr(t, suiFramework),
				Module:     "sui",
				Name:       "SUI",
				TypeParams: []txn.TypeTag{},
			}),
		},
		Args: []txn.Argument{coinInput, amountInput},
	})

	txBytes, err := builder.Build()
	if err != nil {
		t.Fatal(err)
	}

	gotHex := hex.EncodeToString(txBytes)
	if gotHex != v.TxHex {
		t.Errorf("move_call hex mismatch:\n  got: %s\n  exp: %s", gotHex, v.TxHex)
	}
}

func TestBuilderNoSender(t *testing.T) {
	builder := tx.New()
	builder.SetGasData(txn.GasData{})
	_, err := builder.Build()
	if err == nil {
		t.Error("expected error for missing sender")
	}
}

func TestBuilderNoGasData(t *testing.T) {
	builder := tx.New()
	builder.SetSender(txn.SuiAddress{})
	_, err := builder.Build()
	if err == nil {
		t.Error("expected error for missing gas data")
	}
}

func TestBuilderBuildBase64(t *testing.T) {
	builder := tx.New()
	builder.SetSender(txn.SuiAddress{})
	builder.SetGasData(txn.GasData{
		Payment: []txn.ObjectRef{},
		Owner:   txn.SuiAddress{},
		Price:   1000,
		Budget:  5000000,
	})

	b64, err := builder.BuildBase64()
	if err != nil {
		t.Fatal(err)
	}
	if len(b64) == 0 {
		t.Error("expected non-empty base64")
	}
}

func TestPureHelpers(t *testing.T) {
	// Just verify they don't panic
	_ = tx.PureU8(42)
	_ = tx.PureU16(1000)
	_ = tx.PureU32(100000)
	_ = tx.PureU64(1000000000)
	_ = tx.PureBool(true)
	_ = tx.PureString("hello")
	_ = tx.PureAddress(txn.SuiAddress{})
	_ = tx.PureBytes([]byte{1, 2, 3})
}

func TestSignWithKeypair(t *testing.T) {
	kp, _ := ed25519.New()

	builder := tx.New()
	builder.SetSender(txn.SuiAddress{})
	builder.SetGasData(txn.GasData{
		Payment: []txn.ObjectRef{},
		Owner:   txn.SuiAddress{},
		Price:   1000,
		Budget:  5000000,
	})

	txBytes, sig, err := builder.Sign(kp)
	if err != nil {
		t.Fatal(err)
	}
	if len(txBytes) == 0 {
		t.Error("expected tx bytes")
	}
	if len(sig) == 0 {
		t.Error("expected signature")
	}

	_ = txBytes // verified by checking sig is non-empty
}
