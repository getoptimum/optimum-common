# Package: utils

**File:** `multiaddrutil.go`

## Functions

### AddressInfoFromString

```go
func AddressInfoFromString(s string) (peer.AddrInfo, error)
```

AddressInfoFromString converts a JSON representation produced by AddressInfoToString back to peer.AddrInfo

---

### AddressInfoToString

```go
func AddressInfoToString(localPeer peer.AddrInfo) string
```

AddressInfoToString converts a peer.AddrInfo into a JSON representation consistent with libp2p logging output

---

### GetIPProtocol

```go
func GetIPProtocol(ipStr string) string
```

GetIPProtocol returns the multiaddr protocol string ("ip4" or "ip6") for the given IP address string.
Returns "ip4" for IPv4 addresses and "ip6" for IPv6 addresses.
If the IP string cannot be parsed, it defaults to "ip4".

---

### MultiAddressBuilder

```go
func MultiAddressBuilder(ip net.IP, tcpPort int) ([]multiaddr.Multiaddr, error)
```

MultiAddressBuilder takes in an IP address and port to produce libp2p-compatible multiaddresses
