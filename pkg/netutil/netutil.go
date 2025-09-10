package netutil

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sort"
	"time"
)

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
	return sortAddresses(result), nil
}

// sortAddresses makes results deterministic across runs, ipv4 first
func sortAddresses(ipAddrs []net.IP) []net.IP {
	sort.Slice(ipAddrs, func(i, j int) bool {
		return ipAddrs[i].To4() != nil && ipAddrs[j].To4() == nil
	})
	return ipAddrs
}

// GetOutboundIP returns the preferred outbound IP of this machine. If
// REMOTE_HOST is set, it will attempt to retrieve the address from that host
// first.
func GetOutboundIP() (string, error) {
	if remote := os.Getenv("REMOTE_HOST"); remote != "" {
		resp, err := http.Get(fmt.Sprintf("http://%s/return_ip", remote))
		if err == nil {
			defer resp.Body.Close() //nolint
			data, err := io.ReadAll(resp.Body)
			if err == nil {
				var rr struct {
					IP string `json:"ip"`
				}
				if err = json.Unmarshal(data, &rr); err == nil && rr.IP != "" {
					return rr.IP, nil
				}
			}
		}
		// fall back to local detection on error
	}

	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close() //nolint

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

// GetIPWithExternal attempts to retrieve the IP from an external service.
// If remoteAddr is empty or fails, it falls back to GetOutboundIP.
func GetIPWithExternal(remoteAddr string) (string, error) {
	if remoteAddr != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		url := fmt.Sprintf("http://%s/return_ip", remoteAddr)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
		if err == nil {
			resp, err := http.DefaultClient.Do(req)
			if err == nil {
				defer resp.Body.Close() //nolint:errcheck
				if resp.StatusCode == http.StatusOK {
					var r struct {
						IP string `json:"ip"`
					}
					if err := json.NewDecoder(resp.Body).Decode(&r); err == nil && r.IP != "" {
						return r.IP, nil
					}
				}
			}
		}
	}
	return GetOutboundIP()
}

// GetInterfaces production default, extracted for testing
var GetInterfaces = net.Interfaces
var ListAddrs = func(iface net.Interface) ([]net.Addr, error) {
	return iface.Addrs()
}

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

// GetOutboundP2PAddr builds a libp2p-compatible multiaddress using the
// machine's outbound IP and the provided port.
func GetOutboundP2PAddr(port int) (string, error) {
	ipaddr, err := GetOutboundIP()
	if err != nil {
		return "", fmt.Errorf("failed to get ipaddress: %w", err)
	}
	return fmt.Sprintf("/ip4/%s/tcp/%d", ipaddr, port), nil
}
