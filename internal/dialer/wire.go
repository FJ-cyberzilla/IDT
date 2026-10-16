package dialer

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"strings"
)

// buildQuery constructs a minimal DNS query packet for an A record.
// ~30 bytes, zero dependencies, random transaction ID.
func buildQuery(target string) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint16(rand.Intn(65535)))
	binary.Write(buf, binary.BigEndian, uint16(0x0100)) // RD
	binary.Write(buf, binary.BigEndian, uint16(1))      // QDCOUNT
	binary.Write(buf, binary.BigEndian, uint16(0))      // ANCOUNT
	binary.Write(buf, binary.BigEndian, uint16(0))      // NSCOUNT
	binary.Write(buf, binary.BigEndian, uint16(0))      // ARCOUNT

	for _, label := range strings.Split(target, ".") {
		buf.WriteByte(byte(len(label)))
		buf.WriteString(label)
	}
	buf.WriteByte(0)
	binary.Write(buf, binary.BigEndian, uint16(1)) // TYPE A
	binary.Write(buf, binary.BigEndian, uint16(1)) // CLASS IN
	return buf.Bytes()
}
