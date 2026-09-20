package identity

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha512"
	"crypto/x509"
	"encoding/hex"
	"fmt"
)

func FingerprintFromCert(cert *x509.Certificate) (string, error) {
	pub, ok := cert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("certificate does not contain an ECDSA public key")
	}

	pubBytes := elliptic.MarshalCompressed(pub.Curve, pub.X, pub.Y)
	hash := sha512.Sum512_256(pubBytes)
	return hex.EncodeToString(hash[:]), nil
}

func VerifyPeerID(expectedID string, onVerified func(observedID string)) func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
	return func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
		if len(rawCerts) == 0 {
			return fmt.Errorf("no certificate presented")
		}

		cert, err := x509.ParseCertificate(rawCerts[0])
		if err != nil {
			return fmt.Errorf("failed to parse peer certificate: %w", err)
		}

		observedID, err := FingerprintFromCert(cert)
		if err != nil {
			return err
		}

		if expectedID != "" && observedID != expectedID {
			return fmt.Errorf("peer key mismatch: expected %s, got %s", expectedID, observedID)
		}

		if onVerified != nil {
			onVerified(observedID)
		}
		return nil
	}
}
