# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2026-03-12

### Added

- **Idiomatic Query API**: 18 convenience methods that hide protobuf details behind plain Go parameters and functional options
  - **Ledger**: `GetObjectByID`, `BatchGetObjectsByIDs`, `GetTransactionByDigest`, `BatchGetTransactionsByDigests`, `GetCheckpointBySequenceNumber`, `GetCheckpointByDigest`, `GetCurrentEpoch`
  - **State**: `GetCoinInfoByType`, `GetBalanceByOwner`, `ListBalancesByOwner`, `ListDynamicFieldsByParent`, `ListObjectsByOwner`
  - **Packages**: `GetPackageByID`, `GetDatatypeByName`, `GetFunctionByName`, `ListPackageVersionsByID`
  - **Names**: `ResolveName`, `ResolveAddress`
- **`ReadMask` helper**: `client.ReadMask("object_id", "version")` replaces `&fieldmaskpb.FieldMask{Paths: []string{...}}`
- **Functional option types**: `GetObjectOption`, `GetTransactionOption`, `GetCheckpointOption`, `GetEpochOption`, `ListOption`, `ListOwnedObjectsOption`, `ListPackageVersionsOption`
- 20 new integration tests for the convenience API
- gRPC query documentation with usage examples in README

### Notes

- All existing raw `pb.*` methods remain unchanged — fully backward compatible

## [1.0.0] - 2026-03-10

### Added

- **6 Signature Schemes**: Ed25519, Secp256k1, Secp256r1, Multisig, zkLogin, Passkey
- **Key Derivation**: BIP-39 mnemonic generation, SLIP-0010 (Ed25519), BIP-32 (ECDSA hardened)
- **Wallet Management**: Create, import, and sign across all signature schemes
- **gRPC Client**: 7 services, 23 methods with functional options (WithTLS, WithAPIKey, WithHeaders)
- **gRPC-Web Transport**: Browser-compatible transport with no C extension needed
- **Transaction Types**: BCS-serializable types, 4 vectors match Rust byte-for-byte
- **Transaction Builder**: Auto-resolving programmable transaction builder (gas price, budget, coin selection)
- **Transaction Commands**: SplitCoins, MergeCoins, TransferObjects, MoveCall
- **Multisig**: Address derivation, bitmap calculation, signature combining
- **zkLogin**: Poseidon hash, address derivation, authenticator assembly
- **Passkey**: Address derivation and authenticator assembly
- 409 tests
- CI pipeline for Go 1.23, 1.24, and 1.25
