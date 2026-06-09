# Golb — A Lightweight Layer 7 Load Balancer

A simple, fast, and configurable HTTP reverse proxy / load balancer written in Go.

## Features

- Round-robin and least-connections load balancing strategies
- Active health checks (HTTP, TCP) with configurable intervals
- Retry on failure with configurable max retries
- YAML-based configuration
- Prometheus metrics endpoint
- Graceful shutdown
- Docker support

## Quick Start

```yaml
# config.yaml
server:
  port: 8080

upstreams:
  - name: app-servers
    strategy: round-robin
    targets:
      - url: http://localhost:3001
      - url: http://localhost:3002
      - url: http://localhost:3003
    health_check:
      path: /health
      interval: 10s
      timeout: 3s
```

```bash
golb --config config.yaml
```

## Installation

```bash
go install github.com/Raphasha27/golb/cmd/golb@latest
```

Or using Docker:

```bash
docker run -v $(pwd)/config.yaml:/etc/golb/config.yaml ghcr.io/raphasha27/golb
```

<br/>

---

<h3 align="center">🐍 Part of the <a href="https://github.com/Raphasha27">Raphasha27</a> Ecosystem</h3>

<p align="center">
  <a href="https://github.com/Raphasha27/Raphasha27">
    <img src="https://img.shields.io/badge/Back_to_Profile-0D1117?style=for-the-badge&logo=github&logoColor=white" />
  </a>
  &nbsp;
  <a href="https://raphasha27.github.io/Raphasha27/ai-snake-game/">
    <img src="https://img.shields.io/badge/▶_Play_AI_Snake-0EA5E9?style=for-the-badge&logo=javascript&logoColor=white" />
  </a>
</p>

## License

MIT
