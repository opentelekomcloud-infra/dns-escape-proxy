package proxy

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func httpClientWithDNS(resolver *net.Resolver) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Resolver: resolver,
			}).DialContext,
		},
		Timeout: 2 * time.Second,
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
			require.NoError(t, err)
			require.True(t, resp.Response)
			assert.NotEmpty(t, resp.Answer)
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
		require.NoError(t, err)
		require.True(t, resp.Response)
		assert.Equal(t, 1, len(p.resolutionCache))
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
		require.NoError(t, err)
		require.True(t, resp.Response)
		assert.Equal(t, 0, len(p.resolutionCache))
	})
}

func TestFunctional(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	r := Resolver(ctx, 0, "udp", defaultDNS)
	h := httpClientWithDNS(r)
	resp, err := h.Get("https://open-telekom-cloud.com/en")
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}
