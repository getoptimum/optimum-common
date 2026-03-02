package net

import (
	"fmt"
	"net"

	"github.com/getoptimum/optimum-common/pkg/logger"
	commonSlices "github.com/getoptimum/optimum-common/pkg/slices"
	"github.com/multiformats/go-multiaddr"
)

func MustBuildAdvertisedAddresses(log logger.AppLogger, publicIPV4, publicIPV6 string, listenPort int) []multiaddr.Multiaddr {
	mas, err := BuildAdvertisedAddresses(log, publicIPV4, publicIPV6, listenPort)
	if err != nil {
		log.Fatal("failed to build advertised addresses", err,
			logger.WithString("public_ipv4", publicIPV4),
			logger.WithString("public_ipv6", publicIPV6),
			logger.WithInt("listen_port", listenPort),
		)
	}
	return mas
}

func BuildAdvertisedAddresses(log logger.AppLogger, publicIPV4, publicIPV6 string, listenPort int) ([]multiaddr.Multiaddr, error) {
	if listenPort <= 0 || listenPort > 65535 {
		return nil, fmt.Errorf("invalid listenPort: %d", listenPort)
	}
	result := make([]multiaddr.Multiaddr, 0, 8)

	publicAddressV4, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", publicIPV4, listenPort))
	if err != nil {
		// advertised address is critical for our application, if it is not valid, there is no reason to continue
		return nil, fmt.Errorf("failed to build advertised address: %w", err)
	}
	result = append(result, publicAddressV4)

	if publicIPV6 != "" {
		publicAddressV6, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip6/%s/tcp/%d", publicIPV6, listenPort))
		if err != nil {
			log.Error("failed to build advertised address v6", err, logger.WithString("public_ip", publicIPV6))
		} else {
			result = append(result, publicAddressV6)
		}
	}

	for _, ipStr := range GetInterfaceIPs() {
		ip := net.ParseIP(ipStr)
		if ip == nil {
			continue
		}

		var ma multiaddr.Multiaddr
		if ip.To4() != nil {
			ma, err = multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", ipStr, listenPort))
		} else {
			ma, err = multiaddr.NewMultiaddr(fmt.Sprintf("/ip6/%s/tcp/%d", ipStr, listenPort))
		}
		if err != nil {
			continue
		}
		result = append(result, ma)
	}
	return commonSlices.UniqueSlice(result), nil
}
