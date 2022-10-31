package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
)

const (
	tsifHashLength = 6

	tsifSocketPath      = "/tmp/firecracker.sock"
	tsifKernelImagePath = "/tmp/vmlinux.bin"
	tsifRootDrivePath   = "/tmp/root-drive.ext4"
	tsifKernelArgs      = "console=ttyS0 reboot=k panic=1 pci=off nomodules rw"
)

func main() {
	version := "1.30.0"
	tsif := New(version)

	// cmd := firecracker.VMCommandBuilder{}.
	// 	WithStdin(freezeReader{}).
	// 	WithStdout(io.Discard).
	// 	WithStderr(io.Discard).
	// 	WithSocketPath(e.socketFilePath).
	// 	Build(ctx)

}

func getFirecrackerConfig(tsif TailscaleInFirecracker) (*firecracker.Config, error) {
	originalFS, err := os.Open(tsif.originalRootDrive)
	if err != nil {
		err = fmt.Errorf("open original RootFS file: %w", err)
		return nil, err
	}
	defer originalFS.Close()

	rootFS, err := os.CreateTemp("", fmt.Sprintf("rootfs-%s-*.ext4", tsif.hostname))
	if err != nil {
		return nil, err
	}

	bytesWritten, err := io.Copy(rootFS, originalFS)
	if err != nil {
		return nil, err
	}

	log.Printf("Copied %.2fMB from original root drive to %s", bytesWritten/(1024*1024), rootFS.Name())

	err = rootFS.Sync()
	if err != nil {
		return nil, err
	}

	cfg := firecracker.Config{
		SocketPath:      tsif.socketFilePath,
		KernelImagePath: tsif.kernelImagePath,
		KernelArgs:      tsif.kernelArgs,
		Drives: []models.Drive{
			{
				DriveID:      firecracker.String("1"),
				IsReadOnly:   firecracker.Bool(false),
				IsRootDevice: firecracker.Bool(true),
				PathOnHost:   firecracker.String(rootFS.Name()),
			},
		},

		NetworkInterfaces: []firecracker.NetworkInterface{},
		FifoLogWriter:     nil,
		VsockDevices:      []firecracker.VsockDevice{},
		MachineCfg:        models.MachineConfiguration{},
		DisableValidation: false,
		JailerCfg:         &firecracker.JailerConfig{},
		VMID:              "",
		NetNS:             "",
		ForwardSignals:    []os.Signal{},
		Seccomp:           firecracker.SeccompConfig{},
		MmdsAddress:       []byte{},
		MmdsVersion:       "",
	}

}
