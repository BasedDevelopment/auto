package main

import (
	"log"
	"os/exec"
)

var binaries = []string{
	"virt-install",
	"qemu-img",
	"genisoimage",
}

func checkPaths() {
	for _, b := range binaries {
		_, err := exec.LookPath(b)
		if err != nil {
			log.Fatalf("Could not find %s in PATH", b)
		}
	}
}
