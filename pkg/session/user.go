package session

import (
	"sync"

	"github.com/gusmin/gate/pkg/backend"
)

// userInfos are the logged in user related informations.
type userInfos struct {
	pubKey   []byte
	user     safeUser     // accessed during background polling
	machines safeMachines // accessed during background polling
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
