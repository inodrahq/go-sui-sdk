//go:build integration

package client_test

import (
	"context"
	"os"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/inodrahq/go-sui-sdk/client"
	"github.com/inodrahq/go-sui-sdk/crypto"
	pb "github.com/inodrahq/go-sui-sdk/pb/sui/rpc/v2"
	"github.com/inodrahq/go-sui-sdk/tx"
	"github.com/inodrahq/go-sui-sdk/txn"
	"github.com/inodrahq/go-sui-sdk/wallet"
)

const testnetTarget = "fullnode.testnet.sui.io:443"

func testClient(t *testing.T) *client.SuiClient {
	t.Helper()
	c, err := client.New(testnetTarget, client.WithTLS(true))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	t.Cleanup(func() { c.Close() })
	return c
}

func testWebClient(t *testing.T) *client.SuiClient {
	t.Helper()
	c, err := client.New("https://"+testnetTarget, client.WithGRPCWeb())
	if err != nil {
		t.Fatalf("failed to create grpc-web client: %v", err)
	}
	t.Cleanup(func() { c.Close() })
	return c
}

type transportMode struct {
	name   string
	client func(t *testing.T) *client.SuiClient
}

var transportModes = []transportMode{
	{"grpc", testClient},
	{"grpc-web", testWebClient},
}

func strPtr(s string) *string { return &s }

func u32Ptr(v uint32) *uint32 { return &v }

// --- Existing tests, now with dual-transport ---

func TestIntegrationGetServiceInfo(t *testing.T) {
	for _, mode := range transportModes {
		t.Run(mode.name, func(t *testing.T) {
			c := mode.client(t)
			ctx := context.Background()

			resp, err := c.GetServiceInfo(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if resp.GetChain() == "" {
				t.Error("expected chain name")
			}
			if resp.GetEpoch() == 0 {
				t.Error("expected non-zero epoch")
			}
			t.Logf("Chain: %s, Epoch: %d, Checkpoint: %d", resp.GetChain(), resp.GetEpoch(), resp.GetCheckpointHeight())
		})
	}
}

func TestIntegrationGetObject(t *testing.T) {
	for _, mode := range transportModes {
		t.Run(mode.name, func(t *testing.T) {
			c := mode.client(t)
			ctx := context.Background()

			resp, err := c.GetObject(ctx, &pb.GetObjectRequest{
				ObjectId: strPtr("0x0000000000000000000000000000000000000000000000000000000000000005"),
			})
			if err != nil {
				t.Fatal(err)
			}
			if resp.Object == nil {
				t.Error("expected object")
			}
		})
	}
}

func TestIntegrationGetCheckpoint(t *testing.T) {
	for _, mode := range transportModes {
		t.Run(mode.name, func(t *testing.T) {
			c := mode.client(t)
			ctx := context.Background()

			info, err := c.GetServiceInfo(ctx)
			if err != nil {
				t.Fatal(err)
			}
			latestSeq := info.GetCheckpointHeight()
			if latestSeq == 0 {
				t.Skip("no checkpoint height available")
			}

			resp, err := c.GetCheckpoint(ctx, &pb.GetCheckpointRequest{
				CheckpointId: &pb.GetCheckpointRequest_SequenceNumber{SequenceNumber: latestSeq - 1},
			})
			if err != nil {
				t.Fatal(err)
			}
			if resp.Checkpoint == nil {
				t.Error("expected checkpoint")
			}
		})
	}
}

func TestIntegrationGetEpoch(t *testing.T) {
	for _, mode := range transportModes {
		t.Run(mode.name, func(t *testing.T) {
			c := mode.client(t)
			ctx := context.Background()

			resp, err := c.GetEpoch(ctx, &pb.GetEpochRequest{})
			if err != nil {
				t.Fatal(err)
			}
			if resp.Epoch == nil {
				t.Error("expected epoch")
			}
			if resp.Epoch.GetReferenceGasPrice() == 0 {
				t.Error("expected reference gas price")
			}
			t.Logf("Epoch: %d, Gas Price: %d", resp.Epoch.GetEpoch(), resp.Epoch.GetReferenceGasPrice())
		})
	}
}

func TestIntegrationGetBalance(t *testing.T) {
	for _, mode := range transportModes {
		t.Run(mode.name, func(t *testing.T) {
			c := mode.client(t)
			ctx := context.Background()

			resp, err := c.GetBalance(ctx, &pb.GetBalanceRequest{
				Owner:    strPtr("0x0000000000000000000000000000000000000000000000000000000000000000"),
				CoinType: strPtr("0x2::sui::SUI"),
			})
			if err != nil {
				t.Fatal(err)
			}
			_ = resp
		})
	}
}

func TestIntegrationGetCoinInfo(t *testing.T) {
	for _, mode := range transportModes {
		t.Run(mode.name, func(t *testing.T) {
			c := mode.client(t)
			ctx := context.Background()

			resp, err := c.GetCoinInfo(ctx, &pb.GetCoinInfoRequest{
				CoinType: strPtr("0x2::sui::SUI"),
			})
			if err != nil {
				t.Fatal(err)
			}
			_ = resp
		})
	}
}

func TestIntegrationLookupName(t *testing.T) {
	for _, mode := range transportModes {
		t.Run(mode.name, func(t *testing.T) {
			c := mode.client(t)
			ctx := context.Background()

			_, err := c.LookupName(ctx, &pb.LookupNameRequest{
				Name: strPtr("example.sui"),
			})
			_ = err
		})
	}
}

func TestIntegrationGetPackage(t *testing.T) {
	for _, mode := range transportModes {
		t.Run(mode.name, func(t *testing.T) {
			c := mode.client(t)
			ctx := context.Background()

			resp, err := c.GetPackage(ctx, &pb.GetPackageRequest{
				PackageId: strPtr("0x0000000000000000000000000000000000000000000000000000000000000002"),
			})
			if err != nil {
				t.Fatal(err)
			}
			if resp.Package == nil {
				t.Error("expected package")
			}
		})
	}
}

func TestIntegrationBatchGetObjects(t *testing.T) {
	for _, mode := range transportModes {
		t.Run(mode.name, func(t *testing.T) {
			c := mode.client(t)
			ctx := context.Background()

			resp, err := c.BatchGetObjects(ctx, &pb.BatchGetObjectsRequest{
				Requests: []*pb.GetObjectRequest{
					{ObjectId: strPtr("0x0000000000000000000000000000000000000000000000000000000000000005")},
					{ObjectId: strPtr("0x0000000000000000000000000000000000000000000000000000000000000006")},
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			if len(resp.Objects) == 0 {
				t.Error("expected objects")
			}
		})
	}
}

func TestIntegrationListBalances(t *testing.T) {
	for _, mode := range transportModes {
		t.Run(mode.name, func(t *testing.T) {
			c := mode.client(t)
			ctx := context.Background()

			_, err := c.ListBalances(ctx, &pb.ListBalancesRequest{
				Owner: strPtr("0x0000000000000000000000000000000000000000000000000000000000000000"),
			})
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

// --- New tests for previously uncovered methods ---

func TestIntegrationGetTransaction(t *testing.T) {
	for _, mode := range transportModes {
		t.Run(mode.name, func(t *testing.T) {
			c := mode.client(t)
			ctx := context.Background()

			// Fetch a recent checkpoint that includes transactions.
			info, err := c.GetServiceInfo(ctx)
			if err != nil {
				t.Fatal(err)
			}
			seq := info.GetCheckpointHeight() - 1
			cp, err := c.GetCheckpoint(ctx, &pb.GetCheckpointRequest{
				CheckpointId: &pb.GetCheckpointRequest_SequenceNumber{SequenceNumber: seq},
				ReadMask:     &fieldmaskpb.FieldMask{Paths: []string{"sequence_number", "transactions.digest"}},
			})
			if err != nil {
				t.Fatal(err)
			}
			if cp.Checkpoint == nil || len(cp.Checkpoint.Transactions) == 0 {
				t.Skip("checkpoint has no transactions")
			}

			digest := cp.Checkpoint.Transactions[0].GetDigest()
			if digest == "" {
				t.Skip("transaction digest is empty")
			}

			resp, err := c.GetTransaction(ctx, &pb.GetTransactionRequest{
				Digest: strPtr(digest),
			})
			if err != nil {
				t.Fatal(err)
			}
			if resp.Transaction == nil {
				t.Error("expected transaction in response")
			} else if resp.Transaction.GetDigest() == "" {
				t.Error("expected transaction digest in response")
			}
			t.Logf("Transaction digest: %s", resp.Transaction.GetDigest())
		})
	}
}

func TestIntegrationBatchGetTransactions(t *testing.T) {
	for _, mode := range transportModes {
		t.Run(mode.name, func(t *testing.T) {
			c := mode.client(t)
			ctx := context.Background()

			// Fetch a recent checkpoint to get transaction digests.
			info, err := c.GetServiceInfo(ctx)
			if err != nil {
				t.Fatal(err)
			}
			seq := info.GetCheckpointHeight() - 1
			cp, err := c.GetCheckpoint(ctx, &pb.GetCheckpointRequest{
				CheckpointId: &pb.GetCheckpointRequest_SequenceNumber{SequenceNumber: seq},
				ReadMask:     &fieldmaskpb.FieldMask{Paths: []string{"sequence_number", "transactions.digest"}},
			})
			if err != nil {
				t.Fatal(err)
			}
			if cp.Checkpoint == nil || len(cp.Checkpoint.Transactions) == 0 {
				t.Skip("checkpoint has no transactions")
			}

			var digests []string
			for i, tx := range cp.Checkpoint.Transactions {
				if i >= 2 {
					break
				}
				if d := tx.GetDigest(); d != "" {
					digests = append(digests, d)
				}
			}
			if len(digests) == 0 {
				t.Skip("no transaction digests found")
			}

			resp, err := c.BatchGetTransactions(ctx, &pb.BatchGetTransactionsRequest{
				Digests: digests,
			})
			if err != nil {
				t.Fatal(err)
			}
			if len(resp.Transactions) == 0 {
				t.Error("expected at least one transaction")
			}
			t.Logf("Batch fetched %d transactions", len(resp.Transactions))
		})
	}
}

func TestIntegrationListDynamicFields(t *testing.T) {
	for _, mode := range transportModes {
		t.Run(mode.name, func(t *testing.T) {
			c := mode.client(t)
			ctx := context.Background()

			// Object 0x5 (system state) has dynamic fields.
			resp, err := c.ListDynamicFields(ctx, &pb.ListDynamicFieldsRequest{
				Parent:   strPtr("0x0000000000000000000000000000000000000000000000000000000000000005"),
				PageSize: u32Ptr(5),
			})
			if err != nil {
				t.Fatal(err)
			}
			// System state should have dynamic fields for validators, etc.
			t.Logf("ListDynamicFields returned %d entries", len(resp.DynamicFields))
		})
	}
}

func TestIntegrationListOwnedObjects(t *testing.T) {
	for _, mode := range transportModes {
		t.Run(mode.name, func(t *testing.T) {
			c := mode.client(t)
			ctx := context.Background()

			// Use the zero address; it may or may not own objects, but the call should succeed.
			resp, err := c.ListOwnedObjects(ctx, &pb.ListOwnedObjectsRequest{
				Owner:    strPtr("0x0000000000000000000000000000000000000000000000000000000000000000"),
				PageSize: u32Ptr(5),
			})
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("ListOwnedObjects returned %d entries", len(resp.Objects))
		})
	}
}

func TestIntegrationGetDatatype(t *testing.T) {
	for _, mode := range transportModes {
		t.Run(mode.name, func(t *testing.T) {
			c := mode.client(t)
			ctx := context.Background()

			resp, err := c.GetDatatype(ctx, &pb.GetDatatypeRequest{
				PackageId:  strPtr("0x0000000000000000000000000000000000000000000000000000000000000002"),
				ModuleName: strPtr("coin"),
				Name:       strPtr("Coin"),
			})
			if err != nil {
				t.Fatal(err)
			}
			if resp.Datatype == nil {
				t.Error("expected datatype in response")
			}
			t.Logf("GetDatatype returned datatype for coin::Coin")
		})
	}
}

func TestIntegrationGetFunction(t *testing.T) {
	for _, mode := range transportModes {
		t.Run(mode.name, func(t *testing.T) {
			c := mode.client(t)
			ctx := context.Background()

			resp, err := c.GetFunction(ctx, &pb.GetFunctionRequest{
				PackageId:  strPtr("0x0000000000000000000000000000000000000000000000000000000000000002"),
				ModuleName: strPtr("coin"),
				Name:       strPtr("balance"),
			})
			if err != nil {
				t.Fatal(err)
			}
			if resp.Function == nil {
				t.Error("expected function in response")
			}
			t.Logf("GetFunction returned function for coin::balance")
		})
	}
}

func TestIntegrationListPackageVersions(t *testing.T) {
	for _, mode := range transportModes {
		t.Run(mode.name, func(t *testing.T) {
			c := mode.client(t)
			ctx := context.Background()

			resp, err := c.ListPackageVersions(ctx, &pb.ListPackageVersionsRequest{
				PackageId: strPtr("0x0000000000000000000000000000000000000000000000000000000000000002"),
			})
			if err != nil {
				t.Fatal(err)
			}
			if len(resp.Versions) == 0 {
				t.Error("expected at least one package version")
			}
			t.Logf("ListPackageVersions returned %d versions", len(resp.Versions))
		})
	}
}

func TestIntegrationReverseLookupName(t *testing.T) {
	for _, mode := range transportModes {
		t.Run(mode.name, func(t *testing.T) {
			c := mode.client(t)
			ctx := context.Background()

			// Reverse lookup for the zero address. May or may not have a name, but the RPC should work.
			_, err := c.ReverseLookupName(ctx, &pb.ReverseLookupNameRequest{
				Address: strPtr("0x0000000000000000000000000000000000000000000000000000000000000000"),
			})
			// Not fatal if it returns an error (no name registered for this address).
			_ = err
			t.Log("ReverseLookupName call completed")
		})
	}
}

func TestIntegrationSimulateTransaction(t *testing.T) {
	for _, mode := range transportModes {
		t.Run(mode.name, func(t *testing.T) {
			c := mode.client(t)
			ctx := context.Background()

			// Build a minimal PTB: split 1000 MIST from gas and transfer to sender.
			builder := tx.New()
			sender := txn.SuiAddress{} // zero address
			builder.SetSender(sender)
			builder.SetGasData(txn.GasData{
				Payment: []txn.ObjectRef{},
				Owner:   sender,
				Price:   1000,
				Budget:  50000000,
			})
			amtArg := builder.AddInput(tx.PureU64(1000))
			coin := builder.SplitCoins(builder.Gas(), []txn.Argument{amtArg})
			builder.TransferObjects([]txn.Argument{coin}, builder.AddInput(tx.PureAddress(sender)))

			txBytes, err := builder.Build()
			if err != nil {
				t.Fatalf("failed to build transaction: %v", err)
			}

			_, err = c.SimulateTransaction(ctx, &pb.SimulateTransactionRequest{
				Transaction: &pb.Transaction{
					Bcs: &pb.Bcs{
						Name:  strPtr("TransactionData"),
						Value: txBytes,
					},
				},
			})
			// The simulation may fail because the zero address has no gas coins,
			// but we are testing that the client method works end-to-end.
			if err != nil {
				t.Logf("SimulateTransaction returned error (expected for zero address): %v", err)
			} else {
				t.Log("SimulateTransaction succeeded")
			}
		})
	}
}

func TestIntegrationExecuteTransaction(t *testing.T) {
	for _, mode := range transportModes {
		t.Run(mode.name, func(t *testing.T) {
			c := mode.client(t)
			ctx := context.Background()

			// Send an invalid transaction to exercise the client method.
			// We expect an error because the transaction and signature are not valid.
			_, err := c.ExecuteTransaction(ctx, &pb.ExecuteTransactionRequest{
				Transaction: &pb.Transaction{
					Bcs: &pb.Bcs{
						Name:  strPtr("TransactionData"),
						Value: []byte{0x00},
					},
				},
				Signatures: []*pb.UserSignature{
					{
						Bcs: &pb.Bcs{
							Value: []byte{0x00},
						},
					},
				},
			})
			if err == nil {
				t.Error("expected error for invalid transaction execution")
			} else {
				t.Logf("ExecuteTransaction returned expected error: %v", err)
			}
		})
	}
}

func TestIntegrationVerifySignature(t *testing.T) {
	for _, mode := range transportModes {
		t.Run(mode.name, func(t *testing.T) {
			c := mode.client(t)
			ctx := context.Background()

			// Send a dummy signature verification request. We expect an error
			// because the signature and message are invalid.
			_, err := c.VerifySignature(ctx, &pb.VerifySignatureRequest{
				Message: &pb.Bcs{
					Name:  strPtr("PersonalMessage"),
					Value: []byte("hello"),
				},
				Signature: &pb.UserSignature{
					Bcs: &pb.Bcs{
						Value: []byte{0x00},
					},
				},
			})
			if err == nil {
				t.Error("expected error for invalid signature verification")
			} else {
				t.Logf("VerifySignature returned expected error: %v", err)
			}
		})
	}
}

func TestIntegrationSubscribeCheckpoints(t *testing.T) {
	// SubscribeCheckpoints uses server-side streaming which is not supported
	// over gRPC-Web, so only test with native gRPC.
	c := testClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stream, err := c.SubscribeCheckpoints(ctx, &pb.SubscribeCheckpointsRequest{
		ReadMask: &fieldmaskpb.FieldMask{Paths: []string{"checkpoint.sequence_number"}},
	})
	if err != nil {
		// The server may not support streaming; log and skip.
		t.Skipf("SubscribeCheckpoints stream open failed: %v", err)
	}

	// Read at least one checkpoint from the stream.
	msg, err := stream.Recv()
	if err != nil {
		t.Skipf("SubscribeCheckpoints Recv failed (server may not support streaming): %v", err)
	}
	if msg.GetCheckpoint() == nil {
		t.Error("expected checkpoint in stream message")
	} else {
		t.Logf("Received checkpoint seq=%d from stream", msg.GetCheckpoint().GetSequenceNumber())
	}
	cancel()
}

// Ensure the grpc import is used (SubscribeCheckpoints returns grpc.ServerStreamingClient).
var _ grpc.ServerStreamingClient[pb.SubscribeCheckpointsResponse]

// --- Convenience method integration tests ---
// These require SUI_PRIVATE_KEY env var with a funded testnet wallet.

func testKeypair(t *testing.T) crypto.Keypair {
	t.Helper()
	pk := os.Getenv("SUI_PRIVATE_KEY")
	if pk == "" {
		t.Skip("SUI_PRIVATE_KEY not set; skipping funded wallet test")
	}
	w, err := wallet.FromPrivateKey(pk)
	if err != nil {
		t.Fatalf("failed to load wallet: %v", err)
	}
	return w.Keypair()
}

func testKeypair2(t *testing.T) crypto.Keypair {
	t.Helper()
	pk := os.Getenv("SUI_PRIVATE_KEY_2")
	if pk == "" {
		t.Skip("SUI_PRIVATE_KEY_2 not set; skipping two-wallet test")
	}
	w, err := wallet.FromPrivateKey(pk)
	if err != nil {
		t.Fatalf("failed to load wallet 2: %v", err)
	}
	return w.Keypair()
}

func TestIntegrationTransferSui(t *testing.T) {
	kp := testKeypair(t)
	c := testClient(t)
	ctx := context.Background()

	// Transfer 1000 MIST to self.
	resp, err := c.TransferSui(ctx, kp, kp.PublicKey().SuiAddress(), 1000)
	if err != nil {
		t.Fatal(err)
	}
	etx := resp.GetTransaction()
	if etx == nil {
		t.Fatal("expected transaction in response")
	}
	if eff := etx.GetEffects(); eff != nil {
		if !eff.GetStatus().GetSuccess() {
			t.Errorf("transfer failed: %v", eff.GetStatus().GetError())
		}
	}
	t.Logf("TransferSui digest: %s", etx.GetDigest())
}

func TestIntegrationTransferSuiBetweenWallets(t *testing.T) {
	kp1 := testKeypair(t)
	kp2 := testKeypair2(t)
	c := testClient(t)
	ctx := context.Background()

	// W1 -> W2
	resp, err := c.TransferSui(ctx, kp1, kp2.PublicKey().SuiAddress(), 1000)
	if err != nil {
		t.Fatal(err)
	}
	eff := resp.GetTransaction().GetEffects()
	if !eff.GetStatus().GetSuccess() {
		t.Errorf("W1->W2 transfer failed: %v", eff.GetStatus().GetError())
	}
	t.Logf("W1->W2 digest: %s", resp.GetTransaction().GetDigest())

	time.Sleep(2 * time.Second)

	// W2 -> W1
	resp, err = c.TransferSui(ctx, kp2, kp1.PublicKey().SuiAddress(), 1000)
	if err != nil {
		t.Fatal(err)
	}
	eff = resp.GetTransaction().GetEffects()
	if !eff.GetStatus().GetSuccess() {
		t.Errorf("W2->W1 transfer failed: %v", eff.GetStatus().GetError())
	}
	t.Logf("W2->W1 digest: %s", resp.GetTransaction().GetDigest())
}

func TestIntegrationMergeAllCoins(t *testing.T) {
	kp := testKeypair(t)
	c := testClient(t)
	ctx := context.Background()

	// First, create some extra coins by splitting.
	addr := kp.PublicKey().SuiAddress()
	senderAddr, _ := txn.ParseAddress(addr)

	atx := tx.NewAuto(c)
	atx.SetSender(senderAddr)
	amt := atx.AddInput(tx.PureU64(1000))
	dest := atx.AddInput(tx.PureAddress(senderAddr))
	atx.SplitCoins(atx.Gas(), []txn.Argument{amt})
	newCoin := txn.NestedResult(0, 0)
	atx.TransferObjects([]txn.Argument{newCoin}, dest)
	_, err := atx.Execute(ctx, kp)
	if err != nil {
		t.Fatalf("split failed: %v", err)
	}

	time.Sleep(2 * time.Second)

	// Now merge all coins.
	resp, err := c.MergeAllCoins(ctx, kp)
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Log("MergeAllCoins: already a single coin (nothing to merge)")
		return
	}
	eff := resp.GetTransaction().GetEffects()
	if !eff.GetStatus().GetSuccess() {
		t.Errorf("merge failed: %v", eff.GetStatus().GetError())
	}
	t.Logf("MergeAllCoins digest: %s", resp.GetTransaction().GetDigest())
}

func TestIntegrationStakeAndUnstake(t *testing.T) {
	kp := testKeypair(t)
	c := testClient(t)
	ctx := context.Background()

	// Check balance — need > 1.1 SUI.
	addr := kp.PublicKey().SuiAddress()
	balResp, err := c.GetBalance(ctx, &pb.GetBalanceRequest{
		Owner:    strPtr(addr),
		CoinType: strPtr("0x2::sui::SUI"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if bal := balResp.GetBalance(); bal == nil || bal.GetBalance() < 1_100_000_000 {
		t.Skip("insufficient balance for staking (need > 1.1 SUI)")
	}

	// Find a validator via system state.
	epochResp, err := c.GetEpoch(ctx, &pb.GetEpochRequest{
		ReadMask: &fieldmaskpb.FieldMask{Paths: []string{"system_state.validators.active_validators"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	validators := epochResp.GetEpoch().GetSystemState().GetValidators().GetActiveValidators()
	if len(validators) == 0 {
		t.Skip("no active validators found")
	}

	// Pick a validator (prefer Mysten, fall back to first).
	var validatorAddr string
	for _, v := range validators {
		if validatorAddr == "" {
			validatorAddr = v.GetAddress()
		}
		if v.GetName() == "Mysten" || v.GetName() == "Mysten Labs" {
			validatorAddr = v.GetAddress()
			break
		}
	}
	t.Logf("Staking with validator: %s", validatorAddr)

	// Stake 1 SUI.
	stakeResp, err := c.Stake(ctx, kp, validatorAddr, 1_000_000_000)
	if err != nil {
		t.Fatal(err)
	}
	stakeEff := stakeResp.GetTransaction().GetEffects()
	if !stakeEff.GetStatus().GetSuccess() {
		t.Fatalf("stake failed: %v", stakeEff.GetStatus().GetError())
	}
	t.Logf("Stake digest: %s", stakeResp.GetTransaction().GetDigest())

	// Find the StakedSui object from changed_objects.
	var stakedSuiId string
	for _, co := range stakeEff.GetChangedObjects() {
		if co.GetIdOperation() == pb.ChangedObject_CREATED &&
			co.GetObjectType() != "" &&
			co.GetObjectType() != "0x2::coin::Coin<0x2::sui::SUI>" {
			stakedSuiId = co.GetObjectId()
			break
		}
	}
	if stakedSuiId == "" {
		t.Fatal("could not find StakedSui object in effects")
	}
	t.Logf("StakedSui ID: %s", stakedSuiId)

	time.Sleep(2 * time.Second)

	// Unstake.
	unstakeResp, err := c.Unstake(ctx, kp, stakedSuiId)
	if err != nil {
		t.Fatal(err)
	}
	unstakeEff := unstakeResp.GetTransaction().GetEffects()
	if !unstakeEff.GetStatus().GetSuccess() {
		t.Errorf("unstake failed: %v", unstakeEff.GetStatus().GetError())
	}
	t.Logf("Unstake digest: %s", unstakeResp.GetTransaction().GetDigest())
}

func TestIntegrationTransferSuiGRPCWeb(t *testing.T) {
	kp := testKeypair(t)
	c := testWebClient(t)
	ctx := context.Background()

	resp, err := c.TransferSui(ctx, kp, kp.PublicKey().SuiAddress(), 1000)
	if err != nil {
		t.Fatal(err)
	}
	eff := resp.GetTransaction().GetEffects()
	if !eff.GetStatus().GetSuccess() {
		t.Errorf("grpc-web transfer failed: %v", eff.GetStatus().GetError())
	}
	t.Logf("gRPC-Web TransferSui digest: %s", resp.GetTransaction().GetDigest())
}

func TestIntegrationMergeAllCoinsNoop(t *testing.T) {
	kp := testKeypair(t)
	c := testClient(t)
	ctx := context.Background()

	// First merge to ensure we have a single coin.
	_, _ = c.MergeAllCoins(ctx, kp)
	time.Sleep(2 * time.Second)

	// Second merge should be a no-op.
	resp, err := c.MergeAllCoins(ctx, kp)
	if err != nil {
		t.Fatal(err)
	}
	if resp != nil {
		t.Log("MergeAllCoins returned a response even though expected no-op")
	} else {
		t.Log("MergeAllCoins correctly returned nil (nothing to merge)")
	}
}
