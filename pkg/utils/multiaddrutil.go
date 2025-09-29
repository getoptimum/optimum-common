package utils

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/libp2p/go-libp2p/core/peer"
	b58 "github.com/mr-tron/base58/base58"
	"github.com/multiformats/go-multiaddr"
)

// MultiAddressBuilder takes in an IP address and port to produce libp2p-compatible multiaddresses
func MultiAddressBuilder(ip net.IP, tcpPort int) ([]multiaddr.Multiaddr, error) {
	var ipType string
	switch {
	case ip.To4() != nil:
		ipType = "ip4"
	case ip.To16() != nil:
		ipType = "ip6"
	default:
		return nil, fmt.Errorf("provided IP address is neither IPv4 nor IPv6: %s", ip)
	}

	// Example: /ip4/1.2.3.4./tcp/5678
	multiaddrStr := fmt.Sprintf("/%s/%s/tcp/%d", ipType, ip, tcpPort)
	multiAddrTCP, err := multiaddr.NewMultiaddr(multiaddrStr)
	if err != nil {
		return nil, fmt.Errorf("cannot produce TCP multiaddr format from %s:%d: %w", ip, tcpPort, err)
	}
	return []multiaddr.Multiaddr{multiAddrTCP}, nil
}

// AddressInfoToString converts a peer.AddrInfo into a JSON representation consistent with libp2p logging output
func AddressInfoToString(peer peer.AddrInfo) string {
	data, _ := json.Marshal(peer.Loggable())
	return string(data)
}

// AddressInfoFromString converts a JSON representation produced by AddressInfoToString back to peer.AddrInfo
func AddressInfoFromString(s string) (peer.AddrInfo, error) {
	type tmp struct {
		ID    string                `json:"peerID"`
		Addrs []multiaddr.Multiaddr `json:"addrs"`
	}

	var res tmp
	if err := json.Unmarshal([]byte(s), &res); err != nil {
		return peer.AddrInfo{}, err
	}
	pID, err := b58.Decode(res.ID)
	if err != nil {
		return peer.AddrInfo{}, fmt.Errorf("failed to parse peer ID %q: %w", res.ID, err)
	}
	return peer.AddrInfo{
		ID:    peer.ID(pID),
		Addrs: res.Addrs,
	}, nil
}
