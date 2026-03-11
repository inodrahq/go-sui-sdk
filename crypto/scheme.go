// Package crypto provides signature scheme definitions and cryptographic interfaces for Sui.
package crypto

// SignatureScheme represents a Sui signature scheme flag byte.
type SignatureScheme byte

const (
	Ed25519Scheme   SignatureScheme = 0x00
	Secp256k1Scheme SignatureScheme = 0x01
	Secp256r1Scheme SignatureScheme = 0x02
	MultiSigScheme  SignatureScheme = 0x03
	ZkLoginScheme   SignatureScheme = 0x05
	PasskeyScheme   SignatureScheme = 0x06
)

// String returns the human-readable name of the signature scheme.
func (s SignatureScheme) String() string {
	switch s {
	case Ed25519Scheme:
		return "Ed25519"
	case Secp256k1Scheme:
		return "Secp256k1"
	case Secp256r1Scheme:
		return "Secp256r1"
	case MultiSigScheme:
		return "MultiSig"
	case ZkLoginScheme:
		return "zkLogin"
	case PasskeyScheme:
		return "Passkey"
	default:
		return "Unknown"
	}
}

// IntentScope identifies the purpose of a signed message.
type IntentScope byte

const (
	IntentTransactionData  IntentScope = 0
	IntentPersonalMessage  IntentScope = 3
)

// IntentPrefix returns the 3-byte intent prefix for a given scope.
// Intent message = intent_prefix || message_bytes
// intent_prefix = [scope, version=0, app_id=0]
func IntentPrefix(scope IntentScope) [3]byte {
	return [3]byte{byte(scope), 0, 0}
}
