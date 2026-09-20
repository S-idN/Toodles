package identity

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha512"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"os"
)

const keyFileName = "node_key.pem"

func LoadOrCreateKey(dir string) (*ecdsa.PrivateKey, error) {
	path := dir + string(os.PathSeparator) + keyFileName

	if _, err := os.Stat(path); err == nil {
		return loadKey(path)
	}

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	if err := saveKey(path, key); err != nil {
		return nil, fmt.Errorf("failed to save key: %w", err)
	}

	fmt.Println("Generated new node identity at", path)
	return key, nil
}

func saveKey(path string, key *ecdsa.PrivateKey) error {
	der, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return err
	}

	block := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: der,
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	return pem.Encode(file, block)
}

func loadKey(path string) (*ecdsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block from %s", path)
	}

	key, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse EC private key: %w", err)
	}

	fmt.Println("Loaded existing node identity from", path)
	return key, nil
}

func NodeIDFromKey(key *ecdsa.PrivateKey) string {
	pubBytes := elliptic.MarshalCompressed(key.PublicKey.Curve, key.PublicKey.X, key.PublicKey.Y)
	hash := sha512.Sum512_256(pubBytes)
	return hex.EncodeToString(hash[:])
}
