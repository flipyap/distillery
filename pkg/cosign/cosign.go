package cosign

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
)

func ParsePublicKey(pemEncodedPubKey []byte) (*ecdsa.PublicKey, error) {
	block, _ := pem.Decode(pemEncodedPubKey)
	if block == nil || (block.Type != "PUBLIC KEY" && block.Type != "CERTIFICATE") {
		return nil, errors.New("failed to decode PEM block containing public key or certificate")
	}

	var ecdsaPub *ecdsa.PublicKey

	if block.Type == "PUBLIC KEY" {
		pub, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, err
		}

		var ok bool
		ecdsaPub, ok = pub.(*ecdsa.PublicKey)
		if !ok {
			return nil, errors.New("not ECDSA public key")
		}
	} else if block.Type == "CERTIFICATE" {
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}

		var ok bool
		ecdsaPub, ok = cert.PublicKey.(*ecdsa.PublicKey)
		if !ok {
			return nil, errors.New("not ECDSA public key")
		}
	}

	return ecdsaPub, nil
}

// VerifySignature verifies the signature of the data using the provided ECDSA public key.
func VerifySignature(pubKey *ecdsa.PublicKey, data, signature []byte) (bool, error) {
	hash := sha256.Sum256(data)
	fmt.Printf("Data hash: %x\n", hash)

	r, s, err := decodeSignature(signature)
	if err != nil {
		return false, err
	}

	fmt.Printf("r: %s\n", r.String())
	fmt.Printf("s: %s\n", s.String())

	valid := ecdsa.Verify(pubKey, hash[:], r, s)

	return valid, nil
}

// decodeSignature decodes a base64 encoded signature into r and s values.
func decodeSignature(signature []byte) (*big.Int, *big.Int, error) { //nolint:gocritic
	sig, err := base64.StdEncoding.DecodeString(string(signature))
	if err != nil {
		return nil, nil, err
	}

	r := new(big.Int).SetBytes(sig[:len(sig)/2])
	s := new(big.Int).SetBytes(sig[len(sig)/2:])

	return r, s, nil
}
