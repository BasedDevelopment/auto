package models

import (
	"net"
	"sync"
	"time"

	"github.com/BasedDevelopment/auto/internal/libvirt"
	"github.com/BasedDevelopment/eve/pkg/status"
	"github.com/google/uuid"
)

type VM struct {
	Mutex    sync.Mutex           `json:"-"`
	Domain   libvirt.Dom          `json:"-"`
	ID       uuid.UUID            `json:"id"`
	CPU      int                  `json:"cpu"`
	Memory   int64                `json:"memory"`
	Nics     map[string]VMNic     `json:"nics"`
	Storages map[string]VMStorage `json:"storages"`
}

type VMState struct {
	State       status.Status `json:"state"`
	StateStr    string        `json:"state_str"`
	StateReason string        `json:"state_reason"`
}

type VMNic struct {
	Mutex   sync.Mutex `json:"-"`
	ID      uuid.UUID  `json:"id"`
	name    string     `json:"name"`
	MAC     string     `json:"mac"`
	IP      []net.IP   `json:"ip"`
	Created time.Time  `json:"created"`
	Updated time.Time  `json:"updated"`
	Remarks string     `json:"remarks"`
	State   string     `json:"state"`
}

type VMStorage struct {
	Mutex   sync.Mutex `json:"-"`
	ID      uuid.UUID  `json:"id"`
	Path    string     `json:"path"`
	Size    int        `json:"size"`
	Created time.Time  `json:"created"`
	Updated time.Time  `json:"updated"`
	Remarks string     `json:"remarks"`
}
