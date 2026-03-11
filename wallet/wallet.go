// Package wallet provides a high-level wallet abstraction for Sui.
package wallet

import (
	"encoding/base64"
	"fmt"

	"github.com/inodrahq/go-sui-sdk/crypto"
	"github.com/inodrahq/go-sui-sdk/crypto/ed25519"
	"github.com/inodrahq/go-sui-sdk/crypto/mnemonic"
	"github.com/inodrahq/go-sui-sdk/crypto/secp256k1"
	"github.com/inodrahq/go-sui-sdk/crypto/secp256r1"
)

// Wallet wraps a Keypair with high-level signing operations.
type Wallet struct {
	keypair crypto.Keypair
}

// Option configures wallet creation.
type Option func(*walletConfig)

type walletConfig struct {
	scheme crypto.SignatureScheme
	path   string
}

// WithScheme sets the signature scheme (default: Ed25519).
func WithScheme(s crypto.SignatureScheme) Option {
	return func(c *walletConfig) {
		c.scheme = s
	}
}

// WithPath sets the derivation path (default: "m/44'/784'/0'/0'/0'").
func WithPath(path string) Option {
	return func(c *walletConfig) {
		c.path = path
	}
}

func defaultConfig() *walletConfig {
	return &walletConfig{
		scheme: crypto.Ed25519Scheme,
		path:   "m/44'/784'/0'/0'/0'",
	}
}

// New generates a new wallet with a random keypair.
func New(opts ...Option) (*Wallet, error) {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	var kp crypto.Keypair
	var err error
	switch cfg.scheme {
	case crypto.Ed25519Scheme:
		kp, err = ed25519.New()
	case crypto.Secp256k1Scheme:
		kp, err = secp256k1.New()
	case crypto.Secp256r1Scheme:
		kp, err = secp256r1.New()
	default:
		return nil, fmt.Errorf("wallet: unsupported scheme %s for random generation", cfg.scheme)
	}
	if err != nil {
		return nil, err
	}
	return &Wallet{keypair: kp}, nil
}

// FromMnemonic creates a wallet from a BIP-39 mnemonic phrase.
func FromMnemonic(mnemonicPhrase string, opts ...Option) (*Wallet, error) {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	if !mnemonic.Validate(mnemonicPhrase) {
		return nil, fmt.Errorf("wallet: invalid mnemonic")
	}

	seed := mnemonic.ToSeed(mnemonicPhrase, "")

	switch cfg.scheme {
	case crypto.Ed25519Scheme:
		derivedKey, err := mnemonic.DeriveEd25519(seed, cfg.path)
		if err != nil {
			return nil, fmt.Errorf("wallet: derive key: %w", err)
		}
		kp, err := ed25519.FromSeed(derivedKey)
		if err != nil {
			return nil, fmt.Errorf("wallet: create keypair: %w", err)
		}
		return &Wallet{keypair: kp}, nil
	case crypto.Secp256k1Scheme:
		derivedKey, err := mnemonic.DeriveBIP32(seed, cfg.path)
		if err != nil {
			return nil, fmt.Errorf("wallet: derive key: %w", err)
		}
		kp, err := secp256k1.FromPrivateKeyBytes(derivedKey)
		if err != nil {
			return nil, fmt.Errorf("wallet: create keypair: %w", err)
		}
		return &Wallet{keypair: kp}, nil
	case crypto.Secp256r1Scheme:
		derivedKey, err := mnemonic.DeriveBIP32(seed, cfg.path)
		if err != nil {
			return nil, fmt.Errorf("wallet: derive key: %w", err)
		}
		kp, err := secp256r1.FromPrivateKeyBytes(derivedKey)
		if err != nil {
			return nil, fmt.Errorf("wallet: create keypair: %w", err)
		}
		return &Wallet{keypair: kp}, nil
	default:
		return nil, fmt.Errorf("wallet: unsupported scheme %s for mnemonic derivation", cfg.scheme)
	}
}

// FromPrivateKey creates a wallet from a bech32-encoded private key (suiprivkey1...).
// The scheme is auto-detected from the flag byte embedded in the key.
func FromPrivateKey(bech32Key string) (*Wallet, error) {
	hrp, data, err := bech32Decode(bech32Key)
	if err != nil {
		return nil, fmt.Errorf("wallet: decode private key: %w", err)
	}
	if hrp != hrpPrivateKey {
		return nil, fmt.Errorf("wallet: invalid HRP %q, expected %q", hrp, hrpPrivateKey)
	}
	if len(data) < 2 {
		return nil, fmt.Errorf("wallet: decoded key too short")
	}

	flag := crypto.SignatureScheme(data[0])
	seed := data[1:]

	var kp crypto.Keypair
	switch flag {
	case crypto.Ed25519Scheme:
		kp, err = ed25519.FromSeed(seed)
	case crypto.Secp256k1Scheme:
		kp, err = secp256k1.FromPrivateKeyBytes(seed)
	case crypto.Secp256r1Scheme:
		kp, err = secp256r1.FromPrivateKeyBytes(seed)
	default:
		return nil, fmt.Errorf("wallet: unsupported scheme flag 0x%02x", data[0])
	}
	if err != nil {
		return nil, fmt.Errorf("wallet: create keypair: %w", err)
	}

	return &Wallet{keypair: kp}, nil
}

// FromKeypair creates a wallet from an existing keypair.
func FromKeypair(kp crypto.Keypair) *Wallet {
	return &Wallet{keypair: kp}
}

// Address returns the Sui address for this wallet.
func (w *Wallet) Address() string {
	return w.keypair.PublicKey().SuiAddress()
}

// PrivateKey returns the bech32-encoded private key (suiprivkey1...).
func (w *Wallet) PrivateKey() (string, error) {
	seed := w.keypair.Seed()
	data := make([]byte, 1+len(seed))
	data[0] = byte(w.keypair.Scheme())
	copy(data[1:], seed)
	return bech32Encode(hrpPrivateKey, data)
}

// Scheme returns the signature scheme.
func (w *Wallet) Scheme() crypto.SignatureScheme {
	return w.keypair.Scheme()
}

// Keypair returns the underlying keypair.
func (w *Wallet) Keypair() crypto.Keypair {
	return w.keypair
}

// SignTransaction signs transaction bytes and returns a base64-encoded Sui signature.
func (w *Wallet) SignTransaction(txBytes []byte) (string, error) {
	return crypto.SignTransaction(w.keypair, txBytes)
}

// SignPersonalMessage signs a personal message and returns a base64-encoded Sui signature.
func (w *Wallet) SignPersonalMessage(msg []byte) (string, error) {
	return crypto.SignPersonalMessage(w.keypair, msg)
}

// SignRaw signs raw bytes (no intent prefix) and returns the raw Sui signature bytes.
func (w *Wallet) SignRaw(msg []byte) ([]byte, error) {
	rawSig, err := w.keypair.Sign(msg)
	if err != nil {
		return nil, err
	}
	pk := w.keypair.PublicKey()
	pkBytes := pk.Bytes()
	suiSig := make([]byte, 1+len(rawSig)+len(pkBytes))
	suiSig[0] = pk.Flag()
	copy(suiSig[1:], rawSig)
	copy(suiSig[1+len(rawSig):], pkBytes)
	return suiSig, nil
}

// SignRawBase64 signs raw bytes and returns base64-encoded Sui signature.
func (w *Wallet) SignRawBase64(msg []byte) (string, error) {
	sig, err := w.SignRaw(msg)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}
