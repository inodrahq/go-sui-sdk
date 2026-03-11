package crypto

import (
	"encoding/base64"

	"golang.org/x/crypto/blake2b"
)

// SignTransaction signs transaction bytes with intent prefix and returns a Sui signature.
// Returns the serialized Sui signature: flag || raw_sig || pubkey_bytes
func SignTransaction(kp Keypair, txBytes []byte) (string, error) {
	return signWithIntent(kp, txBytes, IntentTransactionData)
}

// SignPersonalMessage signs a personal message with intent prefix and returns a Sui signature.
func SignPersonalMessage(kp Keypair, msg []byte) (string, error) {
	return signWithIntent(kp, msg, IntentPersonalMessage)
}

func signWithIntent(kp Keypair, data []byte, scope IntentScope) (string, error) {
	prefix := IntentPrefix(scope)
	intentMsg := make([]byte, 3+len(data))
	copy(intentMsg, prefix[:])
	copy(intentMsg[3:], data)

	hash := blake2b.Sum256(intentMsg)

	rawSig, err := kp.Sign(hash[:])
	if err != nil {
		return "", err
	}

	pk := kp.PublicKey()
	pkBytes := pk.Bytes()

	// Sui signature format: flag || raw_sig || pubkey
	suiSig := make([]byte, 1+len(rawSig)+len(pkBytes))
	suiSig[0] = pk.Flag()
	copy(suiSig[1:], rawSig)
	copy(suiSig[1+len(rawSig):], pkBytes)

	return base64.StdEncoding.EncodeToString(suiSig), nil
}
