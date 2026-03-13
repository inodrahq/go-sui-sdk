package client

import "google.golang.org/protobuf/types/known/fieldmaskpb"

// ReadMask builds a FieldMask from the given field paths.
// Use it instead of constructing fieldmaskpb.FieldMask directly.
//
//	cli.GetObjectByID(ctx, "0x5", client.WithObjectReadMask(client.ReadMask("object_id", "version", "balance")))
func ReadMask(paths ...string) *fieldmaskpb.FieldMask {
	return &fieldmaskpb.FieldMask{Paths: paths}
}

// --- GetObject options ---

// GetObjectOption configures a GetObjectByID call.
type GetObjectOption func(*getObjectOpts)

type getObjectOpts struct {
	version  *uint64
	readMask *fieldmaskpb.FieldMask
}

// WithObjectVersion requests a specific version of the object.
func WithObjectVersion(v uint64) GetObjectOption {
	return func(o *getObjectOpts) { o.version = &v }
}

// WithObjectReadMask sets the read mask for the object request.
func WithObjectReadMask(m *fieldmaskpb.FieldMask) GetObjectOption {
	return func(o *getObjectOpts) { o.readMask = m }
}

// --- GetTransaction options ---

// GetTransactionOption configures a GetTransactionByDigest call.
type GetTransactionOption func(*getTransactionOpts)

type getTransactionOpts struct {
	readMask *fieldmaskpb.FieldMask
}

// WithTransactionReadMask sets the read mask for the transaction request.
func WithTransactionReadMask(m *fieldmaskpb.FieldMask) GetTransactionOption {
	return func(o *getTransactionOpts) { o.readMask = m }
}

// --- GetCheckpoint options ---

// GetCheckpointOption configures a GetCheckpointBySequenceNumber or GetCheckpointByDigest call.
type GetCheckpointOption func(*getCheckpointOpts)

type getCheckpointOpts struct {
	readMask *fieldmaskpb.FieldMask
}

// WithCheckpointReadMask sets the read mask for the checkpoint request.
func WithCheckpointReadMask(m *fieldmaskpb.FieldMask) GetCheckpointOption {
	return func(o *getCheckpointOpts) { o.readMask = m }
}

// --- GetEpoch options ---

// GetEpochOption configures a GetCurrentEpoch call.
type GetEpochOption func(*getEpochOpts)

type getEpochOpts struct {
	epoch    *uint64
	readMask *fieldmaskpb.FieldMask
}

// WithEpochNumber requests a specific epoch instead of the current one.
func WithEpochNumber(e uint64) GetEpochOption {
	return func(o *getEpochOpts) { o.epoch = &e }
}

// WithEpochReadMask sets the read mask for the epoch request.
func WithEpochReadMask(m *fieldmaskpb.FieldMask) GetEpochOption {
	return func(o *getEpochOpts) { o.readMask = m }
}

// --- List options (shared by ListBalances, ListDynamicFields) ---

// ListOption configures paginated list calls (ListBalancesByOwner, ListDynamicFieldsByParent).
type ListOption func(*listOpts)

type listOpts struct {
	pageSize  *uint32
	pageToken []byte
	readMask  *fieldmaskpb.FieldMask
}

// WithPageSize sets the maximum number of items to return.
func WithPageSize(n uint32) ListOption {
	return func(o *listOpts) { o.pageSize = &n }
}

// WithPageToken sets the page token for retrieving the next page.
func WithPageToken(token []byte) ListOption {
	return func(o *listOpts) { o.pageToken = token }
}

// WithListReadMask sets the read mask for list requests.
func WithListReadMask(m *fieldmaskpb.FieldMask) ListOption {
	return func(o *listOpts) { o.readMask = m }
}

// --- ListOwnedObjects options ---

// ListOwnedObjectsOption configures a ListObjectsByOwner call.
type ListOwnedObjectsOption func(*listOwnedObjectsOpts)

type listOwnedObjectsOpts struct {
	pageSize   *uint32
	pageToken  []byte
	readMask   *fieldmaskpb.FieldMask
	objectType *string
}

// WithOwnedObjectsPageSize sets the maximum number of objects to return.
func WithOwnedObjectsPageSize(n uint32) ListOwnedObjectsOption {
	return func(o *listOwnedObjectsOpts) { o.pageSize = &n }
}

// WithOwnedObjectsPageToken sets the page token for retrieving the next page.
func WithOwnedObjectsPageToken(token []byte) ListOwnedObjectsOption {
	return func(o *listOwnedObjectsOpts) { o.pageToken = token }
}

// WithOwnedObjectsReadMask sets the read mask for the owned objects request.
func WithOwnedObjectsReadMask(m *fieldmaskpb.FieldMask) ListOwnedObjectsOption {
	return func(o *listOwnedObjectsOpts) { o.readMask = m }
}

// WithObjectType filters owned objects by type (e.g. "0x2::coin::Coin<0x2::sui::SUI>").
func WithObjectType(t string) ListOwnedObjectsOption {
	return func(o *listOwnedObjectsOpts) { o.objectType = &t }
}

// --- ListPackageVersions options ---

// ListPackageVersionsOption configures a ListPackageVersionsByID call.
type ListPackageVersionsOption func(*listPackageVersionsOpts)

type listPackageVersionsOpts struct {
	pageSize  *uint32
	pageToken []byte
}

// WithPackageVersionsPageSize sets the maximum number of versions to return.
func WithPackageVersionsPageSize(n uint32) ListPackageVersionsOption {
	return func(o *listPackageVersionsOpts) { o.pageSize = &n }
}

// WithPackageVersionsPageToken sets the page token for retrieving the next page.
func WithPackageVersionsPageToken(token []byte) ListPackageVersionsOption {
	return func(o *listPackageVersionsOpts) { o.pageToken = token }
}
