package core

import (
	"sync"

	"github.com/gusmin/gate/pkg/backend"
)

// session are the logged in user related informations.
type session struct {
	pubKey   []byte       // initialized in loadSSHPublickey
	user     safeUser     // updated during background polling
	machines safeMachines // updated during background polling
}

type safeMachines struct {
	mu       sync.RWMutex
	machines []backend.Machine
}

func (m *safeMachines) get() []backend.Machine {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.machines
}

func (m *safeMachines) set(machines []backend.Machine) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.machines = machines
}

type safeUser struct {
	mu   sync.RWMutex
	user backend.User
}

func (u *safeUser) get() backend.User {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.user
}

func (u *safeUser) set(user backend.User) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.user = user
}
