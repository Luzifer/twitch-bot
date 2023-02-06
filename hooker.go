package main

import (
	"sync"

	"github.com/gofrs/uuid/v3"
)

type (
	hooker struct {
		hooks map[string]func()
		lock  sync.RWMutex
	}
)

func newHooker() *hooker { return &hooker{hooks: map[string]func(){}} }

func (h *hooker) Ping() {
	h.lock.RLock()
	defer h.lock.RUnlock()

	for _, hf := range h.hooks {
		hf()
	}
}

func (h *hooker) Register(hook func()) func() {
	h.lock.Lock()
	defer h.lock.Unlock()

	id := uuid.Must(uuid.NewV4()).String()
	h.hooks[id] = hook

	return func() { h.unregister(id) }
}

func (h *hooker) unregister(id string) {
	h.lock.Lock()
	defer h.lock.Unlock()

	delete(h.hooks, id)
}
