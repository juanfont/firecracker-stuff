package main

import (
	"fmt"
	"net/netip"

	tsif "github.com/juanfont/firecracker-stuff"
)

func main() {
	fmt.Println("LOL")

	addr := netip.MustParsePrefix("172.26.0.1/24")

	firecrackerNetworking, err := tsif.NewFirecrackerManager(addr)
	if err != nil {
		panic(err)
	}

	tsif, err := tsif.New(firecrackerNetworking, "1.30.0")
	if err != nil {
		panic(err)
	}

	fmt.Println(tsif)
}
