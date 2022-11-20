package tsif

import (
	"fmt"
	"io"
	"log"
	"net/netip"
	"os"
	"syscall"

	"github.com/coreos/go-iptables/iptables"
	"github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
	"github.com/juanfont/headscale"
	sysctl "github.com/lorenzosaino/go-sysctl"
	"github.com/vishvananda/netlink"
)

// setupBridgeNetwork creates a bridge and a bridge for the tap interfaces
// to be connected to.
// https://github.com/firecracker-microvm/firecracker/blob/main/docs/network-setup.md#advanced-setting-up-a-bridge-interface
// https://gist.github.com/s8sg/1acbe50c0d2b9be304cf46fa1e832847
func SetupBridgeNetwork(addr netip.Prefix) (*string, error) {
	hash, err := headscale.GenerateRandomStringDNSSafe(scenarioHashLength)
	if err != nil {
		return nil, err
	}

	la := netlink.NewLinkAttrs()
	la.Name = fmt.Sprintf("br-hs-%s", hash)
	br := &netlink.Bridge{LinkAttrs: la}
	err = netlink.LinkAdd(br)
	if err != nil && err != syscall.EEXIST {
		return nil, err
	}

	log.Printf("Created bridge %s", br.Name)

	panic("bridge up!")

	netlinkAddr, err := netlink.ParseAddr(addr.String())
	if err != nil {
		return nil, err
	}

	err = netlink.AddrAdd(br, netlinkAddr)
	if err != nil {
		return nil, err
	}

	err = sysctl.Set("net.ipv4.ip_forward", "1")
	if err != nil {
		return nil, err
	}

	value, err := sysctl.Get("net.ipv4.ip_forward")
	if err != nil {
		return nil, err
	}

	if value != "1" {
		return nil, fmt.Errorf("net.ipv4.ip_forward is not set to 1")
	}

	ipTables, err := iptables.NewWithProtocol(iptables.ProtocolIPv4)
	if err != nil {
		return nil, err
	}

	// sudo iptables --insert FORWARD --in-interface br0 -j ACCEPT
	primaryLink, err := findPrimaryLink()
	if err != nil {
		return nil, err
	}

	log.Printf("Primary link: %s", primaryLink.Attrs().Name)

	// sudo iptables --table nat --append POSTROUTING --out-interface enp3s0 -j MASQUERADE
	err = ipTables.Append("nat", "POSTROUTING", "-o", primaryLink.Attrs().Name, "-j", "MASQUERADE")
	if err != nil {
		return nil, err
	}

	// sudo iptables --insert FORWARD --in-interface br0 -j ACCEPT
	err = ipTables.Append("filter", "FORWARD", "-i", br.Attrs().Name, "-j", "ACCEPT")
	if err != nil {
		return nil, err
	}

	// routes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	// if err != nil {
	// 	return nil, err
	// }

	// for _, r := range routes {
	// 	fmt.Println(r)
	// }

	// for _, route := range routes {
	// 	if route.Dst

	// netlink.LinkSetMaster(eth1, mybridge)

	// version := "1.30.0"
	// tsif, err := New(version)
	// if err != nil {
	// 	log.Fatalf("failed to create tsif: %v", err)
	// }

	// fmt.Println(tsif)

	// cmd := firecracker.VMCommandBuilder{}.
	// 	WithStdin(freezeReader{}).
	// 	WithStdout(io.Discard).
	// 	WithStderr(io.Discard).
	// 	WithSocketPath(e.socketFilePath).
	// 	Build(ctx)
	return &br.Attrs().Name, nil
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

	return &cfg, nil
}

// findPrimaryLink returns the link that is used to connect to the default gateway (i.e.,
// the link that has internet access).
func findPrimaryLink() (netlink.Link, error) {
	routes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		return nil, err
	}

	for _, r := range routes {
		if r.Dst == nil {
			link, err := netlink.LinkByIndex(r.LinkIndex)
			if err != nil {
				return nil, err
			}

			return link, nil
		}
	}

	return nil, fmt.Errorf("no primary interface found")
}
