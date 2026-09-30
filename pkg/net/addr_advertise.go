package net

import (
	"errors"
	"fmt"
	"net"

	"github.com/getoptimum/optimum-common/pkg/logger"
	"github.com/getoptimum/optimum-common/pkg/slices"
	"github.com/multiformats/go-multiaddr"
)

func MustBuildAdvertisedAddresses(
	log logger.AppLogger,
	publicIPV4,
	publicIPV6 string,
	listenPort int,
) []multiaddr.Multiaddr {
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

func BuildAdvertisedAddresses(
	log logger.AppLogger,
	publicIPV4,
	publicIPV6 string,
	listenPort int,
) ([]multiaddr.Multiaddr, error) {
	if listenPort <= 0 || listenPort > 65535 {
		return nil, fmt.Errorf("invalid listenPort: %d", listenPort)
	}
	result, err := publicAddresses(log, publicIPV4, publicIPV6, fmt.Sprintf("tcp/%d", listenPort))
	if err != nil {
		return nil, err
	}

	for _, ipStr := range GetInterfaceIPs() {
		if net.ParseIP(ipStr) == nil { // GetInterfaceIPs returns only ipv4, no need check for ipv6 here
			continue
		}
		ma, errI := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", ipStr, listenPort))
		if errI != nil {
			continue
		}
		result = append(result, ma)
	}
	return slices.UniqueSlice(result), nil
}

func MustBuildAdvertisedQUICAddresses(
	log logger.AppLogger,
	publicIPV4,
	publicIPV6 string,
	listenPort int,
) []multiaddr.Multiaddr {
	mas, err := BuildAdvertisedQUICAddresses(log, publicIPV4, publicIPV6, listenPort)
	if err != nil {
		log.Fatal("failed to build advertised QUIC addresses", err,
			logger.WithString("public_ipv4", publicIPV4),
			logger.WithString("public_ipv6", publicIPV6),
			logger.WithInt("listen_port", listenPort),
		)
	}
	return mas
}

func BuildAdvertisedQUICAddresses(
	log logger.AppLogger,
	publicIPV4,
	publicIPV6 string,
	listenPort int,
) ([]multiaddr.Multiaddr, error) {
	if listenPort <= 0 || listenPort > 65535 {
		return nil, fmt.Errorf("invalid listenPort: %d", listenPort)
	}
	result, err := publicAddresses(log, publicIPV4, publicIPV6, fmt.Sprintf("udp/%d/quic-v1", listenPort))
	if err != nil {
		return nil, err
	}

	for _, ipStr := range GetInterfaceIPs() {
		if net.ParseIP(ipStr) == nil {
			continue
		}
		ma, errI := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/udp/%d/quic-v1", ipStr, listenPort))
		if errI != nil {
			continue
		}
		result = append(result, ma)
	}
	return slices.UniqueSlice(result), nil
}

// publicAddresses builds the public addresses for transport (e.g. "tcp/4001").
// GetExternalIPs leaves publicIPV4 empty on an IPv6-only host, so either
// family alone is enough, but the one that is set must be valid.
func publicAddresses(log logger.AppLogger, publicIPV4, publicIPV6, transport string) ([]multiaddr.Multiaddr, error) {
	if publicIPV4 == "" && publicIPV6 == "" {
		return nil, errors.New("no public IP address to advertise")
	}
	result := make([]multiaddr.Multiaddr, 0, 8)

	if publicIPV4 != "" {
		publicAddressV4, err := multiaddr.NewMultiaddr("/ip4/" + publicIPV4 + "/" + transport)
		if err != nil {
			// advertised address is critical for our application, if it is not valid, there is no reason to continue
			return nil, fmt.Errorf("failed to build advertised address: %w", err)
		}
		result = append(result, publicAddressV4)
	}

	if publicIPV6 != "" {
		publicAddressV6, err := multiaddr.NewMultiaddr("/ip6/" + publicIPV6 + "/" + transport)
		switch {
		case err != nil && publicIPV4 == "":
			return nil, fmt.Errorf("failed to build advertised address: %w", err)
		case err != nil:
			log.Error("failed to build advertised address v6", err, logger.WithString("public_ip", publicIPV6))
		default:
			result = append(result, publicAddressV6)
		}
	}
	return result, nil
}
