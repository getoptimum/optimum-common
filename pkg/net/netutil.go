package net

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	stdnet "net"
	stdhttp "net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	// optimumBootstrapURL internal service to get public IP
	optimumBootstrapURL = "https://bootstrap.getoptimum.io"
)

// ExternalIPs holds separately-discovered IPv4 and IPv6 public addresses.
type ExternalIPs struct {
	IPv4 string
	IPv6 string
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
func ipAddresses() ([]stdnet.IP, error) {
	ifaces, err := stdnet.Interfaces()
	if err != nil {
		return nil, err
	}
	var result []stdnet.IP
	for _, iface := range ifaces {
		if iface.Flags&stdnet.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&stdnet.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			var ip stdnet.IP
			switch v := addr.(type) {
			case *stdnet.IPNet:
				ip = v.IP
			case *stdnet.IPAddr:
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
func SortAddresses(ipAddrs []stdnet.IP) []stdnet.IP {
	sort.Slice(ipAddrs, func(i, j int) bool {
		return ipAddrs[i].To4() != nil && ipAddrs[j].To4() == nil
	})
	return ipAddrs
}

// GetOutboundIP returns preferred outbound ip of this machine.
func GetOutboundIP() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return DetectIPViaCloudflareTrace(ctx, cloudflareTraceURL, "tcp")
}

// GetExternalIPs discovers IPv4 and IPv6 public addresses separately by
// forcing the address family at the transport layer. Both fields may be
// empty if the corresponding address family is unavailable. At least one
// must succeed or an error is returned.
func GetExternalIPs(ctx context.Context) (*ExternalIPs, error) {
	result := &ExternalIPs{}
	var (
		mu   sync.Mutex
		errs []error
		wg   sync.WaitGroup
	)

	for _, entry := range []struct {
		network string
		target  *string
	}{
		{"tcp4", &result.IPv4},
		{"tcp6", &result.IPv6},
	} {
		wg.Add(1)
		go func(network string, target *string) {
			defer wg.Done()
			ip, err := DetectIPViaCloudflareTrace(ctx, cloudflareTraceURL, network)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, fmt.Errorf("%s: %w", network, err))
				return
			}
			*target = ip
		}(entry.network, entry.target)
	}
	wg.Wait()

	if result.IPv4 == "" && result.IPv6 == "" {
		return nil, fmt.Errorf("no external IPs discovered: %w", errors.Join(errs...))
	}
	return result, nil
}

// newIPTransport creates an http.Transport that forces connections to
// use the given network (e.g. "tcp4", "tcp6", "tcp").
func newIPTransport(network string) *stdhttp.Transport {
	return &stdhttp.Transport{
		DialContext: func(ctx context.Context, _, addr string) (stdnet.Conn, error) {
			return (&stdnet.Dialer{}).DialContext(ctx, network, addr)
		},
	}
}

const cloudflareTraceURL = optimumBootstrapURL + "/cdn-cgi/trace"

// DetectIPViaCloudflareTrace fetches the Cloudflare /cdn-cgi/trace endpoint
// and parses the ip= field. The traceURL parameter specifies the endpoint URL,
// and the network parameter controls the address family ("tcp", "tcp4", or "tcp6").
func DetectIPViaCloudflareTrace(ctx context.Context, traceURL, network string) (string, error) {
	client := &stdhttp.Client{
		Transport: newIPTransport(network),
		Timeout:   10 * time.Second,
	}

	req, err := stdhttp.NewRequestWithContext(ctx, stdhttp.MethodGet, traceURL, nil)
	if err != nil {
		return "", fmt.Errorf("cloudflare trace: create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("cloudflare trace: request failed: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != stdhttp.StatusOK {
		return "", fmt.Errorf("cloudflare trace: unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		return "", fmt.Errorf("cloudflare trace: read body: %w", err)
	}

	return ParseCloudflareTrace(string(body))
}

// ParseCloudflareTrace extracts and validates the ip= field from the
// Cloudflare trace response body (key=value\n format).
func ParseCloudflareTrace(body string) (string, error) {
	scanner := bufio.NewScanner(strings.NewReader(body))
	for scanner.Scan() {
		line := scanner.Text()
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		if key == "ip" {
			ip := strings.TrimSpace(value)
			if stdnet.ParseIP(ip) == nil {
				return "", fmt.Errorf("cloudflare trace: invalid IP %q", ip)
			}
			return ip, nil
		}
	}
	return "", fmt.Errorf("cloudflare trace: ip field not found")
}

var (
	// GetInterfaces production default, extracted for testing
	GetInterfaces = stdnet.Interfaces
	ListAddrs     = func(iface stdnet.Interface) ([]stdnet.Addr, error) {
		return iface.Addrs()
	}
)

// GetPrivateIPs returns a slice of private IPv4 addresses.
func GetPrivateIPs() ([]string, error) {
	privateCIDRs := []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}

	// pre-allocation
	privateNets := make([]*stdnet.IPNet, 0, len(privateCIDRs))
	for _, cidr := range privateCIDRs {
		_, n, err := stdnet.ParseCIDR(cidr)
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
			ipNet, ok := addr.(*stdnet.IPNet)
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

// IsPrivateOrULA checks if an IP address is in a private range (RFC1918) or IPv6 ULA (fc00::/7).
func IsPrivateOrULA(ip stdnet.IP) bool {
	// IPv4 private ranges
	if v4 := ip.To4(); v4 != nil {
		switch {
		case v4[0] == 10:
			return true
		case v4[0] == 172 && v4[1] >= 16 && v4[1] <= 31:
			return true
		case v4[0] == 192 && v4[1] == 168:
			return true
		}
		return false
	}
	// IPv6 unique-local fc00::/7
	return ip.To16() != nil && (ip[0]&0xfe) == 0xfc
}

// IsGlobalUnicast checks if an IP address is a global unicast address.
// Excludes link-local, multicast, and private/ULA addresses.
func IsGlobalUnicast(ip stdnet.IP) bool {
	// Go's IP.IsGlobalUnicast handles most cases well.
	if !ip.IsGlobalUnicast() {
		return false
	}
	// Exclude link-local (169.254/16, fe80::/10) and multicast.
	if ip.IsLinkLocalUnicast() || ip.IsMulticast() {
		return false
	}
	// Exclude RFC1918 & ULA here; those are handled via includePrivate.
	if IsPrivateOrULA(ip) {
		return false
	}
	return true
}

// GetInterfaceIPs returns all interface IP addresses (IPv4 only) that are either
// global unicast or private/ULA addresses. Returns an empty slice on error.
func GetInterfaceIPs() []string {
	result := make([]string, 0)
	interfaces, err := stdnet.Interfaces()
	if err != nil {
		return result
	}
	for _, itf := range interfaces {
		if itf.Flags&stdnet.FlagUp == 0 { // Skip down interfaces
			continue
		}
		addresses, errA := itf.Addrs()
		if errA != nil {
			continue
		}
		for _, a := range addresses {
			var ip stdnet.IP
			switch v := a.(type) {
			case *stdnet.IPNet:
				ip = v.IP
			case *stdnet.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}

			ip = ip.To4()
			if ip == nil {
				continue
				// Maybe IPv6, but we just ignore it for now
			}
			if IsGlobalUnicast(ip) || IsPrivateOrULA(ip) {
				result = append(result, ip.String())
			}
		}
	}
	return result
}
