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
	"errors"
	"os"
	"path/filepath"

	"github.com/BasedDevelopment/auto/pkg/models"
)

func (hv *HV) DestroyVM(vm *models.VM) error {
	if err := hv.ensureConn(); err != nil {
		return err
	}

	return hv.Libvirt.DestroyVM(vm.Domain)
}

func (hv *HV) UndefineVM(vm *models.VM) error {
	if err := hv.ensureConn(); err != nil {
		return err
	}

	if err := hv.Libvirt.UndefineVM(vm.Domain); err != nil {
		return err
	}

	return hv.InitVMs()
}

func (hv *HV) DeleteDiskFile(path string) error {
	if _, err := os.Stat(path); err != nil {
		return errors.New("file does not exist")
	}

	if err := os.Remove(path); err != nil {
		return err
	}

	// see if the directory is empty, if it is, remove the directory
	dir := filepath.Dir(path)
	entry, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	if len(entry) != 0 {
		return nil
	}
	return os.Remove(dir)
}
