# dns-escape-proxy

A workaround to escape system DNS restrictions and resolve address using some DoH service.

## Description

Built-in `net.Resolver` supports only UDP and TCP DNS servers. As there are potential firewall limitations on using
arbitrary ports in some environments, we can't really trust they are available.

DoH, on the other hand is served on standard HTTPS port, not a DNS-specific one, so unlikely is a subject of firewall
limitations.

When `proxy.Resolver` is called, we create a UDP/TCP DNS server on a given local port, simply redirecting all input
requests to an external DoH service.

## Example

### Configure HTTP client to use with resolver

```go
package main

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/opentelekomcloud-infra/dns-escape-proxy/proxy"
)

func makeRequest() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // stop service on return
	resolver := proxy.Resolver(ctx, 0, "udp", proxy.GoogleDoH)
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Resolver: resolver,
			}).DialContext,
		},
		Timeout: 2 * time.Second,
	}
	_, err := client.Get("https://open-telekom-cloud.com/en")
	return err
}
```

Using port `0` meaning that random free port will be used.

Note that `Resolver` caches requests, so it makes sense to let it leave as much as possible.

### Configure default HTTP client to use the resolver

```go
package main

import (
	"net/http"

	_ "github.com/opentelekomcloud-infra/dns-escape-proxy/proxy/automount"
)

func makeRequest() error {
	_, err := http.Get("https://open-telekom-cloud.com/en")
	return err
}
```

By default, new resolver is mount on random free port and will use `proxy.GoogleDoH` as DoH service. You can change this
behaviour passing `ldflags` during build, e.g.

```shell
go -ldflags="-X 'github.com/opentelekomcloud-infra/dns-escape-proxy/proxy/automount.Port=12773'" ...
```

## Limitations

This package is not expected to be used on any kind of server, and wasn't tested in such conditions, so potentially can
have problems under heavier load.

This package was tested to be compatible with Google and CloudFlare DoH services
