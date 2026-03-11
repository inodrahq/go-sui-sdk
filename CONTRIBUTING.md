# Contributing

Contributions are welcome! Here's how to get started.

## Development Setup

```bash
git clone git@github.com:inodrahq/go-sui-sdk.git
cd go-sui-sdk
go mod download
```

This package depends on `inodrahq/go-bcs`. For local development against the BCS source, add a replace directive to your `go.mod`:

```go
replace github.com/inodrahq/go-bcs => ../go-bcs
```

## Running Tests

```bash
go test ./... -race
```

409 tests covering cryptography, key derivation, wallet management, transaction building, multisig, zkLogin, passkey, and gRPC transport.

## Linting

```bash
go vet ./...
```

## Pull Requests

1. Fork the repository and create a branch from `main`
2. Add tests for any new functionality
3. Ensure all tests pass with `-race` enabled
4. Run `go vet ./...` and fix any issues
5. Keep commits focused and write clear commit messages
6. Open a pull request against `main`

## Code Style

- Follow standard Go conventions (`gofmt`, `goimports`)
- Use meaningful variable and function names
- Write godoc comments for all exported types and functions
- Keep packages focused and well-scoped

## Reporting Issues

- Use GitHub Issues for bug reports and feature requests
- For security vulnerabilities, see [SECURITY.md](SECURITY.md)
