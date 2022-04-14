package dns_proxy

import (
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/miekg/dns"
)

func assertEquals(t *testing.T, expected, actual any) {
	t.Helper()
	if expected != actual {
		t.Fatalf("Expected %v (%[1]T) but got %v(%[2]T)", expected, actual)
	}
}

func assertNotEmpty[T any](t *testing.T, actual []T) {
	t.Helper()
	if len(actual) < 1 {
		t.Fatal("Expected non-empty slice")
	}
}

func assertNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Got unexpected error: %v (%[1]T)", err)
	}
}

func httpClientWithDNS(resolver *net.Resolver) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				Resolver:  resolver,
			}).DialContext,
		},
		Timeout: 3 * time.Millisecond,
	}
}

func TestResolve(t *testing.T) {
	addresses := []string{"https://cloudflare-dns.com/dns-query", "https://dns.google/dns-query"}
	for _, address := range addresses {
		p := &Proxy{}
		t.Run(address, func(t *testing.T) {
			resp, err := p.resolve(&dns.Msg{
				Question: []dns.Question{
					{
						Name:   "google.com.",
						Qtype:  dns.TypeA,
						Qclass: dns.ClassINET,
					},
				},
			}, address)
			assertNoErr(t, err)
			assertEquals(t, true, resp.Response)
			assertNotEmpty(t, resp.Answer)
		})
	}
}

func TestCacheUsed(t *testing.T) {
	address := "https://cloudflare-dns.com/dns-query"

	t.Run("existing", func(t *testing.T) {
		p := &Proxy{}

		resp, err := p.resolve(&dns.Msg{
			Question: []dns.Question{
				{
					Name:   "google.com.",
					Qtype:  dns.TypeA,
					Qclass: dns.ClassINET,
				},
			},
		}, address)
		assertNoErr(t, err)
		assertEquals(t, true, resp.Response)
		assertEquals(t, 1, len(p.resolutionCache))
	})

	t.Run("non-existing", func(t *testing.T) {
		p := &Proxy{}

		resp, err := p.resolve(&dns.Msg{
			Question: []dns.Question{
				{
					Name:   "jasdfasdflk.com.",
					Qtype:  dns.TypeA,
					Qclass: dns.ClassINET,
				},
			},
		}, address)
		assertNoErr(t, err)
		assertEquals(t, true, resp.Response)
		assertEquals(t, 0, len(p.resolutionCache))
	})
}
