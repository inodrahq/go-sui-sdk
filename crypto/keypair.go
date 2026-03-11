package crypto

// Keypair represents a signing keypair for any supported Sui signature scheme.
type Keypair interface {
	// Scheme returns the signature scheme of this keypair.
	Scheme() SignatureScheme

	// PublicKey returns the public key associated with this keypair.
	PublicKey() PublicKey

	// Sign signs the raw message bytes (without intent prefix or hashing).
	// The caller is responsible for intent-prefixing and hashing as needed.
	Sign(msg []byte) ([]byte, error)

	// Seed returns the raw private key bytes (32 bytes for all schemes).
	Seed() []byte
}

// PublicKey represents a public key for any supported Sui signature scheme.
type PublicKey interface {
	// Scheme returns the signature scheme of this public key.
	Scheme() SignatureScheme

	// Bytes returns the raw public key bytes (without flag).
	Bytes() []byte

	// SuiAddress returns the Sui address derived from this public key.
	// Format: "0x" + hex(Blake2b256(flag || pubkey_bytes))[:64]
	SuiAddress() string

	// Flag returns the signature scheme flag byte.
	Flag() byte

	// Verify verifies a signature against a message.
	Verify(msg, sig []byte) bool
}
