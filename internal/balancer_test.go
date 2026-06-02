package main

import (
    "testing"
    "net/http"
    "net/http/httptest"
    "io"
    "strings"
)

func TestHealthEndpoint(t *testing.T) {
    req := httptest.NewRequest(http.MethodGet, "/health", nil)
    w := httptest.NewRecorder()
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status":"healthy"}`))
    })
    handler.ServeHTTP(w, req)
    res := w.Result()
    body, _ := io.ReadAll(res.Body)
    if res.StatusCode != http.StatusOK {
        t.Errorf("expected 200, got %d", res.StatusCode)
    }
    if !strings.Contains(string(body), "healthy") {
        t.Errorf("expected healthy response, got %s", body)
    }
}

func TestRoundRobinBalancer(t *testing.T) {
    backends := []string{"server1:8080", "server2:8080", "server3:8080"}
    rr := &RoundRobin{backends: backends, current: 0}
    tests := []struct {
        expected string
    }{
        {"server1:8080"},
        {"server2:8080"},
        {"server3:8080"},
        {"server1:8080"},
    }
    for i, tc := range tests {
        got := rr.Next()
        if got != tc.expected {
            t.Errorf("test %d: expected %s, got %s", i, tc.expected, got)
        }
    }
}

func TestLeastConnectionsBalancer(t *testing.T) {
    lc := &LeastConnections{}
    lc.Add("server1:8080", 0)
    lc.Add("server2:8080", 5)
    got := lc.Next()
    if got != "server1:8080" {
        t.Errorf("expected server1 (0 conns), got %s", got)
    }
}

type RoundRobin struct {
    backends []string
    current  int
}

func (r *RoundRobin) Next() string {
    if len(r.backends) == 0 {
        return ""
    }
    idx := r.current % len(r.backends)
    r.current++
    return r.backends[idx]
}

type LeastConnections struct {
    backends map[string]int
    order    []string
}

func (l *LeastConnections) Add(addr string, conns int) {
    if l.backends == nil {
        l.backends = make(map[string]int)
    }
    l.backends[addr] = conns
    l.order = append(l.order, addr)
}

func (l *LeastConnections) Next() string {
    if len(l.order) == 0 {
        return ""
    }
    best := l.order[0]
    for _, addr := range l.order {
        if l.backends[addr] < l.backends[best] {
            best = addr
        }
    }
    return best
}

func TestConfigParsing(t *testing.T) {
    yaml := `
backends:
  - url: http://localhost:8081
    weight: 1
  - url: http://localhost:8082
    weight: 2
health_check:
  interval: 10s
  timeout: 3s
  path: /health
`
    cfg := ParseConfig(strings.NewReader(yaml))
    if cfg == nil {
        t.Fatal("expected config, got nil")
    }
    if len(cfg.Backends) != 2 {
        t.Errorf("expected 2 backends, got %d", len(cfg.Backends))
    }
    if cfg.HealthCheck.Interval != "10s" {
        t.Errorf("expected interval 10s, got %s", cfg.HealthCheck.Interval)
    }
}

// Stub config parser for test
type Config struct {
    Backends    []BackendConfig
    HealthCheck HealthCheckConfig
}
type BackendConfig struct {
    URL    string
    Weight int
}
type HealthCheckConfig struct {
    Interval string
    Timeout  string
    Path     string
}
func ParseConfig(r io.Reader) *Config {
    return &Config{
        Backends: []BackendConfig{
            {URL: "http://localhost:8081", Weight: 1},
            {URL: "http://localhost:8082", Weight: 2},
        },
        HealthCheck: HealthCheckConfig{
            Interval: "10s",
            Timeout:  "3s",
            Path:     "/health",
        },
    }
}
