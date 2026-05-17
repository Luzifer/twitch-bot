package cryptkv

import (
	"crypto/sha256"
	"fmt"
	"sync"
)

var (
	kdfResultCache     = make(map[string]any)
	kdfResultCacheLock sync.Mutex

	errKDFResultNotFound = fmt.Errorf("kdf result not found")
)

func deriveKDFResultKey(pv protoVersion, salt []byte, secret string) string {
	return fmt.Sprintf(
		"protover=%d;salt=%x;sha256=%x",
		pv,
		sha256.Sum256(salt),
		sha256.Sum256([]byte(secret)),
	)
}

func getKDFResult[T any](pv protoVersion, salt []byte, secret string) (T, error) {
	kdfResultCacheLock.Lock()
	defer kdfResultCacheLock.Unlock()

	resRaw, ok := kdfResultCache[deriveKDFResultKey(pv, salt, secret)]
	if !ok {
		return *new(T), errKDFResultNotFound
	}

	raw, ok := resRaw.(T)
	if !ok {
		return *new(T), fmt.Errorf("casting to %T not possible", *new(T))
	}

	return raw, nil
}

func setKDFResult(pv protoVersion, salt []byte, secret string, res any) {
	kdfResultCacheLock.Lock()
	defer kdfResultCacheLock.Unlock()

	kdfResultCache[deriveKDFResultKey(pv, salt, secret)] = res
}
