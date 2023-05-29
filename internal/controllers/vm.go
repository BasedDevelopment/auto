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
	"strconv"

	"github.com/BasedDevelopment/auto/pkg/models"
	"github.com/rs/zerolog/log"
)

type VM models.VM

func (hv *HV) InitVMs() error {
	if err := hv.ensureConn(); err != nil {
		return err
	}

	// Get VMs from libvirt
	doms, err := hv.Libvirt.GetVMs()
	if err != nil {
		return err
	}

	hv.Mutex.Lock()
	defer hv.Mutex.Unlock()

	// Marshall the HV.VMs struct in
	for id, dom := range doms {
		hv.VMs[id] = &models.VM{
			Domain: dom,
			ID:     id,
		}
		go hv.fetchVMSpecs(hv.VMs[id])
	}

	return nil
}

func (hv *HV) fetchVMSpecs(vm *models.VM) {
	if err := hv.ensureConn(); err != nil {
		log.Error().Err(err).Msg("Failed to ensure connection")
	}

	specs, err := hv.Libvirt.GetVMSpecs(vm.Domain)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get VM specs")
	}

	vm.Mutex.Lock()
	defer vm.Mutex.Unlock()

	cpuInt, err := strconv.Atoi(specs.Vcpu.Text)
	if err != nil {
		log.Error().Err(err).Msg("Failed to convert CPU count to int")
	}
	vm.CPU = cpuInt

	mem, err := strconv.ParseInt(specs.Memory.Text, 10, 64)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse memory")
	}

	switch specs.Memory.Unit {
	case "KiB":
		mem = mem * 1024 * 1024
	case "MiB":
		mem = mem * 1024 * 1024 * 1024
	case "GiB":
		mem = mem * 1024 * 1024 * 1024 * 1024
	}

	vm.Memory = mem
}

func (hv *HV) GetVMState(vm *models.VM) (models.VMState, error) {
	if err := hv.ensureConn(); err != nil {
		return models.VMState{}, err
	}

	stateInt, stateStr, reasonStr, err := hv.Libvirt.GetVMState(vm.Domain)
	if err != nil {
		return models.VMState{}, err
	}

	return models.VMState{
		State:       stateInt,
		StateStr:    stateStr,
		StateReason: reasonStr,
	}, nil
}

func (hv *HV) GetVMConsole(vm *models.VM) (string, error) {
	if err := hv.ensureConn(); err != nil {
		return "", err
	}

	return hv.Libvirt.GetVMConsole(vm.Domain)
}
