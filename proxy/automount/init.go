// Package automount has the only purpose: to automatically start DNS-DoH proxy and replace default Go resolver
// with resolver returned by dns_proxy.ProxyResolver. This package should only be imported for side effects.
// You can change used options during build using `ldflags`.
package automount

import (
	"context"
	"net"

	proxy "github.com/opentelekomcloud-infra/dns-escape-proxy/proxy"
)

// this is automount defaults
var (
	RemoteDNS = "https://dns.google/dns-query"
	Port      = 12332
	Network   = "udp"
)

func init() {
	net.DefaultResolver = proxy.ProxyResolver(context.Background(), Port, Network, RemoteDNS)
}
