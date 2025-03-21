package proxy

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"
)

func TCPWorker(loopbackAddr, forwardAddr net.IP) {
	forwardTCPAddr := &net.TCPAddr{
		IP:   forwardAddr,
		Port: 53,
	}

	// Start TCP listener
	tcpAddr := &net.TCPAddr{IP: loopbackAddr, Port: 53}
	tcpListener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		fmt.Printf("Failed to start TCP server: %v\n", err)
		fmt.Printf("Did you run sudo ifconfig lo0 alias %s up?\n", loopbackAddr)
		os.Exit(1)
	}
	defer tcpListener.Close()
	fmt.Printf("Listening on %s:53 (TCP). Forwarding to %v\n", loopbackAddr, forwardTCPAddr)

	for {
		conn, err := tcpListener.Accept()
		if err != nil {
			fmt.Printf("Error accepting TCP connection: %v\n", err)
			continue
		}
		go handleTCPRequest(conn, forwardTCPAddr)
	}
}

func handleTCPRequest(conn net.Conn, forwardAddr *net.TCPAddr) {
	defer conn.Close()
	defer fmt.Printf("Closed TCP connection from %v\n\n", conn.RemoteAddr())

	// First two bytes in TCP DNS are length
	lengthBytes := make([]byte, 2)
	_, err := io.ReadFull(conn, lengthBytes)
	if err != nil {
		fmt.Printf("Error reading TCP length: %v\n", err)
		return
	}

	length := binary.BigEndian.Uint16(lengthBytes)
	buffer := make([]byte, length)
	_, err = io.ReadFull(conn, buffer)
	if err != nil {
		fmt.Printf("Error reading TCP message: %v\n", err)
		return
	}

	fmt.Printf("Received TCP DNS packet from %v:\n%s\n", conn.RemoteAddr(), hex.Dump(buffer))

	if len(buffer) >= 12 {
		questionName, _ := parseDNSQuestion(buffer, 12)
		fmt.Printf("Resolving: %s via %v\n", questionName, forwardAddr)
	}

	// Forward to upstream
	upstreamConn, err := net.DialTCP("tcp", nil, forwardAddr)
	if err != nil {
		fmt.Printf("Failed to connect to upstream DNS TCP: %v\n", err)
		return
	}
	defer upstreamConn.Close()

	// Write length prefix and message
	_, err = upstreamConn.Write(lengthBytes)
	if err != nil {
		fmt.Printf("Failed to write TCP length to upstream: %v\n", err)
		return
	}
	_, err = upstreamConn.Write(buffer)
	if err != nil {
		fmt.Printf("Failed to write TCP message to upstream: %v\n", err)
		return
	}

	// Read response length
	_, err = io.ReadFull(upstreamConn, lengthBytes)
	if err != nil {
		fmt.Printf("Failed to read TCP response length: %v\n", err)
		return
	}

	length = binary.BigEndian.Uint16(lengthBytes)
	response := make([]byte, length)
	_, err = io.ReadFull(upstreamConn, response)
	if err != nil {
		fmt.Printf("Failed to read TCP response: %v\n", err)
		return
	}

	fmt.Printf("Response from %s (TCP):\n%s\n", forwardAddr, hex.Dump(response))

	// Send response back to client
	_, err = conn.Write(lengthBytes)
	if err != nil {
		fmt.Printf("Failed to write TCP length to client: %v\n", err)
		return
	}
	_, err = conn.Write(response)
	if err != nil {
		fmt.Printf("Failed to write TCP response to client: %v\n", err)
	}
}
