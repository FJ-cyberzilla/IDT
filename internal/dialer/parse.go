package dialer

import (
	"encoding/binary"
	"errors"
	"net"
)

// ParseA extracts A records from a DNS response.
func ParseA(resp []byte) ([]net.IP, error) {
	if len(resp) < 12 {
		return nil, errors.New("short packet")
	}
	ancount := binary.BigEndian.Uint16(resp[6:8])
	if ancount == 0 {
		return nil, errors.New("no answers")
	}
	off, err := skipName(resp, 12)
	if err != nil {
		return nil, err
	}
	off += 4

	var ips []net.IP
	for i := 0; i < int(ancount); i++ {
		off, err = skipName(resp, off)
		if err != nil {
			break
		}
		if off+10 > len(resp) {
			break
		}
		rrType := binary.BigEndian.Uint16(resp[off : off+2])
		rdlen := binary.BigEndian.Uint16(resp[off+8 : off+10])
		off += 10
		if off+int(rdlen) > len(resp) {
			break
		}
		if rrType == 1 && rdlen == 4 {
			ips = append(ips, net.IPv4(resp[off], resp[off+1], resp[off+2], resp[off+3]))
		}
		off += int(rdlen)
	}
	if len(ips) == 0 {
		return nil, errors.New("no A records")
	}
	return ips, nil
}

func skipName(buf []byte, off int) (int, error) {
	for {
		if off >= len(buf) {
			return 0, errors.New("overflow")
		}
		length := int(buf[off])
		if length == 0 {
			return off + 1, nil
		}
		if length&0xC0 == 0xC0 {
			if off+2 > len(buf) {
				return 0, errors.New("bad pointer")
			}
			return off + 2, nil
		}
		off += 1 + length
	}
}
