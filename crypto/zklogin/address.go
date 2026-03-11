package zklogin

import (
	"fmt"

	"golang.org/x/crypto/blake2b"

	"github.com/inodrahq/go-sui-sdk/crypto"
)

// DeriveAddress derives a zkLogin Sui address from issuer and address seed.
// Blake2b-256(0x05 || len(iss) || iss_bytes || address_seed_bytes[32])
func DeriveAddress(iss, addressSeed string) (string, error) {
	iss = NormalizeIssuer(iss)

	seedBytes, err := AddressSeedToBytes(addressSeed)
	if err != nil {
		return "", err
	}

	data := make([]byte, 0, 1+1+len(iss)+32)
	data = append(data, byte(crypto.ZkLoginScheme))
	data = append(data, byte(len(iss)))
	data = append(data, []byte(iss)...)
	data = append(data, seedBytes...)

	hash := blake2b.Sum256(data)
	return "0x" + fmt.Sprintf("%064x", hash), nil
}

// ComputeAddress computes a zkLogin address from JWT claims and salt.
func ComputeAddress(iss, aud, sub, salt string) (string, error) {
	seed, err := GenAddressSeed(salt, "sub", sub, aud)
	if err != nil {
		return "", err
	}
	return DeriveAddress(iss, seed)
}
