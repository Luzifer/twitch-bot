// Package locker contains a way to interact with arbitrary locks
package locker

import "sync"

var (
	locks       = map[string]*sync.RWMutex{}
	locksOLocks sync.RWMutex
)

// LockByKey takes a key to lock and locks the corresponding RWMutex
func LockByKey(key string) { getLockByKey(key).Lock() }

// RLockByKey takes a key to lock and read-locks the corresponding RWMutex
func RLockByKey(key string) { getLockByKey(key).RLock() }

// RUnlockByKey takes a key to lock and read-unlocks the corresponding RWMutex
func RUnlockByKey(key string) { getLockByKey(key).RUnlock() }

// UnlockByKey takes a key to lock and unlocks the corresponding RWMutex
func UnlockByKey(key string) { getLockByKey(key).Unlock() }

// WithLock takes a key to lock and a function to execute during the
// lock of this key
func WithLock(key string, fn func()) {
	LockByKey(key)
	defer UnlockByKey(key)

	fn()
}

// WithRLock takes a key to lock and a function to execute during the
// read-lock of this key
func WithRLock(key string, fn func()) {
	RLockByKey(key)
	defer RUnlockByKey(key)

	fn()
}

func getLockByKey(key string) *sync.RWMutex {
	locksOLocks.Lock()
	defer locksOLocks.Unlock()

	if locks[key] == nil {
		locks[key] = new(sync.RWMutex)
	}

	return locks[key]
}
