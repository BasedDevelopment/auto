package models

import (
	"net"
	"sync"
	"time"

	"github.com/BasedDevelopment/auto/internal/libvirt"
	"github.com/BasedDevelopment/eve/pkg/status"
	"github.com/google/uuid"
)

// The struct for the hypervisor status (more transient), this is different from the
// struct in eve, which persists the data to the database.
type HV struct {
	Mutex          sync.Mutex            `json:"-"`
	IP             net.IP                `json:"-"`
	Port           int                   `json:"-"`
	CPUModel       string                `json:"cpu_model"`
	Arch           string                `json:"arch"`
	RAMTotal       uint64                `json:"total_ram"`
	RAMFree        uint64                `json:"free_ram"`
	CPUCount       int32                 `json:"cpu_count"`
	CPUFrequency   int32                 `json:"cpu_frequency_mhz"`
	NUMANodes      int32                 `json:"numa_nodes"`
	CPUSockets     int32                 `json:"cpu_sockets"`
	CPUCores       int32                 `json:"cpu_cores"`
	CPUThreads     int32                 `json:"cpu_threads"`
	Brs            map[string]*HVBr      `json:"-"`
	Storages       map[string]*HVStorage `json:"-"`
	VMs            map[uuid.UUID]*VM     `json:"-"`
	Status         status.Status         `json:"status"`
	StatusReason   string                `json:"status_reason"`
	QemuVersion    string                `json:"qemu_version"`
	LibvirtVersion string                `json:"libvirt_version"`
	Libvirt        *libvirt.Libvirt      `json:"-"`
}

type HVBr struct {
	Name    string `json:"name"`
	Remarks string `json:"remarks"`
}

type HVStorage struct {
	ID      uuid.UUID `json:"id"`
	Size    int       `json:"size"`
	Used    int       `json:"used"`
	Free    int       `json:"free"`
	Updated time.Time `json:"updated"`
	Remarks string    `json:"remarks"`
}
