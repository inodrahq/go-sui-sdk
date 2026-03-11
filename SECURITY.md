# Security Policy

## Supported Versions

| Version | Supported          |
|---------|--------------------|
| 1.0.x   | :white_check_mark: |

## Reporting a Vulnerability

If you discover a security vulnerability in this package, please report it responsibly.

**Do not open a public GitHub issue for security vulnerabilities.**

Instead, please email **dimitris@inodra.com** with:

- A description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

You will receive an acknowledgment within 48 hours. We will work with you to understand and address the issue before any public disclosure.

## Scope

This package handles cryptographic key management, transaction signing, and blockchain communication. Security-relevant areas include:

- Private key handling and derivation (Ed25519, Secp256k1, Secp256r1)
- BIP-39 mnemonic generation and validation
- Transaction signing and signature verification
- Multisig signature assembly
- zkLogin proof handling and address derivation
- Passkey authenticator assembly
- gRPC and gRPC-Web transport layer security
- Input validation for all blockchain types
