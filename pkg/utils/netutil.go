package utils

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sort"
	"time"
)

const (
	// optimumBootstrapURL internal service to get public IP
	optimumBootstrapURL = "https://bootstrap.getoptimum.io"
)

type bootstrapRemoteIP struct {
	IP string `json:"ip"`
}

// ExternalIP returns the first non-loopback IP address available.
func ExternalIP() (string, error) {
	ips, err := ipAddresses()
	if err != nil {
		return "", err
	}
	if len(ips) == 0 {
		return "127.0.0.1", nil
	}
	return ips[0].String(), nil
}

// ipAddresses looks through all the network interfaces
// returns usable, non-loopback IP addresses
func ipAddresses() ([]net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	var result []net.IP
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
				continue
			}
			result = append(result, ip)
		}
	}
	return SortAddresses(result), nil
}

// SortAddresses sorts a set of addresses in the order of ipv4 -> ipv6.
// IPv4 addresses are placed before IPv6 addresses.
func SortAddresses(ipAddrs []net.IP) []net.IP {
	sort.Slice(ipAddrs, func(i, j int) bool {
		return ipAddrs[i].To4() != nil && ipAddrs[j].To4() == nil
	})
	return ipAddrs
}

// GetOutboundIP returns preferred outbound ip of this machine
// then falls back to local detection
func GetOutboundIP() (string, error) {
	url := fmt.Sprintf("%s/api/v1/ip", optimumBootstrapURL)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	resp, code, _ := GetCurl[bootstrapRemoteIP](ctx, url, nil)
	if code == http.StatusOK && resp != nil {
		return resp.IP, nil
	}

	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", fmt.Errorf("dial fallback failed: %w", err)
	}
	defer conn.Close() //nolint:errcheck

	udp, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok || udp.IP == nil {
		return "", fmt.Errorf("unexpected LocalAddr type: %T", conn.LocalAddr())
	}
	return udp.IP.String(), nil
}

var (
	// GetInterfaces production default, extracted for testing
	GetInterfaces = net.Interfaces
	ListAddrs     = func(iface net.Interface) ([]net.Addr, error) {
		return iface.Addrs()
	}
)

// GetPrivateIPs returns a slice of private IPv4 addresses.
func GetPrivateIPs() ([]string, error) {
	privateCIDRs := []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}

	// pre-allocation
	privateNets := make([]*net.IPNet, 0, len(privateCIDRs))
	for _, cidr := range privateCIDRs {
		_, n, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, fmt.Errorf("parse CIDR %q: %w", cidr, err)
		}
		privateNets = append(privateNets, n)
	}

	ifaces, err := GetInterfaces()
	if err != nil {
		return nil, err
	}

	// capacity -> at least one addr per iface
	result := make([]string, 0, len(ifaces))

	for _, iface := range ifaces {
		addrs, err := ListAddrs(iface)
		if err != nil {
			// avoid failing test bc of problematic iface
			continue
		}
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok || ipNet.IP.IsLoopback() || ipNet.IP.To4() == nil {
				continue
			}
			for _, n := range privateNets {
				if n.Contains(ipNet.IP) {
					result = append(result, ipNet.IP.String())
					break
				}
			}
		}
	}
	return result, nil
}

// GetOutboundTCPP2PAddr builds a libp2p-compatible multiaddress using the for tcp transport
// machine's outbound IP and the provided port.
func GetOutboundTCPP2PAddr(port int) (string, error) {
	ipaddr, err := GetOutboundIP()
	if err != nil {
		return "", fmt.Errorf("failed to get ipaddress: %w", err)
	}
	return fmt.Sprintf("/ip4/%s/tcp/%d", ipaddr, port), nil
}

// GetOutboundQUICP2PAddr builds a libp2p-compatible multiaddress for the QUIC transport
// machine's outbound IP and the provided port.
func GetOutboundQUICP2PAddr(port int) (string, error) {
	ipaddr, err := GetOutboundIP()
	if err != nil {
		return "", fmt.Errorf("failed to get ipaddress: %w", err)
	}
	return fmt.Sprintf("/ip4/%s/udp/%d/quic-v1", ipaddr, port), nil
}
