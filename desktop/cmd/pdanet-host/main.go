package main

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"example.com/pdanet-open/desktop/protocol"
)

func main() {
	ctx := context.Background()
	_ = ctx

	addr := ":10209"
	if v := os.Getenv("PDANET_LISTEN"); v != "" {
		addr = v
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	log.Printf("pdanet-open host listening on %s", addr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("accept: %v", err)
			continue
		}
		go handle(conn)
	}
}

func handle(conn net.Conn) {
	defer conn.Close()
	log.Printf("client connected: %s", conn.RemoteAddr())
	defer log.Printf("client disconnected: %s", conn.RemoteAddr())

	r := bufio.NewReader(conn)
	for {
		frame, err := protocol.Read(r)
		if err != nil {
			if err != io.EOF {
				log.Printf("read frame: %v", err)
			}
			return
		}

		switch frame.Type {
		case protocol.FramePing:
			if err := protocol.Write(conn, protocol.Frame{Type: protocol.FramePong}); err != nil {
				log.Printf("write pong: %v", err)
				return
			}
		case protocol.FrameIP:
			log.Printf("IP packet received: %d bytes, version=%d", len(frame.Payload), ipVersion(frame.Payload))
			reply, ok := icmpEchoReply(frame.Payload)
			if ok {
				if err := protocol.Write(conn, protocol.Frame{Type: protocol.FrameIP, Payload: reply}); err != nil {
					log.Printf("write ICMP reply: %v", err)
					return
				}
				log.Printf("ICMP echo reply sent: %d bytes", len(reply))
			}
		default:
			log.Printf("unknown frame type: %d", frame.Type)
		}
	}
}

// icmpEchoReply recognizes a minimal Ethernet-free IPv4 + ICMP echo request
// and returns a standards-compliant echo response. This is a lab transport
// test, not the Internet forwarding path yet.
func icmpEchoReply(p []byte) ([]byte, bool) {
	if len(p) < 28 || p[0]>>4 != 4 {
		return nil, false
	}
	ihl := int(p[0]&0x0f) * 4
	if ihl < 20 || len(p) < ihl+8 || p[9] != 1 {
		return nil, false
	}
	if p[ihl] != 8 { // ICMP echo request
		return nil, false
	}

	out := append([]byte(nil), p...)
	copy(out[12:16], p[16:20])
	copy(out[16:20], p[12:16])
	out[ihl] = 0 // echo reply

	binary.BigEndian.PutUint16(out[10:12], 0)
	binary.BigEndian.PutUint16(out[10:12], checksum(out[:ihl]))

	binary.BigEndian.PutUint16(out[ihl+2:ihl+4], 0)
	binary.BigEndian.PutUint16(out[ihl+2:ihl+4], checksum(out[ihl:]))
	return out, true
}

func ipVersion(b []byte) int {
	if len(b) == 0 {
		return 0
	}
	return int(b[0] >> 4)
}

func checksum(b []byte) uint16 {
	var sum uint32
	for len(b) >= 2 {
		sum += uint32(binary.BigEndian.Uint16(b[:2]))
		b = b[2:]
	}
	if len(b) == 1 {
		sum += uint32(b[0]) << 8
	}
	for (sum >> 16) != 0 {
		sum = (sum & 0xffff) + (sum >> 16)
	}
	return ^uint16(sum)
}

var _ = fmt.Sprintf
