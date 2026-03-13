package client

import (
	"context"

	pb "github.com/inodrahq/go-sui-sdk/pb/sui/rpc/v2"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// ============================================================================
// LedgerService convenience methods
// ============================================================================

// GetObjectByID retrieves a single object by its ID.
//
//	resp, err := cli.GetObjectByID(ctx, "0x5")
//	resp, err := cli.GetObjectByID(ctx, "0x5", client.WithObjectVersion(42),
//	    client.WithObjectReadMask(client.ReadMask("object_id", "version", "balance")))
func (c *SuiClient) GetObjectByID(ctx context.Context, objectID string, opts ...GetObjectOption) (*pb.GetObjectResponse, error) {
	var o getObjectOpts
	for _, fn := range opts {
		fn(&o)
	}
	return c.GetObject(ctx, &pb.GetObjectRequest{
		ObjectId: strPtr(objectID),
		Version:  o.version,
		ReadMask: o.readMask,
	})
}

// BatchGetObjectsByIDs retrieves multiple objects in a single request.
//
//	resp, err := cli.BatchGetObjectsByIDs(ctx, []string{"0x5", "0x6"})
//	resp, err := cli.BatchGetObjectsByIDs(ctx, ids, client.ReadMask("object_id", "version"))
func (c *SuiClient) BatchGetObjectsByIDs(ctx context.Context, objectIDs []string, readMask ...*fieldmaskpb.FieldMask) (*pb.BatchGetObjectsResponse, error) {
	reqs := make([]*pb.GetObjectRequest, len(objectIDs))
	for i, id := range objectIDs {
		reqs[i] = &pb.GetObjectRequest{ObjectId: strPtr(id)}
	}
	req := &pb.BatchGetObjectsRequest{Requests: reqs}
	if len(readMask) > 0 {
		req.ReadMask = readMask[0]
	}
	return c.BatchGetObjects(ctx, req)
}

// GetTransactionByDigest retrieves a transaction by its digest.
//
//	resp, err := cli.GetTransactionByDigest(ctx, "AbC123...")
//	resp, err := cli.GetTransactionByDigest(ctx, digest,
//	    client.WithTransactionReadMask(client.ReadMask("digest", "effects")))
func (c *SuiClient) GetTransactionByDigest(ctx context.Context, digest string, opts ...GetTransactionOption) (*pb.GetTransactionResponse, error) {
	var o getTransactionOpts
	for _, fn := range opts {
		fn(&o)
	}
	return c.GetTransaction(ctx, &pb.GetTransactionRequest{
		Digest:   strPtr(digest),
		ReadMask: o.readMask,
	})
}

// BatchGetTransactionsByDigests retrieves multiple transactions in a single request.
//
//	resp, err := cli.BatchGetTransactionsByDigests(ctx, []string{"abc", "def"})
//	resp, err := cli.BatchGetTransactionsByDigests(ctx, digests, client.ReadMask("digest", "effects"))
func (c *SuiClient) BatchGetTransactionsByDigests(ctx context.Context, digests []string, readMask ...*fieldmaskpb.FieldMask) (*pb.BatchGetTransactionsResponse, error) {
	req := &pb.BatchGetTransactionsRequest{Digests: digests}
	if len(readMask) > 0 {
		req.ReadMask = readMask[0]
	}
	return c.BatchGetTransactions(ctx, req)
}

// GetCheckpointBySequenceNumber retrieves a checkpoint by its sequence number.
//
//	resp, err := cli.GetCheckpointBySequenceNumber(ctx, 12345)
func (c *SuiClient) GetCheckpointBySequenceNumber(ctx context.Context, seq uint64, opts ...GetCheckpointOption) (*pb.GetCheckpointResponse, error) {
	var o getCheckpointOpts
	for _, fn := range opts {
		fn(&o)
	}
	return c.GetCheckpoint(ctx, &pb.GetCheckpointRequest{
		CheckpointId: &pb.GetCheckpointRequest_SequenceNumber{SequenceNumber: seq},
		ReadMask:      o.readMask,
	})
}

// GetCheckpointByDigest retrieves a checkpoint by its digest.
//
//	resp, err := cli.GetCheckpointByDigest(ctx, "AbC123...")
func (c *SuiClient) GetCheckpointByDigest(ctx context.Context, digest string, opts ...GetCheckpointOption) (*pb.GetCheckpointResponse, error) {
	var o getCheckpointOpts
	for _, fn := range opts {
		fn(&o)
	}
	return c.GetCheckpoint(ctx, &pb.GetCheckpointRequest{
		CheckpointId: &pb.GetCheckpointRequest_Digest{Digest: digest},
		ReadMask:     o.readMask,
	})
}

// GetCurrentEpoch retrieves epoch information. By default returns the current epoch.
//
//	resp, err := cli.GetCurrentEpoch(ctx)
//	resp, err := cli.GetCurrentEpoch(ctx, client.WithEpochNumber(42),
//	    client.WithEpochReadMask(client.ReadMask("epoch", "reference_gas_price")))
func (c *SuiClient) GetCurrentEpoch(ctx context.Context, opts ...GetEpochOption) (*pb.GetEpochResponse, error) {
	var o getEpochOpts
	for _, fn := range opts {
		fn(&o)
	}
	return c.GetEpoch(ctx, &pb.GetEpochRequest{
		Epoch:    o.epoch,
		ReadMask: o.readMask,
	})
}

// ============================================================================
// StateService convenience methods
// ============================================================================

// GetCoinInfoByType retrieves metadata for a coin type.
//
//	resp, err := cli.GetCoinInfoByType(ctx, "0x2::sui::SUI")
func (c *SuiClient) GetCoinInfoByType(ctx context.Context, coinType string) (*pb.GetCoinInfoResponse, error) {
	return c.GetCoinInfo(ctx, &pb.GetCoinInfoRequest{
		CoinType: strPtr(coinType),
	})
}

// GetBalanceByOwner retrieves the balance of a coin type for an address.
//
//	resp, err := cli.GetBalanceByOwner(ctx, "0xABC...", "0x2::sui::SUI")
func (c *SuiClient) GetBalanceByOwner(ctx context.Context, owner, coinType string) (*pb.GetBalanceResponse, error) {
	return c.GetBalance(ctx, &pb.GetBalanceRequest{
		Owner:    strPtr(owner),
		CoinType: strPtr(coinType),
	})
}

// ListBalancesByOwner lists all coin balances for an address.
//
//	resp, err := cli.ListBalancesByOwner(ctx, "0xABC...")
//	resp, err := cli.ListBalancesByOwner(ctx, owner, client.WithPageSize(10))
func (c *SuiClient) ListBalancesByOwner(ctx context.Context, owner string, opts ...ListOption) (*pb.ListBalancesResponse, error) {
	var o listOpts
	for _, fn := range opts {
		fn(&o)
	}
	return c.ListBalances(ctx, &pb.ListBalancesRequest{
		Owner:     strPtr(owner),
		PageSize:  o.pageSize,
		PageToken: o.pageToken,
	})
}

// ListDynamicFieldsByParent lists the dynamic fields of an object.
//
//	resp, err := cli.ListDynamicFieldsByParent(ctx, "0x5")
//	resp, err := cli.ListDynamicFieldsByParent(ctx, parent,
//	    client.WithPageSize(10), client.WithListReadMask(client.ReadMask("field_id")))
func (c *SuiClient) ListDynamicFieldsByParent(ctx context.Context, parent string, opts ...ListOption) (*pb.ListDynamicFieldsResponse, error) {
	var o listOpts
	for _, fn := range opts {
		fn(&o)
	}
	return c.ListDynamicFields(ctx, &pb.ListDynamicFieldsRequest{
		Parent:    strPtr(parent),
		PageSize:  o.pageSize,
		PageToken: o.pageToken,
		ReadMask:  o.readMask,
	})
}

// ListObjectsByOwner lists objects owned by an address.
//
//	resp, err := cli.ListObjectsByOwner(ctx, "0xABC...")
//	resp, err := cli.ListObjectsByOwner(ctx, owner,
//	    client.WithObjectType("0x2::coin::Coin<0x2::sui::SUI>"),
//	    client.WithOwnedObjectsPageSize(10),
//	    client.WithOwnedObjectsReadMask(client.ReadMask("object_id", "balance")))
func (c *SuiClient) ListObjectsByOwner(ctx context.Context, owner string, opts ...ListOwnedObjectsOption) (*pb.ListOwnedObjectsResponse, error) {
	var o listOwnedObjectsOpts
	for _, fn := range opts {
		fn(&o)
	}
	return c.ListOwnedObjects(ctx, &pb.ListOwnedObjectsRequest{
		Owner:      strPtr(owner),
		PageSize:   o.pageSize,
		PageToken:  o.pageToken,
		ReadMask:   o.readMask,
		ObjectType: o.objectType,
	})
}

// ============================================================================
// MovePackageService convenience methods
// ============================================================================

// GetPackageByID retrieves a Move package by its ID.
//
//	resp, err := cli.GetPackageByID(ctx, "0x2")
func (c *SuiClient) GetPackageByID(ctx context.Context, packageID string) (*pb.GetPackageResponse, error) {
	return c.GetPackage(ctx, &pb.GetPackageRequest{
		PackageId: strPtr(packageID),
	})
}

// GetDatatypeByName retrieves a Move struct or enum definition.
//
//	resp, err := cli.GetDatatypeByName(ctx, "0x2", "coin", "Coin")
func (c *SuiClient) GetDatatypeByName(ctx context.Context, packageID, moduleName, name string) (*pb.GetDatatypeResponse, error) {
	return c.GetDatatype(ctx, &pb.GetDatatypeRequest{
		PackageId:  strPtr(packageID),
		ModuleName: strPtr(moduleName),
		Name:       strPtr(name),
	})
}

// GetFunctionByName retrieves a Move function definition.
//
//	resp, err := cli.GetFunctionByName(ctx, "0x2", "coin", "balance")
func (c *SuiClient) GetFunctionByName(ctx context.Context, packageID, moduleName, name string) (*pb.GetFunctionResponse, error) {
	return c.GetFunction(ctx, &pb.GetFunctionRequest{
		PackageId:  strPtr(packageID),
		ModuleName: strPtr(moduleName),
		Name:       strPtr(name),
	})
}

// ListPackageVersionsByID lists all versions of a package.
//
//	resp, err := cli.ListPackageVersionsByID(ctx, "0x2")
//	resp, err := cli.ListPackageVersionsByID(ctx, id, client.WithPackageVersionsPageSize(100))
func (c *SuiClient) ListPackageVersionsByID(ctx context.Context, packageID string, opts ...ListPackageVersionsOption) (*pb.ListPackageVersionsResponse, error) {
	var o listPackageVersionsOpts
	for _, fn := range opts {
		fn(&o)
	}
	return c.ListPackageVersions(ctx, &pb.ListPackageVersionsRequest{
		PackageId: strPtr(packageID),
		PageSize:  o.pageSize,
		PageToken: o.pageToken,
	})
}

// ============================================================================
// NameService convenience methods
// ============================================================================

// ResolveName resolves a Sui name to an address.
//
//	resp, err := cli.ResolveName(ctx, "example.sui")
func (c *SuiClient) ResolveName(ctx context.Context, name string) (*pb.LookupNameResponse, error) {
	return c.LookupName(ctx, &pb.LookupNameRequest{
		Name: strPtr(name),
	})
}

// ResolveAddress resolves an address to its Sui name.
//
//	resp, err := cli.ResolveAddress(ctx, "0xABC...")
func (c *SuiClient) ResolveAddress(ctx context.Context, address string) (*pb.ReverseLookupNameResponse, error) {
	return c.ReverseLookupName(ctx, &pb.ReverseLookupNameRequest{
		Address: strPtr(address),
	})
}
