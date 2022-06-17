// Package scan provides types and functions to perform TCP port scan on the
// list of hosts provided
package scan

import (
	"fmt"
	"net"
	"time"
)

// PortState represents the state of single TCP port
type PortState struct {
	Port int
	Open state
}

type state bool

func (s state) String() string {
	if s {
		return "open"
	}
	return "closed"
}

// scanPort performs scan of single TCP port
func scanPort(host string, port int) PortState {
	p := PortState {
		Port: port,
	}
	address := net.JoinHostPort(host, fmt.Sprintf("%d", port))  // try another way
	scanConn, err := net.DialTimeout("tcp", address, 1 * time.Second)
	if err != nil {
		return p
	}
	scanConn.Close()
	p.Open = true

	return p
}

// Results represents the scan result for a single host
type Results struct {
	Host string
	NotFound bool
	PortStates []PortState
}

// Run performs a port scan on a hosts list
func Run(hl *HostsList, ports []int) []Results {
	results := make([]Results, 0, len(hl.Hosts))

	for _, h := range hl.Hosts {
		r := Results {
			Host: h,
		}
		if _, err := net.LookupHost(h); err != nil {
			r.NotFound = true
			results = append(results, r)
			continue
		}
		for _, p := range ports {
			r.PortStates = append(r.PortStates, scanPort(h, p))
		}
		results = append(results, r)
	}

	return results
}
