# go-sui-sdk

Go SDK for [Sui](https://sui.io) - gRPC-native, all 6 signature schemes, auto-resolving transaction builder.

## Install

```
go get github.com/inodrahq/go-sui-sdk
```

Requires Go 1.23+.

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/inodrahq/go-sui-sdk/client"
    "github.com/inodrahq/go-sui-sdk/tx"
    "github.com/inodrahq/go-sui-sdk/txn"
    "github.com/inodrahq/go-sui-sdk/wallet"
    pb "github.com/inodrahq/go-sui-sdk/pb/sui/rpc/v2"
)

func main() {
    ctx := context.Background()

    // Connect to Sui testnet (native gRPC)
    cli, _ := client.New("fullnode.testnet.sui.io:443")
    defer cli.Close()

    // Or use gRPC-Web (HTTP/1.1) — same API, one flag
    // cli, _ := client.New("fullnode.testnet.sui.io:443", client.WithGRPCWeb())

    // Read chain info
    info, _ := cli.GetServiceInfo(ctx)
    fmt.Printf("Chain: %s, Epoch: %d\n", info.GetChain(), info.GetEpoch())

    // Create wallet from mnemonic
    w, _ := wallet.FromMnemonic("your twelve word mnemonic ...")
    fmt.Println("Address:", w.Address())

    // Build and execute a transaction
    addr, _ := txn.ParseAddress(w.Address())
    atx := tx.NewAuto(cli)
    atx.SetSender(addr)

    amt := atx.AddInput(tx.PureU64(1_000_000_000))
    atx.SplitCoins(atx.Gas(), []txn.Argument{amt})         // cmd 0
    coin := txn.NestedResult(0, 0)                          // first coin from split
    atx.TransferObjects([]txn.Argument{coin}, atx.AddInput(tx.PureAddress(addr)))

    resp, _ := atx.Execute(ctx, w.Keypair())
    fmt.Println("Digest:", resp.Transaction.GetDigest())
}
```

## Transport

Both transports expose the same API — switch with a single option:

```go
// Native gRPC (HTTP/2) — default
cli, _ := client.New("fullnode.testnet.sui.io:443")

// gRPC-Web (HTTP/1.1) — for proxies, Cloudflare, restricted environments
cli, _ := client.New("fullnode.testnet.sui.io:443", client.WithGRPCWeb())
```

### Client Options

```go
client.New(target,
    client.WithTLS(true),                          // TLS (default: true)
    client.WithAPIKey("your-key"),                  // x-api-key header
    client.WithHeaders(map[string]string{...}),     // custom headers
    client.WithTimeout(30 * time.Second),           // RPC timeout
    client.WithGRPCWeb(),                           // use gRPC-Web
)
```

## Crypto

All 6 Sui signature schemes:

| Scheme    | Flag   | Package            |
|-----------|--------|--------------------|
| Ed25519   | `0x00` | `crypto/ed25519`   |
| Secp256k1 | `0x01` | `crypto/secp256k1` |
| Secp256r1 | `0x02` | `crypto/secp256r1` |
| MultiSig  | `0x03` | `crypto/multisig`  |
| zkLogin   | `0x05` | `crypto/zklogin`   |
| Passkey   | `0x06` | `crypto/passkey`   |

### Wallets

```go
// Random wallet (default Ed25519)
w, _ := wallet.New()

// From mnemonic with scheme selection
w, _ := wallet.FromMnemonic(mnemonic, wallet.WithScheme(crypto.Secp256k1Scheme))

// From existing keypair
w := wallet.FromKeypair(kp)

// Sign
sig, _ := w.SignTransaction(txBytes)
```

### HD Derivation

```go
// Ed25519: SLIP-0010
seed := mnemonic.ToSeed(phrase, "")
key, _ := mnemonic.DeriveEd25519(seed, "m/44'/784'/0'/0'/0'")

// Secp256k1/r1: BIP-32
key, _ := mnemonic.DeriveBIP32(seed, "m/54'/784'/0'/0/0")
```

## Transaction Builder

### Auto-Resolving (recommended)

Automatically resolves gas price, gas budget (via simulation), and gas coins:

```go
atx := tx.NewAuto(cli)
atx.SetSender(addr)

// Build commands
atx.SplitCoins(atx.Gas(), []txn.Argument{atx.AddInput(tx.PureU64(1000))}) // cmd 0
coin := txn.NestedResult(0, 0)
atx.TransferObjects([]txn.Argument{coin}, atx.AddInput(tx.PureAddress(recipient)))

// Build, sign, and execute in one call
resp, err := atx.Execute(ctx, keypair)
```

### Manual

```go
builder := tx.New()
builder.SetSender(addr)
builder.SetGasData(txn.GasData{...})

// Add commands...
txBytes, err := builder.Build()
sig, err := crypto.SignTransaction(kp, txBytes)
```

### Available Commands

- `SplitCoins(coin, amounts)` — split coin into new coins
- `MergeCoins(destination, sources)` — merge coins
- `TransferObjects(objects, recipient)` — transfer objects
- `MoveCall(call)` — call a Move function
- `Publish(modules, deps)` — publish a Move package

### Convenience Methods

High-level methods on `SuiClient` for common operations:

```go
// Transfer SUI (auto gas resolution)
resp, _ := cli.TransferSui(ctx, keypair, recipientAddr, 1_000_000_000)

// Merge all SUI coins into one
resp, _ := cli.MergeAllCoins(ctx, keypair)

// Stake SUI with a validator
resp, _ := cli.Stake(ctx, keypair, validatorAddr, 1_000_000_000)

// Unstake (withdraw stake)
resp, _ := cli.Unstake(ctx, keypair, stakedSuiObjectId)
```

### Input Helpers

```go
tx.PureU8(v)          tx.PureU16(v)        tx.PureU32(v)
tx.PureU64(v)         tx.PureBool(v)       tx.PureString(v)
tx.PureAddress(addr)  tx.PureBytes(b)

tx.ImmOrOwned(ref)    tx.Shared(ref)       tx.Receiving(ref)
```

## gRPC Methods

All 23 methods across 7 Sui gRPC services:

**Ledger:** GetServiceInfo, GetObject, BatchGetObjects, GetTransaction, BatchGetTransactions, GetCheckpoint, GetEpoch

**State:** ListDynamicFields, ListOwnedObjects, GetCoinInfo, GetBalance, ListBalances

**Execution:** ExecuteTransaction, SimulateTransaction

**Packages:** GetPackage, GetDatatype, GetFunction, ListPackageVersions

**Names:** LookupName, ReverseLookupName

**Subscriptions:** SubscribeCheckpoints

**Verification:** VerifySignature

## Testing

416 unit tests + 29 integration tests (both gRPC and gRPC-Web, plus convenience method tests).

```bash
# Unit tests
go test ./... -race

# Integration tests (hits Sui testnet)
go test -tags integration ./client/ -race
```

## License

MIT
