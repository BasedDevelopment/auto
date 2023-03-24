package controllers

import (
	"os"

	"github.com/BasedDevelopment/auto/internal/config"
)

var (
	Images = []string{}
	Disks  = []string{}
)

func CheckStorage() error {
	for _, storage := range config.Config.Storage {
		if !storage.Enabled {
			return nil
		}

		if _, err := os.Stat(storage.Path); os.IsNotExist(err) {
			return err
		}

		if storage.Iso {
			Images = append(Images, storage.Path)
		}

		if storage.Disk {
			Disks = append(Disks, storage.Path)
		}
	}
	return nil
}
