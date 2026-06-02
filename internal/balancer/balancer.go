package balancer

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"

	"github.com/Raphasha27/golb/internal/config"
)

type Balancer struct {
	upstreams []*Upstream
	mu        sync.RWMutex
}

type Upstream struct {
	Name    string
	Targets []*Target
	strategy string
	rrIndex  uint64
}

type Target struct {
	URL     *url.URL
	Weight  int
	Healthy bool
	mu      sync.RWMutex
	rev     *httputil.ReverseProxy
}

func New(cfg *config.Config) *Balancer {
	b := &Balancer{}
	for _, uc := range cfg.Upstreams {
		up := &Upstream{
			Name:     uc.Name,
			strategy: uc.Strategy,
		}
		for _, tc := range uc.Targets {
			u, _ := url.Parse(tc.URL)
			t := &Target{
				URL:     u,
				Weight:  tc.Weight,
				Healthy: true,
				rev: httputil.NewSingleHostReverseProxy(u),
			}
			up.Targets = append(up.Targets, t)
		}
		b.upstreams = append(b.upstreams, up)
	}
	return b
}

func (b *Balancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, up := range b.upstreams {
		target := up.next()
		if target == nil {
			http.Error(w, "no healthy targets", http.StatusServiceUnavailable)
			return
		}
		log.Printf("proxying %s -> %s", r.URL.Path, target.URL)
		target.rev.ServeHTTP(w, r)
		return
	}
	http.Error(w, "no upstreams configured", http.StatusNotFound)
}

func (up *Upstream) next() *Target {
	switch up.strategy {
	case "least-connections":
		return up.leastConnections()
	default:
		return up.roundRobin()
	}
}

func (up *Upstream) roundRobin() *Target {
	idx := atomic.AddUint64(&up.rrIndex, 1)
	targets := up.healthyTargets()
	if len(targets) == 0 {
		return nil
	}
	return targets[idx%uint64(len(targets))]
}

func (up *Upstream) leastConnections() *Target {
	targets := up.healthyTargets()
	if len(targets) == 0 {
		return nil
	}
	// Simple: pick first healthy target
	// In production, track active connections per target
	return targets[0]
}

func (up *Upstream) healthyTargets() []*Target {
	var healthy []*Target
	for _, t := range up.Targets {
		t.mu.RLock()
		if t.Healthy {
			healthy = append(healthy, t)
		}
		t.mu.RUnlock()
	}
	return healthy
}

func (t *Target) SetHealthy(ok bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Healthy = ok
}

func (b *Balancer) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	for _, up := range b.upstreams {
		for _, t := range up.Targets {
			t.mu.RLock()
			status := "up"
			if !t.Healthy {
				status = "down"
			}
			w.Write([]byte(up.Name + "_target{url=\"" + t.URL.String() + "\"} " + status + "\n"))
			t.mu.RUnlock()
		}
	}
}
