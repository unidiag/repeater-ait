package main

import (
	"os"

	"github.com/fatih/color"
)

func main() {

	if len(os.Args) != 3 {
		color.Green(`	This program retransmits incoming UDP to the next port with the addition of the AIT.
	Copyright VITALI TUMASHEUSKI aka @unidiag in 2024 (email: tva@tva.by)
	Usage:		./repeater-ait <udp:port> <hbblink>
	Example:	./repeater-ait udp://eth1@239.0.100.1:1234 http://hbbtv.com/app`)
		os.Exit(1)
	}

	aitRepeater(os.Args[1], os.Args[2])
}
