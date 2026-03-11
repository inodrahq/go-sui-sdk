package txn_test

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"

	"github.com/inodrahq/go-bcs"
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
	data, err := os.ReadFile("testdata/tx_vectors.json")
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

func TestSimpleTransferVector(t *testing.T) {
	vectors := loadTxVectors(t)
	v := vectors["simple_transfer"]

	senderHex := "90f3e6d73b5730f16974f4df1d3441394ebae62186baf83608599f226455afa7"
	recipientHex := "fd233cd9a5dd7e577f16fa523427c75fbc382af1583c39fdf1c6747d2ed807a3"
	gasObjectID := "0000000000000000000000000000000000000000000000000000000000000005"
	gasDigest := "0000000000000000000000000000000000000000000000000000000000000001"

	// BCS encode amount
	amountEnc := bcs.NewEncoder()
	amt := bcs.U64(1000000000)
	amt.MarshalBCS(amountEnc)

	// BCS encode recipient address
	recipientAddr := mustAddr(t, recipientHex)
	recipientEnc := bcs.NewEncoder()
	recipientAddr.MarshalBCS(recipientEnc)

	pt := txn.ProgrammableTransaction{
		Inputs: []txn.CallArg{
			txn.PureCallArg(amountEnc.Bytes()),
			txn.PureCallArg(recipientEnc.Bytes()),
		},
		Commands: []txn.Command{
			txn.SplitCoinsCommand(txn.GasCoin(), []txn.Argument{txn.Input(0)}),
			txn.TransferObjectsCommand([]txn.Argument{txn.Result(0)}, txn.Input(1)),
		},
	}

	tx := txn.TransactionData{
		V1: &txn.TransactionDataV1{
			Kind:   txn.TransactionKind{ProgrammableTx: &pt},
			Sender: mustAddr(t, senderHex),
			GasData: txn.GasData{
				Payment: []txn.ObjectRef{{
					ObjectID: mustAddr(t, gasObjectID),
					Version:  1,
					Digest:   mustDigest(t, gasDigest),
				}},
				Owner:  mustAddr(t, senderHex),
				Price:  1000,
				Budget: 10000000,
			},
			Expiration: txn.TransactionExpiration{},
		},
	}

	enc := bcs.NewEncoder()
	if err := tx.MarshalBCS(enc); err != nil {
		t.Fatal(err)
	}

	gotHex := hex.EncodeToString(enc.Bytes())
	if gotHex != v.TxHex {
		t.Errorf("simple_transfer hex mismatch:\n  got: %s\n  exp: %s", gotHex, v.TxHex)
	}
}

func TestMoveCallWithTypeArgsVector(t *testing.T) {
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

	pt := txn.ProgrammableTransaction{
		Inputs: []txn.CallArg{
			txn.ObjectCallArg(txn.ImmOrOwnedObject(coinRef)),
			txn.PureCallArg(amountEnc.Bytes()),
		},
		Commands: []txn.Command{
			txn.MoveCallCommand(txn.MoveCall{
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
				Args: []txn.Argument{txn.Input(0), txn.Input(1)},
			}),
		},
	}

	tx := txn.TransactionData{
		V1: &txn.TransactionDataV1{
			Kind:   txn.TransactionKind{ProgrammableTx: &pt},
			Sender: mustAddr(t, senderHex),
			GasData: txn.GasData{
				Payment: []txn.ObjectRef{{
					ObjectID: mustAddr(t, gasObjectID),
					Version:  1,
					Digest:   mustDigest(t, gasDigest),
				}},
				Owner:  mustAddr(t, senderHex),
				Price:  1000,
				Budget: 10000000,
			},
			Expiration: txn.TransactionExpiration{},
		},
	}

	enc := bcs.NewEncoder()
	if err := tx.MarshalBCS(enc); err != nil {
		t.Fatal(err)
	}

	gotHex := hex.EncodeToString(enc.Bytes())
	if gotHex != v.TxHex {
		t.Errorf("move_call hex mismatch:\n  got: %s\n  exp: %s", gotHex, v.TxHex)
	}
}

func TestMergeCoinsVector(t *testing.T) {
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

	pt := txn.ProgrammableTransaction{
		Inputs: []txn.CallArg{
			txn.ObjectCallArg(txn.ImmOrOwnedObject(coin1)),
			txn.ObjectCallArg(txn.ImmOrOwnedObject(coin2)),
		},
		Commands: []txn.Command{
			txn.MergeCoinsCommand(txn.Input(0), []txn.Argument{txn.Input(1)}),
		},
	}

	tx := txn.TransactionData{
		V1: &txn.TransactionDataV1{
			Kind:   txn.TransactionKind{ProgrammableTx: &pt},
			Sender: mustAddr(t, senderHex),
			GasData: txn.GasData{
				Payment: []txn.ObjectRef{{
					ObjectID: mustAddr(t, gasObjectID),
					Version:  1,
					Digest:   mustDigest(t, gasDigest),
				}},
				Owner:  mustAddr(t, senderHex),
				Price:  1000,
				Budget: 5000000,
			},
			Expiration: txn.TransactionExpiration{},
		},
	}

	enc := bcs.NewEncoder()
	if err := tx.MarshalBCS(enc); err != nil {
		t.Fatal(err)
	}

	gotHex := hex.EncodeToString(enc.Bytes())
	if gotHex != v.TxHex {
		t.Errorf("merge_coins hex mismatch:\n  got: %s\n  exp: %s", gotHex, v.TxHex)
	}
}

func TestSharedObjectVector(t *testing.T) {
	vectors := loadTxVectors(t)
	v := vectors["shared_object"]

	senderHex := "90f3e6d73b5730f16974f4df1d3441394ebae62186baf83608599f226455afa7"
	sharedObjID := "7777777777777777777777777777777777777777777777777777777777777777"
	gasObjectID := "0000000000000000000000000000000000000000000000000000000000000005"
	gasDigest := "0000000000000000000000000000000000000000000000000000000000000001"

	amountEnc := bcs.NewEncoder()
	amt := bcs.U64(42)
	amt.MarshalBCS(amountEnc)

	pt := txn.ProgrammableTransaction{
		Inputs: []txn.CallArg{
			txn.ObjectCallArg(txn.SharedObject(txn.SharedObjectRef{
				ObjectID:             mustAddr(t, sharedObjID),
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
	}

	tx := txn.TransactionData{
		V1: &txn.TransactionDataV1{
			Kind:   txn.TransactionKind{ProgrammableTx: &pt},
			Sender: mustAddr(t, senderHex),
			GasData: txn.GasData{
				Payment: []txn.ObjectRef{{
					ObjectID: mustAddr(t, gasObjectID),
					Version:  1,
					Digest:   mustDigest(t, gasDigest),
				}},
				Owner:  mustAddr(t, senderHex),
				Price:  1000,
				Budget: 5000000,
			},
			Expiration: txn.TransactionExpiration{},
		},
	}

	enc := bcs.NewEncoder()
	if err := tx.MarshalBCS(enc); err != nil {
		t.Fatal(err)
	}

	gotHex := hex.EncodeToString(enc.Bytes())
	if gotHex != v.TxHex {
		t.Errorf("shared_object hex mismatch:\n  got: %s\n  exp: %s", gotHex, v.TxHex)
	}
}

func TestAddressParsing(t *testing.T) {
	addr, err := txn.ParseAddress("0x1")
	if err != nil {
		t.Fatal(err)
	}
	if addr.Hex() != "0x0000000000000000000000000000000000000000000000000000000000000001" {
		t.Errorf("unexpected: %s", addr.Hex())
	}
}

func TestAddressParseInvalidHex(t *testing.T) {
	_, err := txn.ParseAddress("0xZZZZ")
	if err == nil {
		t.Error("expected error for invalid hex")
	}
}
