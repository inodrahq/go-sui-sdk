// Package tx provides a high-level transaction builder for Sui.
package tx

import (
	"encoding/base64"
	"fmt"

	"github.com/inodrahq/go-bcs"
	"github.com/inodrahq/go-sui-sdk/crypto"
	"github.com/inodrahq/go-sui-sdk/txn"
)

// Transaction is a high-level builder for Sui programmable transactions.
type Transaction struct {
	sender     *txn.SuiAddress
	gasData    *txn.GasData
	expiration txn.TransactionExpiration
	inputs     []txn.CallArg
	commands   []txn.Command
}

// New creates a new transaction builder.
func New() *Transaction {
	return &Transaction{}
}

// SetSender sets the transaction sender address.
func (t *Transaction) SetSender(addr txn.SuiAddress) {
	t.sender = &addr
}

// SetGasData sets the gas payment data directly.
func (t *Transaction) SetGasData(gas txn.GasData) {
	t.gasData = &gas
}

// SetExpiration sets the transaction expiration epoch.
func (t *Transaction) SetExpiration(epoch uint64) {
	e := bcs.U64(epoch)
	t.expiration = txn.TransactionExpiration{Epoch: &e}
}

// Gas returns an Argument referring to the gas coin.
func (t *Transaction) Gas() txn.Argument {
	return txn.GasCoin()
}

// AddInput adds a CallArg input and returns an Argument referencing it.
func (t *Transaction) AddInput(arg txn.CallArg) txn.Argument {
	idx := len(t.inputs)
	t.inputs = append(t.inputs, arg)
	return txn.Input(uint16(idx))
}

// SplitCoins adds a SplitCoins command and returns the Result argument.
func (t *Transaction) SplitCoins(coin txn.Argument, amounts []txn.Argument) txn.Argument {
	idx := len(t.commands)
	t.commands = append(t.commands, txn.SplitCoinsCommand(coin, amounts))
	return txn.Result(uint16(idx))
}

// MergeCoins adds a MergeCoins command.
func (t *Transaction) MergeCoins(destination txn.Argument, sources []txn.Argument) {
	t.commands = append(t.commands, txn.MergeCoinsCommand(destination, sources))
}

// TransferObjects adds a TransferObjects command.
func (t *Transaction) TransferObjects(objects []txn.Argument, recipient txn.Argument) {
	t.commands = append(t.commands, txn.TransferObjectsCommand(objects, recipient))
}

// MoveCall adds a MoveCall command and returns the Result argument.
func (t *Transaction) MoveCall(call txn.MoveCall) txn.Argument {
	idx := len(t.commands)
	t.commands = append(t.commands, txn.MoveCallCommand(call))
	return txn.Result(uint16(idx))
}

// Publish adds a Publish command and returns the Result argument (UpgradeCap).
func (t *Transaction) Publish(modules [][]byte, dependencies []txn.SuiAddress) txn.Argument {
	idx := len(t.commands)
	t.commands = append(t.commands, txn.PublishCommand(modules, dependencies))
	return txn.Result(uint16(idx))
}

// Build serializes the transaction to BCS bytes.
func (t *Transaction) Build() ([]byte, error) {
	if t.sender == nil {
		return nil, fmt.Errorf("tx: sender not set")
	}
	if t.gasData == nil {
		return nil, fmt.Errorf("tx: gas data not set")
	}

	pt := txn.ProgrammableTransaction{
		Inputs:   t.inputs,
		Commands: t.commands,
	}

	td := txn.TransactionData{
		V1: &txn.TransactionDataV1{
			Kind:       txn.TransactionKind{ProgrammableTx: &pt},
			Sender:     *t.sender,
			GasData:    *t.gasData,
			Expiration: t.expiration,
		},
	}

	e := bcs.NewEncoder()
	if err := td.MarshalBCS(e); err != nil {
		return nil, fmt.Errorf("tx: marshal: %w", err)
	}
	return e.Bytes(), nil
}

// Sign builds the transaction and signs it with the given keypair.
// Returns the base64-encoded Sui signature.
func (t *Transaction) Sign(kp crypto.Keypair) (txBytes []byte, signature string, err error) {
	txBytes, err = t.Build()
	if err != nil {
		return nil, "", err
	}

	sig, err := crypto.SignTransaction(kp, txBytes)
	if err != nil {
		return nil, "", err
	}

	return txBytes, sig, nil
}

// BuildBase64 builds the transaction and returns the BCS bytes as base64.
func (t *Transaction) BuildBase64() (string, error) {
	b, err := t.Build()
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}
