package client

import (
	"context"
	"fmt"
	"sort"

	"github.com/inodrahq/go-bcs"
	"github.com/inodrahq/go-sui-sdk/crypto"
	"github.com/inodrahq/go-sui-sdk/internal/base58"
	"github.com/inodrahq/go-sui-sdk/tx"
	"github.com/inodrahq/go-sui-sdk/txn"

	pb "github.com/inodrahq/go-sui-sdk/pb/sui/rpc/v2"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// TransferSui transfers SUI from the signer to a recipient.
// Amount is in MIST (1 SUI = 1_000_000_000 MIST).
func (c *SuiClient) TransferSui(ctx context.Context, kp crypto.Keypair, recipient string, amount uint64) (*pb.ExecuteTransactionResponse, error) {
	senderAddr, err := txn.ParseAddress(kp.PublicKey().SuiAddress())
	if err != nil {
		return nil, fmt.Errorf("transferSui: parse sender: %w", err)
	}
	recipientAddr, err := txn.ParseAddress(recipient)
	if err != nil {
		return nil, fmt.Errorf("transferSui: parse recipient: %w", err)
	}

	atx := tx.NewAuto(c)
	atx.SetSender(senderAddr)

	amtArg := atx.AddInput(tx.PureU64(amount))
	destArg := atx.AddInput(tx.PureAddress(recipientAddr))
	atx.SplitCoins(atx.Gas(), []txn.Argument{amtArg})
	newCoin := txn.NestedResult(0, 0)
	atx.TransferObjects([]txn.Argument{newCoin}, destArg)

	return atx.Execute(ctx, kp)
}

// MergeAllCoins merges all SUI coins owned by the signer into a single coin.
// Returns nil response if there is only one coin (nothing to merge).
func (c *SuiClient) MergeAllCoins(ctx context.Context, kp crypto.Keypair) (*pb.ExecuteTransactionResponse, error) {
	address := kp.PublicKey().SuiAddress()
	senderAddr, err := txn.ParseAddress(address)
	if err != nil {
		return nil, fmt.Errorf("mergeAllCoins: parse address: %w", err)
	}

	coins, err := c.listSuiCoins(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("mergeAllCoins: list coins: %w", err)
	}

	if len(coins) <= 1 {
		return nil, nil // nothing to merge
	}

	// Sort by balance descending — largest coin becomes gas (and merge destination).
	sort.Slice(coins, func(i, j int) bool {
		return coins[i].GetBalance() > coins[j].GetBalance()
	})

	atx := tx.NewAuto(c)
	atx.SetSender(senderAddr)

	// All coins except the largest become merge sources.
	var sources []txn.Argument
	for _, coin := range coins[1:] {
		ref, err := objectToRef(coin)
		if err != nil {
			return nil, fmt.Errorf("mergeAllCoins: convert coin ref: %w", err)
		}
		sources = append(sources, atx.AddInput(tx.ImmOrOwned(ref)))
	}

	atx.MergeCoins(atx.Gas(), sources)

	return atx.Execute(ctx, kp)
}

// Stake stakes SUI with a validator via request_add_stake.
// Amount is in MIST. Minimum stake is 1 SUI (1_000_000_000 MIST).
func (c *SuiClient) Stake(ctx context.Context, kp crypto.Keypair, validatorAddress string, amount uint64) (*pb.ExecuteTransactionResponse, error) {
	senderAddr, err := txn.ParseAddress(kp.PublicKey().SuiAddress())
	if err != nil {
		return nil, fmt.Errorf("stake: parse sender: %w", err)
	}
	validatorAddr, err := txn.ParseAddress(validatorAddress)
	if err != nil {
		return nil, fmt.Errorf("stake: parse validator: %w", err)
	}

	atx := tx.NewAuto(c)
	atx.SetSender(senderAddr)

	// Split stake amount from gas coin.
	amtArg := atx.AddInput(tx.PureU64(amount))
	atx.SplitCoins(atx.Gas(), []txn.Argument{amtArg}) // cmd 0
	stakeCoin := txn.NestedResult(0, 0)

	// SuiSystemState shared object at 0x5.
	sysStateAddr, _ := txn.ParseAddress("0x5")
	sysStateArg := atx.AddInput(tx.Shared(txn.SharedObjectRef{
		ObjectID:             sysStateAddr,
		InitialSharedVersion: bcs.U64(1),
		Mutable:              bcs.Bool(true),
	}))

	validatorPure := atx.AddInput(tx.PureAddress(validatorAddr))

	pkg, _ := txn.ParseAddress("0x3")
	atx.MoveCall(txn.MoveCall{
		Package:  pkg,
		Module:   "sui_system",
		Function: "request_add_stake",
		TypeArgs: []txn.TypeTag{},
		Args:     []txn.Argument{sysStateArg, stakeCoin, validatorPure},
	})

	atx.SetGasBudget(50_000_000)
	return atx.Execute(ctx, kp)
}

// Unstake withdraws a previously staked SUI object via request_withdraw_stake.
func (c *SuiClient) Unstake(ctx context.Context, kp crypto.Keypair, stakedSuiId string) (*pb.ExecuteTransactionResponse, error) {
	senderAddr, err := txn.ParseAddress(kp.PublicKey().SuiAddress())
	if err != nil {
		return nil, fmt.Errorf("unstake: parse sender: %w", err)
	}

	// Fetch the StakedSui object reference.
	objResp, err := c.GetObject(ctx, &pb.GetObjectRequest{
		ObjectId: strPtr(stakedSuiId),
		ReadMask: &fieldmaskpb.FieldMask{Paths: []string{"object_id", "version", "digest"}},
	})
	if err != nil {
		return nil, fmt.Errorf("unstake: get staked sui: %w", err)
	}
	obj := objResp.GetObject()
	if obj == nil {
		return nil, fmt.Errorf("unstake: staked sui object not found: %s", stakedSuiId)
	}

	ssRef, err := objectToRef(obj)
	if err != nil {
		return nil, fmt.Errorf("unstake: convert ref: %w", err)
	}

	atx := tx.NewAuto(c)
	atx.SetSender(senderAddr)

	// SuiSystemState shared object at 0x5.
	sysStateAddr, _ := txn.ParseAddress("0x5")
	sysStateArg := atx.AddInput(tx.Shared(txn.SharedObjectRef{
		ObjectID:             sysStateAddr,
		InitialSharedVersion: bcs.U64(1),
		Mutable:              bcs.Bool(true),
	}))

	ssArg := atx.AddInput(tx.ImmOrOwned(ssRef))

	pkg, _ := txn.ParseAddress("0x3")
	atx.MoveCall(txn.MoveCall{
		Package:  pkg,
		Module:   "sui_system",
		Function: "request_withdraw_stake",
		TypeArgs: []txn.TypeTag{},
		Args:     []txn.Argument{sysStateArg, ssArg},
	})

	atx.SetGasBudget(50_000_000)
	return atx.Execute(ctx, kp)
}

// listSuiCoins lists all SUI coin objects owned by the address.
func (c *SuiClient) listSuiCoins(ctx context.Context, address string) ([]*pb.Object, error) {
	var allCoins []*pb.Object
	var pageToken []byte

	for {
		req := &pb.ListOwnedObjectsRequest{
			Owner:      strPtr(address),
			ObjectType: strPtr("0x2::coin::Coin<0x2::sui::SUI>"),
			ReadMask:   &fieldmaskpb.FieldMask{Paths: []string{"object_id", "version", "digest", "balance"}},
		}
		if pageToken != nil {
			req.PageToken = pageToken
		}

		resp, err := c.ListOwnedObjects(ctx, req)
		if err != nil {
			return nil, err
		}

		allCoins = append(allCoins, resp.GetObjects()...)

		if len(resp.GetNextPageToken()) == 0 {
			break
		}
		pageToken = resp.GetNextPageToken()
	}

	return allCoins, nil
}

// objectToRef converts a protobuf Object to a BCS ObjectRef.
func objectToRef(obj *pb.Object) (txn.ObjectRef, error) {
	addr, err := txn.ParseAddress(obj.GetObjectId())
	if err != nil {
		return txn.ObjectRef{}, fmt.Errorf("parse object id: %w", err)
	}

	digestBytes, err := base58.Decode(obj.GetDigest())
	if err != nil {
		return txn.ObjectRef{}, fmt.Errorf("decode digest: %w", err)
	}
	if len(digestBytes) != 32 {
		return txn.ObjectRef{}, fmt.Errorf("digest length %d, expected 32", len(digestBytes))
	}
	var digest txn.ObjectDigest
	copy(digest[:], digestBytes)

	return txn.ObjectRef{
		ObjectID: addr,
		Version:  bcs.U64(obj.GetVersion()),
		Digest:   digest,
	}, nil
}

func strPtr(s string) *string    { return &s }
func uint32Ptr(v uint32) *uint32 { return &v }
func uint64Ptr(v uint64) *uint64 { return &v }
