package controllers

import (
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"

	"github.com/BasedDevelopment/auto/internal/util"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func (hv *HV) CreateDomain(domId uuid.UUID, req *util.DomainCreateRequest) (err error) {
	// Validation of disk and image path is here due to import cycle
	if err := validation.ValidateStruct(req,
		validation.Field(&req.Image, validation.In(Images)),
		validation.Field(&req.CloudImage, validation.In(CloudImages)),
	); err != nil {
		return err
	}

	if req.Cloud && CloudInitPath == "" {
		return errors.New("no cloud-init path in auto config")
	}

	args := []string{
		"--uuid", domId.String(),
		"--name", req.Hostname,
		"--memory", strconv.Itoa(req.Memory),
		"--vcpus", strconv.Itoa(req.CPU),
		"--os-variant", req.OSVariant,
		//TODO: un hardcode this
		"--boot", "cdrom,fd,hd,network,menu=on",
	}

	for _, disk := range req.Disk {
		// check if disk.Path is empty or in []Disks
		if disk.Path == "" {
			return errors.New("disk path is empty")
		}
		if !util.Contains(Disks, disk.Path) {
			return errors.New("disk path is not in auto config")
		}
		dir := disk.Path + "/" + domId.String()

		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.Mkdir(dir, 0755); err != nil {
				return err
			}
		}

		// If cloudinit, disk zero is the image disk
		diskPath := dir + "/" + strconv.Itoa(disk.Id) + ".qcow2"
		if disk.Id == 0 && req.Cloud {
			if err := hv.CreateCloudDisk(diskPath, disk.Size, req.CloudImage); err != nil {
				return err
			}
		} else {
			if err := hv.CreateDisk(diskPath, disk.Size); err != nil {
				return err
			}
		}
		args = append(args, "--disk", "path="+diskPath+",format=qcow2")
	}

	if req.Cloud {
		cloudInitIsoPath := CloudInitPath + "/" + domId.String() + "-cidata.iso"
		if err := hv.CreateCloudInitIso(cloudInitIsoPath, req.UserData, req.MetaData); err != nil {
			return err
		}
		args = append(args, "--disk", "path="+cloudInitIsoPath+",device=cdrom")
	}

	if req.Image != "" {
		args = append(args, "--disk", "path="+req.Image+",device=cdrom")
	}

	log.Debug().
		Str("command", "virt-install").
		Strs("args", args).
		Msg("create domain")

	out, err := exec.Command("virt-install", args...).CombinedOutput()
	if err != nil {
		log.Error().
			Err(err).
			Str("command", "virt-install").
			Strs("args", args).
			Str("output", string(out)).
			Msg("create domain")
		return err
	}
	log.Debug().
		Str("output", string(out)).
		Msg("create domain")

	return nil
}

func (hv *HV) CreateDisk(path string, size int) error {
	args := []string{"create", "-f", "qcow2", path, strconv.Itoa(size) + "G"}
	log.Debug().
		Str("command", "qemu-img").
		Strs("args", args).
		Msg("create disk")
	out, err := exec.Command("qemu-img", args...).CombinedOutput()
	if err != nil {
		log.Error().
			Err(err).
			Str("command", "qemu-img").
			Strs("args", args).
			Str("output", string(out)).
			Msg("create disk")
		return err
	}
	log.Debug().
		Str("output", string(out)).
		Msg("create disk")

	return nil
}

func (hv *HV) CreateCloudDisk(path string, size int, image string) error {
	args := []string{"create", "-b", image, "-f", "qcow2", "-F", "qcow2", path, strconv.Itoa(size) + "G"}
	log.Debug().
		Str("command", "qemu-img").
		Strs("args", args).
		Msg("create cloud disk")
	out, err := exec.Command("qemu-img", args...).CombinedOutput()
	if err != nil {
		log.Error().
			Err(err).
			Str("command", "qemu-img").
			Strs("args", args).
			Str("output", string(out)).
			Msg("create cloud disk")
		return err
	}
	log.Debug().
		Str("output", string(out)).
		Msg("create cloud disk")
	return nil
}

func (hv *HV) CreateCloudInitIso(path string, userData string, metaData string) error {
	// Write cloud init file
	userDataFile, err := ioutil.TempFile("", "user-data-*.yaml")
	if err != nil {
		return err
	}
	metaDataFile, err := ioutil.TempFile("", "meta-data-*.yaml")
	if err != nil {
		return err
	}

	defer os.Remove(userDataFile.Name())
	defer os.Remove(metaDataFile.Name())

	if _, err := userDataFile.Write([]byte(userData)); err != nil {
		return err
	}

	if err := userDataFile.Close(); err != nil {
		return err
	}

	if _, err := metaDataFile.Write([]byte(metaData)); err != nil {
		return err
	}

	if err := metaDataFile.Close(); err != nil {
		return err
	}

	args := []string{"-output", path, "-V", "cidata", "-r", "-J", userDataFile.Name(), metaDataFile.Name()}
	log.Debug().
		Str("command", "genisoimage").
		Strs("args", args).
		Msg("create cloud init iso")
	out, err := exec.Command("genisoimage", args...).CombinedOutput()
	if err != nil {
		log.Error().
			Err(err).
			Str("command", "genisoimage").
			Strs("args", args).
			Str("output", string(out)).
			Msg("create cloud init iso")
		return err
	}
	log.Debug().
		Str("output", string(out)).
		Msg("create cloud init iso")
	return nil
}
