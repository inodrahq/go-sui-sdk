# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
