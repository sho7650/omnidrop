# OmniDrop Deployments

This directory contains deployment configurations for monitoring infrastructure.

## Directory Structure

```
deployments/
├── docker-compose.yml              # Main compose file for monitoring stack
├── prometheus/
│   └── prometheus.yml              # Prometheus configuration
└── grafana/
    ├── dashboards/
    │   └── omnidrop-overview.json  # Pre-built Grafana dashboard
    └── provisioning/
        ├── datasources/
        │   └── prometheus.yml      # Auto-configure Prometheus datasource
        └── dashboards/
            └── omnidrop.yml        # Auto-import dashboard configuration
```

## Quick Start

```bash
# Start monitoring stack
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f

# Stop stack
docker-compose down
```

## Services

- **Prometheus**: http://localhost:9090 - Metrics collection and querying
- **Grafana**: http://localhost:3000 - Dashboard visualization (admin/admin)

## Documentation

See [docs/monitoring.md](../docs/monitoring.md) for complete setup and usage instructions.
