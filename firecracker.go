package tsif

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/netip"
	"os"

	"github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
)

const (
	tsifKernelImagePath = "./hello-vmlinux.bin"
	tsifRootDrivePath   = "./hello-rootfs.ext4"
	tsifKernelArgs      = "console=ttyS0 reboot=k panic=1 pci=off nomodules rw"
)

type NetworkConfig struct {
	IP      netip.Addr
	Gateway netip.Addr
	Network netip.Prefix

	TapDevice string
	TapMAC    string

	CloudInitURL string
}

func (t *TailscaleInFirecracker) getFirecrackerConfig(networkConfig NetworkConfig) (*firecracker.Config, error) {
	originalFS, err := os.Open(t.originalRootDrive)
	if err != nil {
		err = fmt.Errorf("open original RootFS file: %w", err)
		return nil, err
	}
	defer originalFS.Close()

	rootFS, err := os.CreateTemp("", fmt.Sprintf("rootfs-%s-*.ext4", t.hostname))
	if err != nil {
		return nil, err
	}

	bytesWritten, err := io.Copy(rootFS, originalFS)
	if err != nil {
		return nil, err
	}

	log.Printf("Copied %.2fMB from original root drive to %s", float64(bytesWritten)/(1024.0*1024.0), rootFS.Name())

	err = rootFS.Sync()
	if err != nil {
		return nil, err
	}

	_, netmask, err := net.ParseCIDR(networkConfig.Network.String())
	if err != nil {
		return nil, err
	}

	kernelArgs := fmt.Sprintf("%s ip=%s::%s:%s::eth0:off nameserver=1.1.1.1 cloud-config-url=%s",
		t.kernelArgs,
		networkConfig.IP.String(),
		networkConfig.Gateway.String(),
		net.IP(netmask.Mask).String(),
		networkConfig.CloudInitURL,
	)

	fmt.Println(kernelArgs)

	cfg := firecracker.Config{
		SocketPath:      fmt.Sprintf("/tmp/firecracker-%s.sock", t.hostname),
		KernelImagePath: t.kernelImagePath,
		KernelArgs:      kernelArgs,
		Drives: []models.Drive{
			{
				DriveID:      firecracker.String("1"),
				IsReadOnly:   firecracker.Bool(false),
				IsRootDevice: firecracker.Bool(true),
				PathOnHost:   firecracker.String(rootFS.Name()),
			},
		},
		NetworkInterfaces: []firecracker.NetworkInterface{
			{
				StaticConfiguration: &firecracker.StaticNetworkConfiguration{
					MacAddress:  networkConfig.TapMAC,
					HostDevName: networkConfig.TapDevice,
				},
			},
		},
		MachineCfg: models.MachineConfiguration{
			MemSizeMib: firecracker.Int64(256),
			VcpuCount:  firecracker.Int64(1),
		},
	}

	return &cfg, nil
}
