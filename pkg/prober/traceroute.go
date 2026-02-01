package prober

import (
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type TracerouteRunner struct {
	Target      string
	MaxHops     int
	CountPerHop int // Number of probes per hop (typically 3)
	Timeout     time.Duration
}

func NewTracerouteRunner(target string) *TracerouteRunner {
	return &TracerouteRunner{
		Target:      target,
		MaxHops:     30,
		CountPerHop: 1, // Start simple
		Timeout:     2 * time.Second,
	}
}

// Run executes a traceroute
func (t *TracerouteRunner) Run() (*TraceResult, error) {
	dstAddr, err := net.ResolveIPAddr("ip4", t.Target)
	if err != nil {
		return nil, err
	}

	// Traceroute requires receiving TimeExceeded messages, which usually needs raw sockets
	c, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return nil, fmt.Errorf("traceroute requires root privileges (ip4:icmp): %w", err)
	}
	defer c.Close()

	// Set invalidation on the socket to ensure we can set TTL
	pconn := c.IPv4PacketConn()
	if pconn != nil {
		pconn.SetControlMessage(ipv4.FlagTTL, true)
	}

	res := &TraceResult{
		Target:    t.Target,
		Timestamp: time.Now(),
		Hops:      []HopInfo{},
	}

	id := os.Getpid() & 0xffff

	for ttl := 1; ttl <= t.MaxHops; ttl++ {
		// Prepare probe
		// We use ICMP Echo as the probe (Windows style)
		wm := icmp.Message{
			Type: ipv4.ICMPTypeEcho, Code: 0,
			Body: &icmp.Echo{
				ID: id, Seq: ttl,
				Data: []byte("RouteLens-Trace"),
			},
		}
		wb, err := wm.Marshal(nil)
		if err != nil {
			return nil, err
		}

		// Set TTL
		if pconn != nil {
			pconn.SetTTL(ttl)
		} else {
			// Fallback for systems where IPv4PacketConn wrapper might fail, though unlikely with ip4:icmp
			// Using syscall setsockopt would be the raw way, but pconn.SetTTL is standard x/net
		}

		start := time.Now()
		if _, err := c.WriteTo(wb, dstAddr); err != nil {
			// If sending fails, maybe stop?
			continue
		}

		// Wait for reply
		if err := c.SetReadDeadline(time.Now().Add(t.Timeout)); err != nil {
			hop.IP = "*"
			hop.Loss = 100.0
			res.Hops = append(res.Hops, hop)
			continue
		}
		reply := make([]byte, 1500)
		n, peer, err := c.ReadFrom(reply)

		hop := HopInfo{Hop: ttl}

		if err != nil {
			// Timeout
			hop.IP = "*"
			hop.Loss = 100.0
		} else {
			duration := time.Since(start)
			hop.Latency = duration
			hop.IP = peer.String()

			// Parse message to see what we got
			rm, _ := icmp.ParseMessage(ipv4.ICMPTypeEchoReply.Protocol(), reply[:n])
			if rm != nil {
				switch rm.Type {
				case ipv4.ICMPTypeTimeExceeded:
					// Intermediate router
				case ipv4.ICMPTypeEchoReply:
					// Dest reached
				}
			}
		}

		res.Hops = append(res.Hops, hop)

		if hop.IP == dstAddr.String() {
			break
		}
	}

	return res, nil
}
