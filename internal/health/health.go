package health

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Raphasha27/golb/internal/config"
)

type Checker struct {
	cfg      *config.Config
	client   *http.Client
	callback func(name, url string, healthy bool)
}

func NewChecker(cfg *config.Config) *Checker {
	return &Checker{
		cfg: cfg,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (hc *Checker) Start(ctx context.Context) {
	for _, up := range hc.cfg.Upstreams {
		if up.HealthCheck.Path == "" {
			continue
		}
		up := up
		go func() {
			ticker := time.NewTicker(up.HealthCheck.Interval)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					for _, target := range up.Targets {
						hc.check(target.URL + up.HealthCheck.Path)
					}
				}
			}
		}()
	}
}

func (hc *Checker) check(url string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false
	}

	resp, err := hc.client.Do(req)
	if err != nil {
		log.Printf("health check failed for %s: %v", url, err)
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func (hc *Checker) Stop() {}
