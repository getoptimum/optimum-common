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
	"time"
)

const (
	cloudflareTraceURL = "https://cloudflare.com/cdn-cgi/trace"
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
	return DetectIPViaCloudflareTrace(cloudflareTraceURL, "tcp4")
}

// GetExternalIPs discovers IPv4 and IPv6 public addresses separately by
// forcing the address family at the transport layer. Both fields may be
// empty if the corresponding address family is unavailable. At least one
// must succeed or an error is returned.
func GetExternalIPs() (ipV4, ipV6 string, err error) {
	ipTraceEndpoints := []struct {
		traceURL string
		detector func(traceURL, network string) (ip string, err error)
	}{
		{
			traceURL: cloudflareTraceURL,
			detector: DetectIPViaCloudflareTrace,
		},
		{
			traceURL: "https://bootstrap.getoptimum.io/cdn-cgi/trace",
			detector: DetectIPViaCloudflareTrace,
		},
		{
			traceURL: "https://ifconfig.me/ip",
			detector: DetectIPIfConfigTrace,
		},
	}
	errList := make([]error, 0, len(ipTraceEndpoints)*2)
	for _, cnt := range ipTraceEndpoints {
		if ipV4 == "" {
			ipV4, err = cnt.detector(cnt.traceURL, "tcp4")
			if err != nil {
				errList = append(errList, fmt.Errorf("detect ipv4 via %s: %w", cnt.traceURL, err))
			}
		}
		if ipV6 == "" {
			ipV6, err = cnt.detector(cnt.traceURL, "tcp6")
			if err != nil {
				errList = append(errList, fmt.Errorf("detect ipv6 via %s: %w", cnt.traceURL, err))
			}
		}
	}
	if ipV4 != "" || ipV6 != "" {
		return ipV4, ipV6, nil
	}
	return ipV4, ipV6, fmt.Errorf("failed to detect external IPs: %w", errors.Join(errList...))
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

// DetectIPViaCloudflareTrace fetches the Cloudflare /cdn-cgi/trace endpoint
// and parses the ip= field. The traceURL parameter specifies the endpoint URL,
// and the network parameter controls the address family ("tcp", "tcp4", or "tcp6").
func DetectIPViaCloudflareTrace(traceURL, network string) (ip string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return getIPViaTraceURL(ctx, traceURL, network, ParseCloudflareTrace)
}

func DetectIPIfConfigTrace(traceURL, network string) (ip string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return getIPViaTraceURL(ctx, traceURL, network, parseIfConfigTrace)
}

func getIPViaTraceURL(ctx context.Context, traceURL, network string, decoder func(io.Reader) (string, error)) (ip string, err error) {
	client := &stdhttp.Client{
		Transport: newIPTransport(network),
		Timeout:   10 * time.Second,
	}
	_, code, err := GetCurl[any](ctx, traceURL, nil, WithHTTPClient[any](client), WithDecoder[any](func(res io.Reader) error {
		ip, err = decoder(res)
		return err
	}))
	if err != nil {
		return "", fmt.Errorf("request failed, code: %d, err: %w", code, err)
	}
	if code != stdhttp.StatusOK {
		return "", fmt.Errorf("unexpected status %d", code)
	}
	return ip, nil
}

func parseIfConfigTrace(res io.Reader) (string, error) {
	data, err := io.ReadAll(res)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}
	return string(data), nil
}

// ParseCloudflareTrace extracts and validates the ip= field from the
// Cloudflare trace response body (key=value\n format).
func ParseCloudflareTrace(res io.Reader) (string, error) {
	scanner := bufio.NewScanner(res)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "ip=") {
			continue
		}
		_, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		ip := strings.TrimSpace(value)
		if stdnet.ParseIP(ip) == nil {
			return "", fmt.Errorf("cloudflare trace: invalid IP %q", ip)
		}
		return ip, nil
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
