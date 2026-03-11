package tx_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/inodrahq/go-sui-sdk/crypto/ed25519"
	pb "github.com/inodrahq/go-sui-sdk/pb/sui/rpc/v2"
	"github.com/inodrahq/go-sui-sdk/tx"
	"github.com/inodrahq/go-sui-sdk/txn"
)

type mockResolver struct {
	epochResp    *pb.GetEpochResponse
	epochErr     error
	simResp      *pb.SimulateTransactionResponse
	simErr       error
	execResp     *pb.ExecuteTransactionResponse
	execErr      error
	objectResp   *pb.GetObjectResponse
	objectErr    error
}

func (m *mockResolver) GetObject(_ context.Context, _ *pb.GetObjectRequest) (*pb.GetObjectResponse, error) {
	return m.objectResp, m.objectErr
}

func (m *mockResolver) GetEpoch(_ context.Context, _ *pb.GetEpochRequest) (*pb.GetEpochResponse, error) {
	return m.epochResp, m.epochErr
}

func (m *mockResolver) SimulateTransaction(_ context.Context, _ *pb.SimulateTransactionRequest) (*pb.SimulateTransactionResponse, error) {
	return m.simResp, m.simErr
}

func (m *mockResolver) ExecuteTransaction(_ context.Context, _ *pb.ExecuteTransactionRequest) (*pb.ExecuteTransactionResponse, error) {
	return m.execResp, m.execErr
}

func uint64Ptr(v uint64) *uint64 { return &v }

func TestAutoTransactionBuildWithGasPrice(t *testing.T) {
	mock := &mockResolver{
		epochResp: &pb.GetEpochResponse{
			Epoch: &pb.Epoch{
				ReferenceGasPrice: uint64Ptr(1000),
			},
		},
		simResp: &pb.SimulateTransactionResponse{
			Transaction: &pb.ExecutedTransaction{
				Effects: &pb.TransactionEffects{
					GasUsed: &pb.GasCostSummary{
						ComputationCost: uint64Ptr(1000000),
						StorageCost:     uint64Ptr(500000),
						StorageRebate:   uint64Ptr(200000),
					},
				},
			},
		},
	}

	builder := tx.NewAuto(mock)
	builder.SetSender(txn.SuiAddress{})

	amountInput := builder.AddInput(tx.PureU64(1000000000))
	coin := builder.SplitCoins(builder.Gas(), []txn.Argument{amountInput})
	builder.TransferObjects([]txn.Argument{coin}, builder.AddInput(tx.PureAddress(txn.SuiAddress{})))

	txBytes, err := builder.Build(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(txBytes) == 0 {
		t.Error("expected non-empty tx bytes")
	}
}

func TestAutoTransactionBuildWithManualGas(t *testing.T) {
	mock := &mockResolver{}

	builder := tx.NewAuto(mock)
	builder.SetSender(txn.SuiAddress{})
	builder.SetGasPrice(1000)
	builder.SetGasBudget(5000000)
	builder.SetGasData(txn.GasData{
		Payment: []txn.ObjectRef{},
		Owner:   txn.SuiAddress{},
		Price:   1000,
		Budget:  5000000,
	})

	txBytes, err := builder.Build(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(txBytes) == 0 {
		t.Error("expected non-empty tx bytes")
	}
}

func TestAutoTransactionBuildNoSender(t *testing.T) {
	mock := &mockResolver{}
	builder := tx.NewAuto(mock)
	_, err := builder.Build(context.Background())
	if err == nil {
		t.Error("expected error for missing sender")
	}
}

func TestAutoTransactionBuildEpochError(t *testing.T) {
	mock := &mockResolver{
		epochErr: fmt.Errorf("network error"),
	}
	builder := tx.NewAuto(mock)
	builder.SetSender(txn.SuiAddress{})
	_, err := builder.Build(context.Background())
	if err == nil {
		t.Error("expected error from epoch fetch")
	}
}

func TestAutoTransactionBuildSimulateError(t *testing.T) {
	mock := &mockResolver{
		epochResp: &pb.GetEpochResponse{
			Epoch: &pb.Epoch{
				ReferenceGasPrice: uint64Ptr(1000),
			},
		},
		simErr: fmt.Errorf("simulate failed"),
	}
	builder := tx.NewAuto(mock)
	builder.SetSender(txn.SuiAddress{})
	_, err := builder.Build(context.Background())
	if err == nil {
		t.Error("expected error from simulation")
	}
}

func TestAutoTransactionExecute(t *testing.T) {
	digest := "test-digest"
	mock := &mockResolver{
		epochResp: &pb.GetEpochResponse{
			Epoch: &pb.Epoch{
				ReferenceGasPrice: uint64Ptr(1000),
			},
		},
		simResp: &pb.SimulateTransactionResponse{
			Transaction: &pb.ExecutedTransaction{},
		},
		execResp: &pb.ExecuteTransactionResponse{
			Transaction: &pb.ExecutedTransaction{
				Digest: &digest,
			},
		},
	}

	kp, _ := ed25519.New()
	builder := tx.NewAuto(mock)
	builder.SetSender(txn.SuiAddress{})
	builder.SetGasBudget(5000000)

	resp, err := builder.Execute(context.Background(), kp)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Transaction.GetDigest() != "test-digest" {
		t.Error("expected digest from response")
	}
}

func TestAutoTransactionExecuteError(t *testing.T) {
	mock := &mockResolver{
		epochResp: &pb.GetEpochResponse{
			Epoch: &pb.Epoch{
				ReferenceGasPrice: uint64Ptr(1000),
			},
		},
		simResp: &pb.SimulateTransactionResponse{
			Transaction: &pb.ExecutedTransaction{},
		},
		execErr: fmt.Errorf("execution failed"),
	}

	kp, _ := ed25519.New()
	builder := tx.NewAuto(mock)
	builder.SetSender(txn.SuiAddress{})
	builder.SetGasBudget(5000000)

	_, err := builder.Execute(context.Background(), kp)
	if err == nil {
		t.Error("expected error from execution")
	}
}

func TestAutoTransactionGasPriceFallback(t *testing.T) {
	// Epoch response without ReferenceGasPrice
	mock := &mockResolver{
		epochResp: &pb.GetEpochResponse{
			Epoch: &pb.Epoch{},
		},
		simResp: &pb.SimulateTransactionResponse{
			Transaction: &pb.ExecutedTransaction{},
		},
	}

	builder := tx.NewAuto(mock)
	builder.SetSender(txn.SuiAddress{})

	txBytes, err := builder.Build(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(txBytes) == 0 {
		t.Error("expected non-empty tx bytes with fallback gas price")
	}
}

func TestAutoTransactionMinGasBudget(t *testing.T) {
	// Gas used is very small, should be clamped to 2M minimum
	mock := &mockResolver{
		epochResp: &pb.GetEpochResponse{
			Epoch: &pb.Epoch{
				ReferenceGasPrice: uint64Ptr(1000),
			},
		},
		simResp: &pb.SimulateTransactionResponse{
			Transaction: &pb.ExecutedTransaction{
				Effects: &pb.TransactionEffects{
					GasUsed: &pb.GasCostSummary{
						ComputationCost: uint64Ptr(100),
						StorageCost:     uint64Ptr(50),
						StorageRebate:   uint64Ptr(10),
					},
				},
			},
		},
	}

	builder := tx.NewAuto(mock)
	builder.SetSender(txn.SuiAddress{})

	txBytes, err := builder.Build(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(txBytes) == 0 {
		t.Error("expected non-empty tx bytes")
	}
}
