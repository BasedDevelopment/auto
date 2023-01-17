/*
 * eve - management toolkit for libvirt servers
 * Copyright (C) 2022-2023  BNS Services LLC

 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.

 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.

 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package controllers

import (
	"github.com/BasedDevelopment/auto/internal/libvirt"
	"github.com/BasedDevelopment/auto/pkg/models"
	"github.com/BasedDevelopment/eve/pkg/status"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

var Hypervisor = &HV{}

type HV models.HV

func (hv *HV) Init() error {
	hv.Libvirt = libvirt.Init(hv.IP, hv.Port)
	hv.Brs = make(map[string]*models.HVBr)
	hv.Storages = make(map[string]*models.HVStorage)
	hv.VMs = make(map[uuid.UUID]*models.VM)
	if err := hv.ensureConn(); err != nil {
		return err
	}
	if err := hv.InitVMs(); err != nil {
		return err
	}
	return nil
}

// Get the HV specs
func (hv *HV) getHVSpecs() error {
	if err := hv.ensureConn(); err != nil {
		return err
	}

	qemuVer, err := hv.Libvirt.GetHVQemuVersion()
	if err != nil {
		return err
	}
	hv.QemuVersion = qemuVer

	libvirtVer, err := hv.Libvirt.GetHVLibvirtVersion()
	if err != nil {
		return err
	}
	hv.LibvirtVersion = libvirtVer

	specs, err := hv.Libvirt.GetHVSpecs()
	if err != nil {
		return err
	}

	for _, spec := range specs.Processor.Entry {
		if spec.Name == "version" {
			hv.CPUModel = spec.Text
		}
	}
	if hv.CPUModel == "" {
		log.Warn().
			Msg("Could not find CPU version")
	}
	return nil
}

// Get HV stats
func (hv *HV) getHVStats() error {
	if err := hv.ensureConn(); err != nil {
		return err
	}

	arch, memoryTotal, memoryFree, cpus, mhz, nodes, sockets, cores, threads, err := hv.Libvirt.GetHVStats()
	if err != nil {
		return err
	}

	hv.Arch = arch
	hv.RAMTotal = memoryTotal
	hv.RAMFree = memoryFree
	hv.CPUCount = cpus
	hv.CPUFrequency = mhz
	hv.NUMANodes = nodes
	hv.CPUSockets = sockets
	hv.CPUCores = cores
	hv.CPUThreads = threads

	return nil
}

func (hv *HV) getHVBrs() error {
	if err := hv.ensureConn(); err != nil {
		return err
	}

	brs, err := hv.Libvirt.GetHVBrs()
	if err != nil {
		return err
	}

	for _, br := range brs {
		hv.Brs[br.Name] = &models.HVBr{
			Name: br.Name,
		}
	}
	return nil
}

// Ensure the HV libvirt connection is alive
func (hv *HV) ensureConn() error {
	if !hv.Libvirt.IsConnected() {
		return hv.connect()
	}
	return nil
}

// Connect to the HV libvirt
func (hv *HV) connect() error {
	hv.Mutex.Lock()
	defer hv.Mutex.Unlock()

	err := hv.Libvirt.Connect()
	if err != nil {
		hv.Status = status.StatusUnknown
		hv.StatusReason = err.Error()
		return err
	} else {
		hv.Status = status.StatusRunning
		hv.StatusReason = "Connected to libvirt"
	}
	if err := hv.getHVSpecs(); err != nil {
		return err
	}
	if err := hv.getHVStats(); err != nil {
		return err
	}
	if err := hv.getHVBrs(); err != nil {
		return err
	}
	return nil
}
