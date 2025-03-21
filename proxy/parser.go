package proxy

import (
	"strings"
)

type DNSHeader struct {
	ID      uint16
	Flags   uint16
	QDCount uint16
	ANCount uint16
	NSCount uint16
	ARCount uint16
}

const headerLength = 12

func parseDNSQuestion(packet []byte, offset int) (string, int) {
	var name strings.Builder
	for {
		length := int(packet[offset])
		if length == 0 {
			break
		}
		offset++
		name.Write(packet[offset : offset+length])
		name.WriteByte('.')
		offset += length
	}
	return strings.TrimSuffix(name.String(), "."), offset + 1
}
