package controllers

import (
	"os"

	"github.com/BasedDevelopment/auto/internal/config"
	"github.com/rs/zerolog/log"
)

var (
	CloudImages   = []string{}
	Images        = []string{}
	Disks         = []string{}
	CloudInitPath = ""
)

func CheckStorage() error {
	for _, storage := range config.Config.Storage {
		if !storage.Enabled {
			continue
		}

		if _, err := os.Stat(storage.Path); os.IsNotExist(err) {
			return err
		}

		if storage.CloudImage {
			path := storage.Path + "/cloud-images"
			if _, err := os.Stat(path); os.IsNotExist(err) {
				if err := os.Mkdir(path, 0755); err != nil {
					return err
				}
			}
			CloudImages = append(CloudImages, path)
		}

		if storage.Iso {
			path := storage.Path + "/images"
			if _, err := os.Stat(path); os.IsNotExist(err) {
				if err := os.Mkdir(path, 0755); err != nil {
					return err
				}
			}
			Images = append(Images, path)
		}

		if storage.Disk {
			path := storage.Path + "/disks"
			if _, err := os.Stat(path); os.IsNotExist(err) {
				if err := os.Mkdir(path, 0755); err != nil {
					return err
				}
			}
			Disks = append(Disks, path)
		}
	}

	if config.Config.CloudInit.Enabled {
		if _, err := os.Stat(config.Config.CloudInit.Path); os.IsNotExist(err) {
			return err
		} else {
			CloudInitPath = config.Config.CloudInit.Path
		}
	}

	log.Info().
		Strs("cloud_images", CloudImages).
		Strs("images", Images).
		Strs("disks", Disks).
		Str("cloud_init", CloudInitPath).
		Msg("Storage check complete")

	return nil
}
