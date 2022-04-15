// Package automount has the only purpose: to automatically start DNS-DoH proxy and replace default Go resolver
// with resolver returned by dns_proxy.Resolver. This package should only be imported for side effects.
// You can change used options during build using `ldflags`.
package automount

import (
	"context"
	"net"

	"github.com/opentelekomcloud-infra/dns-escape-proxy/proxy"
)

// this is automount defaults
var (
	RemoteDNS = proxy.GoogleDoH
	Port      = 0 // automatically find free port
	Network   = "udp"
)

func init() {
	net.DefaultResolver = proxy.Resolver(context.Background(), Port, Network, RemoteDNS)
}
