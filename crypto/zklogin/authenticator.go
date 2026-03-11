package zklogin

import (
	"encoding/base64"
	"fmt"
	"math/big"

	"github.com/inodrahq/go-bcs"
	"github.com/inodrahq/go-sui-sdk/crypto"
)

// ZkLoginProof represents a Groth16 proof with BN254 curve points.
type ZkLoginProof struct {
	A []string   // G1 point [x, y] as decimal strings
	B [][]string // G2 point [[x0, x1], [y0, y1]] as decimal strings
	C []string   // G1 point [x, y] as decimal strings
}

// ZkLoginInputs contains everything needed to assemble a zkLogin authenticator.
type ZkLoginInputs struct {
	Proof            ZkLoginProof
	IssBase64Details string // Base64-encoded issuer details
	HeaderBase64     string // JWT header in base64
	AddressSeed      string // Decimal string
}

// Assemble creates a zkLogin authenticator from the inputs, max epoch, and ephemeral signature.
// Returns a base64-encoded authenticator with the zkLogin flag prefix.
func Assemble(inputs ZkLoginInputs, maxEpoch uint64, userSignature string) (string, error) {
	e := bcs.NewEncoder()

	// Proof: Groth16 (a, b, c)
	if err := encodeG1(e, inputs.Proof.A); err != nil {
		return "", err
	}
	if err := encodeG2(e, inputs.Proof.B); err != nil {
		return "", err
	}
	if err := encodeG1(e, inputs.Proof.C); err != nil {
		return "", err
	}

	// iss_base64_details: Vec<u8>
	issBytes, err := base64.StdEncoding.DecodeString(inputs.IssBase64Details)
	if err != nil {
		// Treat as raw bytes if not valid base64
		issBytes = []byte(inputs.IssBase64Details)
	}
	e.WriteULEB128(uint32(len(issBytes)))
	e.WriteBytes(issBytes)

	// header_base64: BCS string
	e.WriteULEB128(uint32(len(inputs.HeaderBase64)))
	e.WriteBytes([]byte(inputs.HeaderBase64))

	// address_seed: BCS string (decimal)
	e.WriteULEB128(uint32(len(inputs.AddressSeed)))
	e.WriteBytes([]byte(inputs.AddressSeed))

	// maxEpoch: u64 LE
	e.WriteU64(maxEpoch)

	// userSignature: Vec<u8>
	sigBytes, err := base64.StdEncoding.DecodeString(userSignature)
	if err != nil {
		return "", fmt.Errorf("zklogin: invalid base64 user signature: %w", err)
	}
	e.WriteULEB128(uint32(len(sigBytes)))
	e.WriteBytes(sigBytes)

	// Prepend flag
	result := make([]byte, 0, 1+len(e.Bytes()))
	result = append(result, byte(crypto.ZkLoginScheme))
	result = append(result, e.Bytes()...)

	return base64.StdEncoding.EncodeToString(result), nil
}

// encodeG1 encodes a BN254 G1 point (2 field elements) as 2 x 32-byte little-endian.
func encodeG1(e *bcs.Encoder, point []string) error {
	if len(point) < 2 {
		return fmt.Errorf("zklogin: G1 point requires 2 elements")
	}
	for i := 0; i < 2; i++ {
		if err := encodeFieldElement(e, point[i]); err != nil {
			return err
		}
	}
	return nil
}

// encodeG2 encodes a BN254 G2 point (4 field elements) as 4 x 32-byte little-endian.
func encodeG2(e *bcs.Encoder, point [][]string) error {
	if len(point) < 2 || len(point[0]) < 2 || len(point[1]) < 2 {
		return fmt.Errorf("zklogin: G2 point requires [[x0,x1],[y0,y1]]")
	}
	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			if err := encodeFieldElement(e, point[i][j]); err != nil {
				return err
			}
		}
	}
	return nil
}

// encodeFieldElement encodes a decimal string as 32-byte little-endian.
func encodeFieldElement(e *bcs.Encoder, decimal string) error {
	val, ok := new(big.Int).SetString(decimal, 10)
	if !ok {
		return fmt.Errorf("zklogin: invalid field element %q", decimal)
	}
	b := val.Bytes() // big-endian
	if len(b) > 32 {
		return fmt.Errorf("zklogin: field element too large")
	}
	// Convert to 32-byte little-endian
	le := make([]byte, 32)
	for i, v := range b {
		le[len(b)-1-i] = v
	}
	e.WriteBytes(le)
	return nil
}
