package tsif

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/netip"
	"strings"

	"github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/juanfont/headscale"
	"tailscale.com/ipn/ipnstate"
)

const (
	scenarioHashLength = 6
	tsifHashLength     = 6
)

type TailscaleInFirecracker struct {
	version  string
	hostname string

	socketFilePath    string
	kernelImagePath   string
	kernelArgs        string
	originalRootDrive string

	firecrackerNetworking *FirecrackerNetworking
}

func New(
	firecrackerNetworking *FirecrackerNetworking,
	version string,
) (*TailscaleInFirecracker, error) {
	hash, err := headscale.GenerateRandomStringDNSSafe(tsifHashLength)
	if err != nil {
		return nil, err
	}

	hostname := fmt.Sprintf("ts-%s-%s", strings.ReplaceAll(version, ".", "-"), hash)

	tsif := &TailscaleInFirecracker{
		version:           version,
		hostname:          hostname,
		kernelImagePath:   tsifKernelImagePath,
		kernelArgs:        tsifKernelArgs,
		originalRootDrive: tsifRootDrivePath,

		firecrackerNetworking: firecrackerNetworking,
	}

	tapDevice, err := firecrackerNetworking.CreateTapDevice()
	if err != nil {
		return nil, err
	}

	tapDeviceAttrs := tapDevice.Attrs()

	networkConfig := NetworkConfig{
		IP:        firecrackerNetworking.NextIP(),
		Gateway:   firecrackerNetworking.Network.Addr(),
		Network:   firecrackerNetworking.Network,
		TapDevice: tapDeviceAttrs.Name,
		TapMAC:    tapDeviceAttrs.HardwareAddr.String(),
	}

	config, err := tsif.getFirecrackerConfig(networkConfig)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	cmd := firecracker.VMCommandBuilder{}.WithSocketPath(config.SocketPath).Build(ctx)
	machine, err := firecracker.NewMachine(ctx, *config, firecracker.WithProcessRunner(cmd))
	if err != nil {
		return nil, err
	}

	log.Printf("Starting machine")
	err = machine.Start(ctx)
	if err != nil {
		return nil, err
	}

	return tsif, nil
}

func (t *TailscaleInFirecracker) Hostname() string {
	return t.hostname
}

func (t *TailscaleInFirecracker) Shutdown() error {
	return errors.New("not implemented")
}

func (t *TailscaleInFirecracker) Version() string {
	return t.version
}

func (t *TailscaleInFirecracker) Execute(command []string) (string, error) {
	return "", errors.New("not implemented")
}

func (t *TailscaleInFirecracker) Up(loginServer, authKey string) error {
	return errors.New("not implemented")
}

func (t *TailscaleInFirecracker) IPs() ([]netip.Addr, error) {
	return nil, errors.New("not implemented")
}

func (t *TailscaleInFirecracker) FQDN() (string, error) {
	return "", errors.New("not implemented")
}

func (t *TailscaleInFirecracker) Status() (*ipnstate.Status, error) {
	return nil, errors.New("not implemented")
}

func (t *TailscaleInFirecracker) WaitForPeers(expected int) error {
	return errors.New("not implemented")
}

func (t *TailscaleInFirecracker) Ping(hostnameOrIP string) error {
	return errors.New("not implemented")
}
