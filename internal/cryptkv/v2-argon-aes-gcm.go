package cryptkv

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"

	"golang.org/x/crypto/argon2"
)

const (
	// Recommendations from RFC 9106 Section 7.4
	v2ArgonTime    = 3
	v2ArgonMemory  = 64 * 1024 // 64 MiB
	v2ArgonThreads = 2

	v2Keylen = 32
)

type (
	v2ArgonAESGCM struct{}
)

var v2KeyExchangeLock sync.Mutex

func (v v2ArgonAESGCM) Decrypt(salt []byte, secret, ciphertext string) (plaintext string, err error) {
	c, err := v.getCipher(salt, secret)
	if err != nil {
		return "", fmt.Errorf("getting cipher: %w", err)
	}

	raw, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("decoding base64 data: %w", err)
	}

	if len(raw) < c.NonceSize() {
		return "", fmt.Errorf("invalid data after base64 decode, too short")
	}

	plainRaw, err := c.Open(nil, raw[:c.NonceSize()], raw[c.NonceSize():], nil)
	if err != nil {
		return "", fmt.Errorf("decrypting data: %w", err)
	}

	return string(plainRaw), nil
}

func (v v2ArgonAESGCM) Encrypt(salt []byte, secret, plaintext string) (ciphertext string, err error) {
	c, err := v.getCipher(salt, secret)
	if err != nil {
		return "", fmt.Errorf("getting cipher: %w", err)
	}

	nonce := make([]byte, c.NonceSize())
	n, err := rand.Read(nonce)
	if err != nil {
		return "", fmt.Errorf("reading random nonce: %w", err)
	}
	if n != c.NonceSize() {
		return "", fmt.Errorf("read invalid nonce size: %d != %d", n, c.NonceSize())
	}

	raw := c.Seal(nil, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(append(nonce, raw...)), nil
}

func (v2ArgonAESGCM) getCipher(salt []byte, secret string) (c cipher.AEAD, err error) {
	// Ensure we're running only once as Argo2ID is quite expensive
	v2KeyExchangeLock.Lock()
	defer v2KeyExchangeLock.Unlock()

	c, err = getKDFResult[cipher.AEAD](protoVersionV2, salt, secret)
	switch {
	case err == nil:
		// Base case: result was already cached
		return c, nil

	case errors.Is(err, errKDFResultNotFound):
		// Well, we will generate it

	default:
		// Something else wrong
		return nil, fmt.Errorf("getting cached KDF result: %w", err)
	}

	key := argon2.IDKey(
		[]byte(secret),
		salt,
		v2ArgonTime,
		v2ArgonMemory,
		v2ArgonThreads,
		v2Keylen,
	)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("creating block cipher: %w", err)
	}

	if c, err = cipher.NewGCM(block); err != nil {
		return nil, fmt.Errorf("creating GCM cipher: %w", err)
	}

	setKDFResult(protoVersionV2, salt, secret, c)
	return c, nil
}
