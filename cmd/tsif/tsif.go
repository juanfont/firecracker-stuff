package main

import (
	"fmt"
	"net/netip"

	tsif "github.com/juanfont/firecracker-stuff"
)

func main() {
	fmt.Println("LOL")

	cidr := netip.MustParsePrefix("172.20.0.1/24")
	bridgeName, err := tsif.SetupBridgeNetwork(cidr)
	if err != nil {
		panic(err)
	}
	fmt.Println(bridgeName)
}
