package core

import (
	"sync"

	"github.com/gusmin/gate/pkg/backend"
)

// session are the logged in user related informations.
type session struct {
	pubKey   []byte   // initialized in loadSSHPublickey
	user     user     // updated during background polling
	machines machines // updated during background polling
}

type machines struct {
	mu       sync.RWMutex
	machines []backend.Machine
}

func (m *machines) get() []backend.Machine {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.machines
}

func (m *machines) set(machines []backend.Machine) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.machines = machines
}

type user struct {
	mu   sync.RWMutex
	user backend.User
}

func (u *user) get() backend.User {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.user
}

func (u *user) set(user backend.User) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.user = user
}
