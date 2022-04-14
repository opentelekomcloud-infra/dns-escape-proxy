// This is Trojan horse to change DNS resolution to be universally correct independent of system configuration

package proxy

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	rdns "github.com/folbricht/routedns"
	"github.com/miekg/dns"
)

const (
	CloudFlareDoH = "https://cloudflare-dns.com/dns-query"
	GoogleDoH     = "https://dns.google/dns-query"
)

type Proxy struct {
	server *rdns.DNSListener

	resolutionCache map[string]*dns.Msg
	cacheLock       sync.RWMutex
}

// Serve starts server and redirects all inputs to the target DNS
// This function blocks until server is stopped
func (p *Proxy) Serve(port int, network, remoteDNS string) error {
	if p.server != nil {
		return fmt.Errorf("server already exist")
	}
	p.server = &rdns.DNSListener{
		Server: &dns.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Net:     network,
			Handler: p.handler(remoteDNS),
		},
	}
	return p.server.ListenAndServe()
}

func (p *Proxy) handler(remoteDNS string) dns.HandlerFunc {
	return func(w dns.ResponseWriter, msg *dns.Msg) {
		var err error
		defer func() { // log error we faced, if any
			_ = w.Close()
			w.Hijack()
			if err != nil {
				log.Println(err)
				return
			}
		}()
		var resp *dns.Msg
		resp, err = p.resolve(msg, remoteDNS)
		if err != nil {
			return
		}
		err = w.WriteMsg(resp)
	}
}

// getCached reads from cache with rw lock
func (p *Proxy) getCached(key string) *dns.Msg {
	p.cacheLock.RLock()
	defer p.cacheLock.RUnlock()
	return p.resolutionCache[key]
}

// getCached reads from cache with rw lock
func (p *Proxy) saveCached(key string, value *dns.Msg) {
	p.cacheLock.Lock()
	defer p.cacheLock.Unlock()
	if p.resolutionCache == nil {
		p.resolutionCache = make(map[string]*dns.Msg)
	}

	p.resolutionCache[key] = value
}

// resolve given address
func (p *Proxy) resolve(msg *dns.Msg, remoteDNS string) (*dns.Msg, error) {
	cacheKey := msg.String()
	if a := p.getCached(cacheKey); a != nil {
		return a, nil
	}
	clientOpts := rdns.DoHClientOptions{
		Method: http.MethodGet,
	}
	clientID := fmt.Sprintf("doh-%s", remoteDNS)
	client, err := rdns.NewDoHClient(clientID, remoteDNS, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("error creating a client: %w", err)
	}
	result, err := client.ResolvePOST(msg)
	if err != nil {
		return nil, fmt.Errorf("error resolving DNS query: %w", err)
	}
	if len(result.Answer) > 0 { // don't cache empty responses
		p.saveCached(cacheKey, result)
	}
	return result, nil
}

// ProxyResolver starts DNS server and returns resolver using it
func ProxyResolver(ctx context.Context, port int, network string, remoteDNS string) *net.Resolver {
	p := new(Proxy)
	go func() {
		err := p.Serve(port, network, remoteDNS)
		if err != nil {
			log.Fatalf("error serving DNS: %v", err) // exit 1, no sense to serve anymore
		}
	}()

	// stop server on context cancel
	go func() {
		for {
			select {
			case <-ctx.Done():
				if err := ctx.Err(); err != nil {
					log.Fatal(err)
				}
			}
		}
	}()

	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(300),
			}
			return d.DialContext(ctx, network, fmt.Sprintf("localhost:%d", port))
		},
	}
}
