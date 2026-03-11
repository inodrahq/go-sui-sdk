package crypto

import "errors"

var (
	// ErrInvalidPublicKey is returned when a public key cannot be parsed.
	ErrInvalidPublicKey = errors.New("crypto: invalid public key")

	// ErrInvalidSignature is returned when a signature is malformed.
	ErrInvalidSignature = errors.New("crypto: invalid signature")

	// ErrUnsupportedScheme is returned for unknown signature schemes.
	ErrUnsupportedScheme = errors.New("crypto: unsupported signature scheme")
)
