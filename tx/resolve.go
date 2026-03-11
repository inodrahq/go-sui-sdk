package tx

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/inodrahq/go-bcs"
	"github.com/inodrahq/go-sui-sdk/crypto"
	"github.com/inodrahq/go-sui-sdk/internal/base58"
	"github.com/inodrahq/go-sui-sdk/txn"

	pb "github.com/inodrahq/go-sui-sdk/pb/sui/rpc/v2"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// Resolver provides methods to auto-resolve transaction parameters via gRPC.
type Resolver interface {
	GetObject(ctx context.Context, req *pb.GetObjectRequest) (*pb.GetObjectResponse, error)
	GetEpoch(ctx context.Context, req *pb.GetEpochRequest) (*pb.GetEpochResponse, error)
	SimulateTransaction(ctx context.Context, req *pb.SimulateTransactionRequest) (*pb.SimulateTransactionResponse, error)
	ExecuteTransaction(ctx context.Context, req *pb.ExecuteTransactionRequest) (*pb.ExecuteTransactionResponse, error)
}

// AutoTransaction is a transaction builder with auto-resolution capabilities.
type AutoTransaction struct {
	Transaction
	client    Resolver
	gasBudget uint64
	gasPrice  uint64
	readMask  *fieldmaskpb.FieldMask
}

// NewAuto creates a new auto-resolving transaction builder.
func NewAuto(client Resolver) *AutoTransaction {
	return &AutoTransaction{
		Transaction: *New(),
		client:      client,
	}
}

// SetGasBudget sets a specific gas budget (skips auto-estimation).
func (t *AutoTransaction) SetGasBudget(budget uint64) {
	t.gasBudget = budget
}

// SetGasPrice sets a specific gas price (skips auto-fetch).
func (t *AutoTransaction) SetGasPrice(price uint64) {
	t.gasPrice = price
}

// SetReadMask sets a custom FieldMask for the ExecuteTransaction response.
// If not set, defaults to digest, effects.status, effects.gas_used, and
// effects.changed_objects.
func (t *AutoTransaction) SetReadMask(mask *fieldmaskpb.FieldMask) {
	t.readMask = mask
}

// Build auto-resolves gas parameters and builds the transaction.
func (t *AutoTransaction) Build(ctx context.Context) ([]byte, error) {
	if t.sender == nil {
		return nil, fmt.Errorf("tx: sender not set")
	}

	// Auto-resolve gas price from current epoch if not set
	gasPrice := t.gasPrice
	if gasPrice == 0 {
		epochResp, err := t.client.GetEpoch(ctx, &pb.GetEpochRequest{})
		if err != nil {
			return nil, fmt.Errorf("tx: get epoch: %w", err)
		}
		if epoch := epochResp.GetEpoch(); epoch != nil {
			gasPrice = epoch.GetReferenceGasPrice()
		}
		if gasPrice == 0 {
			gasPrice = 1000 // fallback
		}
	}

	// If no gas data set, resolve via simulation
	if t.gasData == nil {
		gasBudget := t.gasBudget
		if gasBudget == 0 {
			gasBudget = 50000000 // default 50M MIST, refined by dry-run
		}

		t.gasData = &txn.GasData{
			Payment: []txn.ObjectRef{},
			Owner:   *t.sender,
			Price:   bcs.U64(gasPrice),
			Budget:  bcs.U64(gasBudget),
		}

		// Build initial tx for dry-run
		txBytes, err := t.Transaction.Build()
		if err != nil {
			return nil, fmt.Errorf("tx: build for simulation: %w", err)
		}

		// Simulate to get gas estimate and auto-select gas coins
		simResp, err := t.client.SimulateTransaction(ctx, &pb.SimulateTransactionRequest{
			Transaction: &pb.Transaction{
				Bcs: &pb.Bcs{Value: txBytes},
			},
			DoGasSelection: boolPtr(true),
		})
		if err != nil {
			return nil, fmt.Errorf("tx: simulate: %w", err)
		}

		// Extract gas payment objects from simulation
		if simResp.Transaction != nil {
			if simTx := simResp.Transaction.GetTransaction(); simTx != nil {
				if gp := simTx.GetGasPayment(); gp != nil {
					var payments []txn.ObjectRef
					for _, obj := range gp.GetObjects() {
						ref, err := protoRefToObjectRef(obj)
						if err != nil {
							return nil, fmt.Errorf("tx: parse gas object: %w", err)
						}
						payments = append(payments, ref)
					}
					if len(payments) > 0 {
						t.gasData.Payment = payments
					}
				}
			}

			// Use simulation results to set actual gas budget
			if t.gasBudget == 0 {
				if effects := simResp.Transaction.GetEffects(); effects != nil {
					if gc := effects.GetGasUsed(); gc != nil {
						total := gc.GetComputationCost() + gc.GetStorageCost()
						if rebate := gc.GetStorageRebate(); rebate < total {
							total -= rebate
						}
						// Add 20% buffer
						gasBudget = total + total/5
						if gasBudget < 2000000 {
							gasBudget = 2000000 // minimum 2M MIST
						}
					}
				}
			}
		}

		t.gasData.Price = bcs.U64(gasPrice)
		t.gasData.Budget = bcs.U64(gasBudget)
	}

	return t.Transaction.Build()
}

// Execute builds, signs, and executes the transaction.
func (t *AutoTransaction) Execute(ctx context.Context, kp crypto.Keypair) (*pb.ExecuteTransactionResponse, error) {
	txBytes, err := t.Build(ctx)
	if err != nil {
		return nil, err
	}

	sig, err := crypto.SignTransaction(kp, txBytes)
	if err != nil {
		return nil, fmt.Errorf("tx: sign: %w", err)
	}

	sigBytes, err := base64.StdEncoding.DecodeString(sig)
	if err != nil {
		return nil, fmt.Errorf("tx: decode signature: %w", err)
	}

	mask := t.readMask
	if mask == nil {
		mask = &fieldmaskpb.FieldMask{Paths: []string{
			"digest",
			"effects.status",
			"effects.gas_used",
			"effects.changed_objects",
		}}
	}

	resp, err := t.client.ExecuteTransaction(ctx, &pb.ExecuteTransactionRequest{
		Transaction: &pb.Transaction{
			Bcs: &pb.Bcs{Value: txBytes},
		},
		Signatures: []*pb.UserSignature{
			{Bcs: &pb.Bcs{Value: sigBytes}},
		},
		ReadMask: mask,
	})
	if err != nil {
		return nil, fmt.Errorf("tx: execute: %w", err)
	}

	return resp, nil
}

func boolPtr(b bool) *bool {
	return &b
}

// protoRefToObjectRef converts a protobuf ObjectReference to a BCS ObjectRef.
func protoRefToObjectRef(ref *pb.ObjectReference) (txn.ObjectRef, error) {
	addr, err := txn.ParseAddress(ref.GetObjectId())
	if err != nil {
		return txn.ObjectRef{}, fmt.Errorf("parse object id: %w", err)
	}

	digestBytes, err := base58.Decode(ref.GetDigest())
	if err != nil {
		return txn.ObjectRef{}, fmt.Errorf("decode digest: %w", err)
	}
	var digest txn.ObjectDigest
	if len(digestBytes) != 32 {
		return txn.ObjectRef{}, fmt.Errorf("digest length %d, expected 32", len(digestBytes))
	}
	copy(digest[:], digestBytes)

	return txn.ObjectRef{
		ObjectID: addr,
		Version:  bcs.U64(ref.GetVersion()),
		Digest:   digest,
	}, nil
}
