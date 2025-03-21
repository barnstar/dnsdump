package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"barnstar.com/dnsdump/proxy"
)

var (
	forwardDNSAddr net.IP
	loopbackAddr   net.IP
)

func main() {
	upstream := flag.String("f", "1.1.1.1", "upstream DNS resolver IP")
	listen := flag.String("l", "127.0.1.2", "local IP to listen on")
	flag.Parse()

	// Validate IP addresses
	if net.ParseIP(*upstream) == nil {
		fmt.Printf("Invalid upstream IP address: %s\n", *upstream)
		os.Exit(1)
	}
	if net.ParseIP(*listen) == nil {
		fmt.Printf("Invalid listen IP address: %s\n", *listen)
		os.Exit(1)
	}

	forwardDNSAddr = net.ParseIP(*upstream)
	loopbackAddr = net.ParseIP(*listen)

	if !loopbackAddr.IsLoopback() {
		fmt.Printf("Invalid listen IP address (not a loobackIP): %s\n", loopbackAddr)
		os.Exit(1)
	}

	go proxy.TCPWorker(loopbackAddr, forwardDNSAddr)
	go proxy.UDPWorker(loopbackAddr, forwardDNSAddr)

	select {}
}
