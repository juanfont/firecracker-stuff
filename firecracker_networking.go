package tsif

import (
	"fmt"
	"log"
	"net/netip"
	"syscall"

	"github.com/coreos/go-iptables/iptables"
	"github.com/juanfont/headscale"
	"github.com/lorenzosaino/go-sysctl"
	"github.com/vishvananda/netlink"
)

type FirecrackerNetworking struct {
	Network  netip.Prefix
	NextAddr netip.Addr

	bridge *netlink.Bridge
}

// https://github.com/firecracker-microvm/firecracker/blob/main/docs/network-setup.md#advanced-setting-up-a-bridge-interface
// https://gist.github.com/s8sg/1acbe50c0d2b9be304cf46fa1e832847
func NewFirecrackerNetworking(network netip.Prefix) (*FirecrackerNetworking, error) {
	bridge, err := setupBridge(network)
	if err != nil {
		return nil, err
	}

	err = allowTrafficOnBridge(bridge)
	if err != nil {
		return nil, err
	}

	return &FirecrackerNetworking{
		bridge:   bridge,
		Network:  network,
		NextAddr: network.Addr().Next(),
	}, nil
}

func (f *FirecrackerNetworking) NextIP() netip.Addr {
	addr := f.NextAddr
	f.NextAddr = f.NextAddr.Next()
	return addr
}

func (f *FirecrackerNetworking) CreateTapDevice() (*netlink.Tuntap, error) {
	hash, err := headscale.GenerateRandomStringDNSSafe(scenarioHashLength)
	if err != nil {
		return nil, err
	}

	name := fmt.Sprintf("tap-ts-%s", hash)

	linkAttrs := netlink.NewLinkAttrs()
	linkAttrs.Name = name

	tapDevice := &netlink.Tuntap{
		LinkAttrs:  linkAttrs,
		Mode:       netlink.TUNTAP_MODE_TAP,
		NonPersist: false,
	}

	err = netlink.LinkAdd(tapDevice)
	if err != nil {
		return nil, err
	}

	err = netlink.LinkSetMaster(tapDevice, f.bridge)
	if err != nil {
		return nil, err
	}

	err = netlink.LinkSetUp(tapDevice)
	if err != nil {
		return nil, err
	}

	// we need to do this to refresh the link attributes
	link, err := netlink.LinkByName(name)
	if err != nil {
		return nil, err
	}
	tap, ok := link.(*netlink.Tuntap)
	if !ok {
		return nil, fmt.Errorf("link is not a tap device")
	}

	return tap, nil
}

func setupBridge(prefix netip.Prefix) (*netlink.Bridge, error) {
	hash, err := headscale.GenerateRandomStringDNSSafe(scenarioHashLength)
	if err != nil {
		return nil, err
	}

	la := netlink.NewLinkAttrs()
	la.Name = fmt.Sprintf("br-ts-%s", hash)
	br := &netlink.Bridge{LinkAttrs: la}
	err = netlink.LinkAdd(br)
	if err != nil && err != syscall.EEXIST {
		return nil, err
	}

	log.Printf("Created bridge %s", br.Name)
	netlinkAddr, err := netlink.ParseAddr(prefix.String())
	if err != nil {
		return nil, err
	}

	err = netlink.AddrAdd(br, netlinkAddr)
	if err != nil {
		return nil, err
	}

	err = netlink.LinkSetUp(br)
	if err != nil {
		return nil, err
	}

	return br, nil
}

func allowTrafficOnBridge(bridge *netlink.Bridge) error {
	err := sysctl.Set("net.ipv4.ip_forward", "1")
	if err != nil {
		return err
	}

	value, err := sysctl.Get("net.ipv4.ip_forward")
	if err != nil {
		return err
	}

	if value != "1" {
		return fmt.Errorf("net.ipv4.ip_forward is not set to 1")
	}

	ipTables, err := iptables.NewWithProtocol(iptables.ProtocolIPv4)
	if err != nil {
		return err
	}

	primaryLink, err := findDefaultGatewayInterface()
	if err != nil {
		return err
	}

	log.Printf("Primary link: %s", primaryLink.Attrs().Name)

	// sudo iptables --table nat --append POSTROUTING --out-interface enp3s0 -j MASQUERADE
	err = ipTables.Append("nat", "POSTROUTING", "-o", primaryLink.Attrs().Name, "-j", "MASQUERADE")
	if err != nil {
		return err
	}

	// sudo iptables --insert FORWARD --in-interface br0 -j ACCEPT
	err = ipTables.Append("filter", "FORWARD", "-i", bridge.Attrs().Name, "-j", "ACCEPT")
	if err != nil {
		return err
	}

	return nil
}

// findDefaultGatewayInterface returns the link that is used to connect to the default gateway (i.e.,
// the link that has internet access).
func findDefaultGatewayInterface() (netlink.Link, error) {
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
