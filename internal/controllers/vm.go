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
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type VM models.VM

// Fetch VMs from the DB and Libvirt, marshall them into the HV struct,
// and check for consistency
func (hv *HV) InitVMs() error {
	if err := hv.ensureConn(); err != nil {
		return err
	}

	// Get VMs from libvirt
	libvirtVMs, err := hv.getVMsFromLibvirt()
	if err != nil {
		return err
	}

	hv.Mutex.Lock()
	defer hv.Mutex.Unlock()

	// Marshall the HV.VMs struct in
	for id, dom := range libvirtVMs {
		hv.VMs[id] = &models.VM{
			Domain: dom,
			ID:     id,
		}
	}

	go fetchVMState(hv)

	return nil
}

// Fetch VM state and state reason
func fetchVMState(hv *HV) {
	for uuid := range hv.VMs {
		if err := hv.GetVMState(hv.VMs[uuid]); err != nil {
			log.Error().Err(err).Msg("failed to get VM states")
			return
		}
	}
	log.Info().Msg("VM states fetched")
}

// Get the list of VMs from libvirt
// Will be used to check consistency
func (hv *HV) getVMsFromLibvirt() (doms map[uuid.UUID]libvirt.Dom, err error) {
	if err := hv.ensureConn(); err != nil {
		return nil, err
	}

	doms, err = hv.Libvirt.GetVMs()
	if err != nil {
		return nil, err
	}
	return
}

func (hv *HV) GetVMState(vm *models.VM) (err error) {
	if err := hv.ensureConn(); err != nil {
		return err
	}

	vm.Mutex.Lock()
	defer vm.Mutex.Unlock()

	stateInt, stateStr, reasonStr, err := hv.Libvirt.GetVMState(vm.Domain)
	if err != nil {
		return err
	}

	vm.State = stateInt
	vm.StateStr = stateStr
	vm.StateReason = reasonStr
	return nil
}
