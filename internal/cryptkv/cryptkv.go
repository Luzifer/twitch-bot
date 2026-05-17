// Package cryptkv handles versioned encryption for database key-value secrets.
package cryptkv

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	protoVersionUnknown       protoVersion = iota
	protoVersionLegacyOpenSSL              // OpenSSL AES-256-CBC
	protoVersionV2                         // Argon2ID, AES-256-GCM
)

type (
	protoVersion uint8

	handler interface {
		Decrypt(salt []byte, secret, ciphertext string) (plaintext string, err error)
		Encrypt(salt []byte, secret, plaintext string) (ciphertext string, err error)
	}
)

var (
	currentProto = protoVersionV2
	protos       = map[protoVersion]handler{
		protoVersionLegacyOpenSSL: v1LegacyHandler{},
		protoVersionV2:            v2ArgonAESGCM{},
	}

	errLegacyEncryption = fmt.Errorf("legacy encryption, encrypt no longer supported")
)

// Decrypt decrypts ciphertext produced by supported cryptkv protocols
func Decrypt(salt []byte, secret, ciphertext string) (plaintext string, err error) {
	if strings.HasPrefix(ciphertext, "U2FsdGVkX1") {
		// Legacy migration: Encrypted data was stored in plain OpenSSL
		// compatible format which always has `U2FsdGVkX1` as its prefix due
		// to the `Salted__` prefix. So if we get one of those we convert it
		// transparently into the protoVersionLegacyOpenSSL versioning schema
		ciphertext = fmt.Sprintf("cryptkv%d::%s", protoVersionLegacyOpenSSL, ciphertext)
	}

	protoPrefix, ct, ok := strings.Cut(ciphertext, "::")
	if !ok {
		return "", fmt.Errorf("invalid format found: missing prefix")
	}

	if !strings.HasPrefix(protoPrefix, "cryptkv") {
		return "", fmt.Errorf("invalid crypted data received (missing prefix)")
	}

	protoV, err := strconv.ParseUint(strings.TrimPrefix(protoPrefix, "cryptkv"), 10, 8)
	if err != nil {
		return "", fmt.Errorf("parsing proto-version: %w", err)
	}

	hdl, ok := protos[protoVersion(protoV)]
	if !ok {
		return "", fmt.Errorf("invalid proto version %d", protoV)
	}

	pt, err := hdl.Decrypt(salt, secret, ct)
	if err != nil {
		return "", fmt.Errorf("decrypting secret: %w", err)
	}

	return pt, nil
}

// Encrypt encrypts plaintext using the current cryptkv protocol version.
func Encrypt(salt []byte, secret, plaintext string) (ciphertext string, err error) {
	ct, err := protos[currentProto].Encrypt(salt, secret, plaintext)
	if err != nil {
		return "", fmt.Errorf("encrypting data: %w", err)
	}

	return fmt.Sprintf("cryptkv%d::%s", currentProto, ct), nil
}
