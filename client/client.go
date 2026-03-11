// Package client provides a gRPC client for the Sui blockchain.
package client

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/inodrahq/go-sui-sdk/pb/sui/rpc/v2"
)

// SuiClient wraps all Sui gRPC services.
type SuiClient struct {
	conn    *grpc.ClientConn   // non-nil for native gRPC
	webConn *grpcWebConn       // non-nil for gRPC-Web
	opts    *options
	ledger  pb.LedgerServiceClient
	state   pb.StateServiceClient
	txExec  pb.TransactionExecutionServiceClient
	pkg     pb.MovePackageServiceClient
	name    pb.NameServiceClient
	sub     pb.SubscriptionServiceClient
	sigVer  pb.SignatureVerificationServiceClient
}

// New creates a new SuiClient connected to the given target.
//
// For native gRPC the target is a host:port (e.g. "fullnode.testnet.sui.io:443").
// When WithGRPCWeb() is used the target should be an HTTP(S) URL
// (e.g. "https://fullnode.testnet.sui.io:443"). If no scheme is provided and
// TLS is enabled, "https://" is prepended automatically.
func New(target string, clientOpts ...Option) (*SuiClient, error) {
	opts := defaultOptions()
	for _, o := range clientOpts {
		o(opts)
	}

	// Build headers map.
	headers := make(map[string]string)
	for k, v := range opts.headers {
		headers[k] = v
	}
	if opts.apiKey != "" {
		headers["x-api-key"] = opts.apiKey
	}

	if opts.grpcWeb {
		return newGRPCWebClient(target, opts, headers)
	}

	return newNativeGRPCClient(target, opts, headers)
}

// newNativeGRPCClient creates a SuiClient using native HTTP/2 gRPC.
func newNativeGRPCClient(target string, opts *options, headers map[string]string) (*SuiClient, error) {
	dialOpts := []grpc.DialOption{
		grpc.WithUnaryInterceptor(headerInterceptor(headers, opts.timeout)),
		grpc.WithStreamInterceptor(headerStreamInterceptor(headers, opts.timeout)),
	}

	if opts.tls {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(nil)))
	} else {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.NewClient(target, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("sui client: dial: %w", err)
	}

	return &SuiClient{
		conn:    conn,
		opts:    opts,
		ledger:  pb.NewLedgerServiceClient(conn),
		state:   pb.NewStateServiceClient(conn),
		txExec:  pb.NewTransactionExecutionServiceClient(conn),
		pkg:     pb.NewMovePackageServiceClient(conn),
		name:    pb.NewNameServiceClient(conn),
		sub:     pb.NewSubscriptionServiceClient(conn),
		sigVer:  pb.NewSignatureVerificationServiceClient(conn),
	}, nil
}

// newGRPCWebClient creates a SuiClient using gRPC-Web over HTTP/1.1.
func newGRPCWebClient(target string, opts *options, headers map[string]string) (*SuiClient, error) {
	// Ensure the target has a scheme.
	if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
		if opts.tls {
			target = "https://" + target
		} else {
			target = "http://" + target
		}
	}

	wc := newGRPCWebConn(target, headers, opts.timeout)

	return &SuiClient{
		webConn: wc,
		opts:    opts,
		ledger:  pb.NewLedgerServiceClient(wc),
		state:   pb.NewStateServiceClient(wc),
		txExec:  pb.NewTransactionExecutionServiceClient(wc),
		pkg:     pb.NewMovePackageServiceClient(wc),
		name:    pb.NewNameServiceClient(wc),
		sub:     pb.NewSubscriptionServiceClient(wc),
		sigVer:  pb.NewSignatureVerificationServiceClient(wc),
	}, nil
}

// Close closes the underlying connection. For gRPC-Web this is a no-op.
func (c *SuiClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// IsGRPCWeb returns true if the client is using the gRPC-Web transport.
func (c *SuiClient) IsGRPCWeb() bool {
	return c.webConn != nil
}


// --- LedgerService ---

// GetServiceInfo returns chain identification and current state information.
func (c *SuiClient) GetServiceInfo(ctx context.Context) (*pb.GetServiceInfoResponse, error) {
	return c.ledger.GetServiceInfo(ctx, &pb.GetServiceInfoRequest{})
}

// GetObject retrieves a single object by its ID.
func (c *SuiClient) GetObject(ctx context.Context, req *pb.GetObjectRequest) (*pb.GetObjectResponse, error) {
	return c.ledger.GetObject(ctx, req)
}

// BatchGetObjects retrieves multiple objects in a single request.
func (c *SuiClient) BatchGetObjects(ctx context.Context, req *pb.BatchGetObjectsRequest) (*pb.BatchGetObjectsResponse, error) {
	return c.ledger.BatchGetObjects(ctx, req)
}

// GetTransaction retrieves a transaction by its digest.
func (c *SuiClient) GetTransaction(ctx context.Context, req *pb.GetTransactionRequest) (*pb.GetTransactionResponse, error) {
	return c.ledger.GetTransaction(ctx, req)
}

// BatchGetTransactions retrieves multiple transactions in a single request.
func (c *SuiClient) BatchGetTransactions(ctx context.Context, req *pb.BatchGetTransactionsRequest) (*pb.BatchGetTransactionsResponse, error) {
	return c.ledger.BatchGetTransactions(ctx, req)
}

// GetCheckpoint retrieves a checkpoint by sequence number or digest.
func (c *SuiClient) GetCheckpoint(ctx context.Context, req *pb.GetCheckpointRequest) (*pb.GetCheckpointResponse, error) {
	return c.ledger.GetCheckpoint(ctx, req)
}

// GetEpoch retrieves epoch information.
func (c *SuiClient) GetEpoch(ctx context.Context, req *pb.GetEpochRequest) (*pb.GetEpochResponse, error) {
	return c.ledger.GetEpoch(ctx, req)
}

// --- StateService ---

// ListDynamicFields lists the dynamic fields of an object.
func (c *SuiClient) ListDynamicFields(ctx context.Context, req *pb.ListDynamicFieldsRequest) (*pb.ListDynamicFieldsResponse, error) {
	return c.state.ListDynamicFields(ctx, req)
}

// ListOwnedObjects lists objects owned by an address.
func (c *SuiClient) ListOwnedObjects(ctx context.Context, req *pb.ListOwnedObjectsRequest) (*pb.ListOwnedObjectsResponse, error) {
	return c.state.ListOwnedObjects(ctx, req)
}

// GetCoinInfo retrieves metadata for a coin type.
func (c *SuiClient) GetCoinInfo(ctx context.Context, req *pb.GetCoinInfoRequest) (*pb.GetCoinInfoResponse, error) {
	return c.state.GetCoinInfo(ctx, req)
}

// GetBalance retrieves the balance of a coin type for an address.
func (c *SuiClient) GetBalance(ctx context.Context, req *pb.GetBalanceRequest) (*pb.GetBalanceResponse, error) {
	return c.state.GetBalance(ctx, req)
}

// ListBalances lists all coin balances for an address.
func (c *SuiClient) ListBalances(ctx context.Context, req *pb.ListBalancesRequest) (*pb.ListBalancesResponse, error) {
	return c.state.ListBalances(ctx, req)
}

// --- TransactionExecutionService ---

// ExecuteTransaction submits a signed transaction for execution.
func (c *SuiClient) ExecuteTransaction(ctx context.Context, req *pb.ExecuteTransactionRequest) (*pb.ExecuteTransactionResponse, error) {
	return c.txExec.ExecuteTransaction(ctx, req)
}

// SimulateTransaction dry-runs a transaction without executing it.
func (c *SuiClient) SimulateTransaction(ctx context.Context, req *pb.SimulateTransactionRequest) (*pb.SimulateTransactionResponse, error) {
	return c.txExec.SimulateTransaction(ctx, req)
}

// --- MovePackageService ---

// GetPackage retrieves a Move package by its ID.
func (c *SuiClient) GetPackage(ctx context.Context, req *pb.GetPackageRequest) (*pb.GetPackageResponse, error) {
	return c.pkg.GetPackage(ctx, req)
}

// GetDatatype retrieves a Move struct or enum definition.
func (c *SuiClient) GetDatatype(ctx context.Context, req *pb.GetDatatypeRequest) (*pb.GetDatatypeResponse, error) {
	return c.pkg.GetDatatype(ctx, req)
}

// GetFunction retrieves a Move function definition.
func (c *SuiClient) GetFunction(ctx context.Context, req *pb.GetFunctionRequest) (*pb.GetFunctionResponse, error) {
	return c.pkg.GetFunction(ctx, req)
}

// ListPackageVersions lists all versions of a package.
func (c *SuiClient) ListPackageVersions(ctx context.Context, req *pb.ListPackageVersionsRequest) (*pb.ListPackageVersionsResponse, error) {
	return c.pkg.ListPackageVersions(ctx, req)
}

// --- NameService ---

// LookupName resolves a Sui name to an address.
func (c *SuiClient) LookupName(ctx context.Context, req *pb.LookupNameRequest) (*pb.LookupNameResponse, error) {
	return c.name.LookupName(ctx, req)
}

// ReverseLookupName resolves an address to its Sui name.
func (c *SuiClient) ReverseLookupName(ctx context.Context, req *pb.ReverseLookupNameRequest) (*pb.ReverseLookupNameResponse, error) {
	return c.name.ReverseLookupName(ctx, req)
}

// --- SubscriptionService ---

// SubscribeCheckpoints opens a server-streaming subscription for new checkpoints.
func (c *SuiClient) SubscribeCheckpoints(ctx context.Context, req *pb.SubscribeCheckpointsRequest) (grpc.ServerStreamingClient[pb.SubscribeCheckpointsResponse], error) {
	return c.sub.SubscribeCheckpoints(ctx, req)
}

// --- SignatureVerificationService ---

// VerifySignature verifies a transaction signature on the server side.
func (c *SuiClient) VerifySignature(ctx context.Context, req *pb.VerifySignatureRequest) (*pb.VerifySignatureResponse, error) {
	return c.sigVer.VerifySignature(ctx, req)
}
