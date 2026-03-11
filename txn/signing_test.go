package txn_test

import (
	"encoding/hex"
	"testing"

	"github.com/inodrahq/go-bcs"
	"github.com/inodrahq/go-sui-sdk/crypto"
	"github.com/inodrahq/go-sui-sdk/crypto/ed25519"
	"github.com/inodrahq/go-sui-sdk/txn"
)

func TestSimpleTransferSignature(t *testing.T) {
	vectors := loadTxVectors(t)
	v := vectors["simple_transfer"]

	// Build exact same transaction
	senderHex := "90f3e6d73b5730f16974f4df1d3441394ebae62186baf83608599f226455afa7"
	recipientHex := "fd233cd9a5dd7e577f16fa523427c75fbc382af1583c39fdf1c6747d2ed807a3"
	gasObjectID := "0000000000000000000000000000000000000000000000000000000000000005"
	gasDigest := "0000000000000000000000000000000000000000000000000000000000000001"

	amountEnc := bcs.NewEncoder()
	amt := bcs.U64(1000000000)
	amt.MarshalBCS(amountEnc)

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
	txBytes := enc.Bytes()

	// Sign with the known private key
	privKeyBytes, _ := hex.DecodeString(v.PrivateKey)
	kp, err := ed25519.FromSeed(privKeyBytes)
	if err != nil {
		t.Fatal(err)
	}

	sig, err := crypto.SignTransaction(kp, txBytes)
	if err != nil {
		t.Fatal(err)
	}

	if sig != v.Signature {
		t.Errorf("signature mismatch:\n  got: %s\n  exp: %s", sig, v.Signature)
	}
}
