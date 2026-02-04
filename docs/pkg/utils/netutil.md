# Package: utils

**File:** `netutil.go`

## Functions

### ExternalIP

```go
func ExternalIP() (string, error)
```

ExternalIP returns the first non-loopback IP address available.

---

### GetInterfaceIPs

```go
func GetInterfaceIPs() []string
```

GetInterfaceIPs returns all interface IP addresses (IPv4 only) that are either
global unicast or private/ULA addresses. Returns an empty slice on error.

---

### GetOutboundIP

```go
func GetOutboundIP() (string, error)
```

GetOutboundIP returns preferred outbound ip of this machine
then falls back to local detection

---

### GetOutboundQUICP2PAddr

```go
func GetOutboundQUICP2PAddr(port int) (string, error)
```

GetOutboundQUICP2PAddr builds a libp2p-compatible multiaddress for the QUIC transport
machine's outbound IP and the provided port.

---

### GetOutboundTCPP2PAddr

```go
func GetOutboundTCPP2PAddr(port int) (string, error)
```

GetOutboundTCPP2PAddr builds a libp2p-compatible multiaddress using the for tcp transport
machine's outbound IP and the provided port.

---

### GetPrivateIPs

```go
func GetPrivateIPs() ([]string, error)
```

GetPrivateIPs returns a slice of private IPv4 addresses.

---

### IsGlobalUnicast

```go
func IsGlobalUnicast(ip net.IP) bool
```

IsGlobalUnicast checks if an IP address is a global unicast address.
Excludes link-local, multicast, and private/ULA addresses.

---

### IsPrivateOrULA

```go
func IsPrivateOrULA(ip net.IP) bool
```

IsPrivateOrULA checks if an IP address is in a private range (RFC1918) or IPv6 ULA (fc00::/7).

---

### SortAddresses

```go
func SortAddresses(ipAddrs []net.IP) []net.IP
```

SortAddresses sorts a set of addresses in the order of ipv4 -> ipv6.
IPv4 addresses are placed before IPv6 addresses.

---

## Variables

### GetInterfaces

```go
var GetInterfaces = net.Interfaces
```

---

### ListAddrs

```go
var ListAddrs = ...
```
