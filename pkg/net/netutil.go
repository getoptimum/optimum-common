package net

import (
	"context"
	"fmt"
	"io"
	stdnet "net"
	stdhttp "net/http"
	"sort"
	"strings"
	"time"
)

var BootstrapTraceURL = "https://bootstrap.getoptimum.io/cdn-cgi/trace"

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

func ipAddresses() ([]stdnet.IP, error) {
	ifaces, err := stdnet.Interfaces()
	if err != nil {
		return nil, err
	}
	var result []stdnet.IP
	for _, iface := range ifaces {
		if iface.Flags&stdnet.FlagUp == 0 {
			continue
		}
		if iface.Flags&stdnet.FlagLoopback != 0 {
			continue
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

func SortAddresses(ipAddrs []stdnet.IP) []stdnet.IP {
	sort.Slice(ipAddrs, func(i, j int) bool {
		return ipAddrs[i].To4() != nil && ipAddrs[j].To4() == nil
	})
	return ipAddrs
}

func GetOutboundIP() (string, error) {
	if ip := ipFromCFTrace(); ip != "" {
		return ip, nil
	}
	return ipFromUDPDial()
}

func ipFromCFTrace() string {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := stdhttp.NewRequestWithContext(ctx, stdhttp.MethodGet, BootstrapTraceURL, stdhttp.NoBody)
	if err != nil {
		return ""
	}
	resp, err := stdhttp.DefaultClient.Do(req)
	if err != nil {
		return ""
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024))
	_ = resp.Body.Close()
	if err != nil || resp.StatusCode != stdhttp.StatusOK {
		return ""
	}

	for _, line := range strings.Split(string(body), "\n") {
		if strings.HasPrefix(line, "ip=") {
			ip := strings.TrimSpace(strings.TrimPrefix(line, "ip="))
			if parsed := stdnet.ParseIP(ip); parsed != nil && !IsPrivateOrULA(parsed) {
				return ip
			}
		}
	}
	return ""
}

func ipFromUDPDial() (string, error) {
	conn, err := stdnet.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", fmt.Errorf("dial fallback failed: %w", err)
	}
	defer conn.Close() //nolint:errcheck

	udp, ok := conn.LocalAddr().(*stdnet.UDPAddr)
	if !ok || udp.IP == nil {
		return "", fmt.Errorf("unexpected LocalAddr type: %T", conn.LocalAddr())
	}
	return udp.IP.String(), nil
}

var (
	GetInterfaces = stdnet.Interfaces
	ListAddrs     = func(iface stdnet.Interface) ([]stdnet.Addr, error) {
		return iface.Addrs()
	}
)

func GetPrivateIPs() ([]string, error) {
	privateCIDRs := []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}

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

	result := make([]string, 0, len(ifaces))

	for _, iface := range ifaces {
		addrs, err := ListAddrs(iface)
		if err != nil {
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

func GetOutboundTCPP2PAddr(port int) (string, error) {
	ipaddr, err := GetOutboundIP()
	if err != nil {
		return "", fmt.Errorf("failed to get ipaddress: %w", err)
	}
	return fmt.Sprintf("/ip4/%s/tcp/%d", ipaddr, port), nil
}

func GetOutboundQUICP2PAddr(port int) (string, error) {
	ipaddr, err := GetOutboundIP()
	if err != nil {
		return "", fmt.Errorf("failed to get ipaddress: %w", err)
	}
	return fmt.Sprintf("/ip4/%s/udp/%d/quic-v1", ipaddr, port), nil
}

func IsPrivateOrULA(ip stdnet.IP) bool {
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
	return ip.To16() != nil && (ip[0]&0xfe) == 0xfc
}

func IsGlobalUnicast(ip stdnet.IP) bool {
	if !ip.IsGlobalUnicast() {
		return false
	}
	if ip.IsLinkLocalUnicast() || ip.IsMulticast() {
		return false
	}
	if IsPrivateOrULA(ip) {
		return false
	}
	return true
}

func GetInterfaceIPs() []string {
	result := make([]string, 0)
	interfaces, err := stdnet.Interfaces()
	if err != nil {
		return result
	}
	for _, itf := range interfaces {
		if itf.Flags&stdnet.FlagUp == 0 {
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
			}
			if IsGlobalUnicast(ip) || IsPrivateOrULA(ip) {
				result = append(result, ip.String())
			}
		}
	}
	return result
}
