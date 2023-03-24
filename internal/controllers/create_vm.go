package controllers

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/BasedDevelopment/auto/internal/util"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
)

func (hv *HV) CheckCreateRequest(req *util.DomainCreateRequest) error {
	if err := validation.ValidateStruct(&req,
		validation.Field(&req.Image, validation.In(Images)),
	); err != nil {
		return err
	}

	for _, disk := range req.Disk {
		if err := validation.ValidateStruct(disk,
			validation.Field(&disk.Size, validation.Required),
			validation.Field(&disk.Disk, validation.In(Disks)),
		); err != nil {
			return err
		}
	}
	return nil
}

func (hv *HV) CreateDomain(domId uuid.UUID, req *util.DomainCreateRequest) (err error) {
	if err := hv.CheckCreateRequest(req); err != nil {
		return err
	}

	disks := []string{}
	for _, disk := range req.Disk {
		if disk.Id > 0 {
			continue
		}
		dir := disk.Disk + "/" + domId.String()
		os.Mkdir(dir, 0755)
		path := dir + "/" + strconv.Itoa(disk.Id) + ".qcow2"
		if req.Cloud {
			if err := hv.CreateCloudDisk(path, disk.Size, req.Image); err != nil {
				return err
			}
		} else {
			if err := hv.CreateDisk(path, disk.Size); err != nil {
				return err
			}
		}
	}
	//Append "--disk" for each disk
	diskList := ""
	for _, disk := range disks {
		diskList += "--disk path=" + disk + ",format=qcow2 "
	}

	if req.Cloud {
		if err := hv.CreateCloudInitIso(); err != nil {
			return err
		}
	}

	// TODO: Construct and run command

	cmd := exec.Command(
		"virt-install",
		"--name", domId.String(),
		"--memory", strconv.Itoa(req.Memory),
		"--vcpus", strconv.Itoa(req.CPU),
		"--os-type", req.OS,
		"--os-variant", req.OSVariant,
		diskList,
	)
	fmt.Println(cmd)
	return nil
	// return cmd.Run()
}

func (hv *HV) CreateDisk(path string, size int) error {
	return exec.Command("qemu-img", "create", "-f", "qcow2", path, strconv.Itoa(size)+"G").Run()
}

func (hv *HV) CreateCloudDisk(path string, size int, image string) error {
	return nil
}

func (hv *HV) CreateCloudInitIso() error {
	return nil
}
