package proxy

import (
	"net"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/miekg/dns"
)

func assertEquals(t *testing.T, expected, actual interface{}) {
	t.Helper()
	if expected != actual {
		t.Fatalf("Expected %v (%[1]T) but got %v(%[2]T)", expected, actual)
	}
}

func assertNotEmpty(t *testing.T, actual interface{}) {
	t.Helper()
	v := reflect.ValueOf(actual)
	length := v.Len()
	if length < 1 {
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

const defaultDNS = CloudFlareDoH

func TestResolve(t *testing.T) {
	addresses := []string{GoogleDoH, CloudFlareDoH}
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
		}, defaultDNS)
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
		}, defaultDNS)
		assertNoErr(t, err)
		assertEquals(t, true, resp.Response)
		assertEquals(t, 0, len(p.resolutionCache))
	})
}
