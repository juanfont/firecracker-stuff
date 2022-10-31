package main

import (
	"errors"
	"fmt"
	"net/netip"
	"strings"

	"github.com/juanfont/headscale"
	"tailscale.com/ipn/ipnstate"
)

type TailscaleInFirecracker struct {
	version  string
	hostname string

	socketFilePath    string
	kernelImagePath   string
	kernelArgs        string
	originalRootDrive string
}

func New(
	version string,
) (*TailscaleInFirecracker, error) {
	hash, err := headscale.GenerateRandomStringDNSSafe(tsifHashLength)
	if err != nil {
		return nil, err
	}

	hostname := fmt.Sprintf("ts-%s-%s", strings.ReplaceAll(version, ".", "-"), hash)

	tsif := TailscaleInFirecracker{
		version:           version,
		hostname:          hostname,
		socketFilePath:    tsifSocketPath,
		kernelImagePath:   tsifKernelImagePath,
		kernelArgs:        tsifKernelArgs,
		originalRootDrive: tsifRootDrivePath,
	}

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
