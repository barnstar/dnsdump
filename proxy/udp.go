package proxy

import (
	"encoding/hex"
	"fmt"
	"net"
	"os"
)

func UDPWorker(loopbackAddr, forwardAddr net.IP) {
	forwardUDP := &net.UDPAddr{
		IP:   forwardAddr,
		Port: 53,
	}

	// Start UDP listener
	udpAddr := &net.UDPAddr{IP: loopbackAddr, Port: 53}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Printf("Failed to start UDP server: %v\n", err)
		fmt.Printf("Did you run sudo ifconfig lo0 alias %s up?\n", loopbackAddr.String())
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Printf("Listening on %s:53 (UDP). Forwarding to %v\n", loopbackAddr, forwardUDP)

	buffer := make([]byte, 512)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Printf("Error reading UDP packet: %v\n", err)
			continue
		}

		// Create hex dump of the packet
		dump := hex.Dump(buffer[:n])
		fmt.Printf("Received DNS packet from %v:\n%s\n", remoteAddr, dump)

		// Parse the question section (if packet is long enough)
		if n >= 12 { // Minimum DNS header size
			questionName, _ := parseDNSQuestion(buffer, 12)
			fmt.Printf("Resolving: %s via %v\n", questionName, forwardAddr)
		}

		// Forward the packet to upstream
		upstreamConn, err := net.DialUDP("udp", nil, forwardUDP)
		if err != nil {
			fmt.Printf("Failed to connect to upstream DNS: %v\n\n", err)
			continue
		}

		// Send the original query to upstream
		_, err = upstreamConn.Write(buffer[:n])
		if err != nil {
			fmt.Printf("Failed to forward packet: %v\n\n", err)
			upstreamConn.Close()
			continue
		}

		// Receive response from Cloudflare
		response := make([]byte, 512)
		nRead, err := upstreamConn.Read(response)
		if err != nil {
			fmt.Printf("Failed to receive response: %v\n\n", err)
			upstreamConn.Close()
			continue
		}

		// Log the response
		fmt.Printf("Response from upstream:\n%s\n", hex.Dump(response[:nRead]))

		// Send response back to client
		_, err = conn.WriteToUDP(response[:nRead], remoteAddr)
		if err != nil {
			fmt.Printf("Failed to send response to client: %v\n\n", err)
		}

		upstreamConn.Close()
		fmt.Printf("Closed UDP connection from %v\n\n", remoteAddr)
	}
}
