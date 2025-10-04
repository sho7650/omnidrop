# OmniDrop Monitoring with Prometheus and Grafana

This guide explains how to set up and use Prometheus and Grafana for monitoring the OmniDrop API server.

## Overview

OmniDrop exposes Prometheus metrics at the `/metrics` endpoint. The monitoring stack includes:

- **Prometheus**: Time-series database for metrics collection and storage
- **Grafana**: Visualization dashboard for metrics analysis
- **Docker Compose**: Simplified deployment for local development

## Quick Start

### 1. Start the Monitoring Stack

```bash
# Navigate to deployments directory
cd deployments

# Start Prometheus and Grafana
docker-compose up -d

# Verify services are running
docker-compose ps
```

### 2. Start OmniDrop Server

```bash
# Make sure OmniDrop is running on port 8787 (production)
# or update prometheus.yml if using a different port
TOKEN="your-secret-token" ./omnidrop-server
```

### 3. Access Dashboards

- **Grafana**: http://localhost:3000
  - Default credentials: `admin` / `admin`
  - Pre-configured dashboard: "OmniDrop Overview"

- **Prometheus**: http://localhost:9090
  - Query interface for exploring metrics
  - Target status: http://localhost:9090/targets

## Available Metrics

### HTTP Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `omnidrop_http_requests_total` | Counter | Total HTTP requests by method, endpoint, status |
| `omnidrop_http_request_duration_seconds` | Histogram | Request latency by method, endpoint, status |
| `omnidrop_http_request_size_bytes` | Histogram | Request payload size |
| `omnidrop_http_response_size_bytes` | Histogram | Response payload size |

### Business Metrics - Tasks

| Metric | Type | Description |
|--------|------|-------------|
| `omnidrop_task_creations_total` | Counter | Task creation attempts (success/failure) |
| `omnidrop_task_creation_duration_seconds` | Histogram | Task creation duration |
| `omnidrop_tasks_with_project_total` | Counter | Tasks with project assignment |
| `omnidrop_tasks_with_tags_total` | Counter | Tasks with tags |

### Business Metrics - Files

| Metric | Type | Description |
|--------|------|-------------|
| `omnidrop_file_creations_total` | Counter | File creation attempts (success/failure) |
| `omnidrop_file_creation_duration_seconds` | Histogram | File creation duration |
| `omnidrop_files_size_bytes` | Histogram | Size of created files |

### AppleScript Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `omnidrop_applescript_executions_total` | Counter | AppleScript executions (success/failure) |
| `omnidrop_applescript_execution_duration_seconds` | Histogram | AppleScript execution time |
| `omnidrop_applescript_errors_total` | Counter | AppleScript errors by type |

## Grafana Dashboard

### Pre-configured Panels

The "OmniDrop Overview" dashboard includes:

1. **Summary Stats** (Top Row)
   - Request Rate (req/s)
   - P95 Latency
   - Success Rate (%)
   - Task Creation Rate

2. **HTTP Performance**
   - Requests by Endpoint (time series)
   - Latency by Endpoint (P50/P95/P99)

3. **Business Operations**
   - Task Creation Success/Failure (stacked area)
   - Task Creation Duration (percentiles)

4. **AppleScript Performance**
   - Execution Rate (success/failure)
   - Execution Duration (percentiles)
   - Errors by Type

### Customizing Dashboards

1. Navigate to http://localhost:3000
2. Login with admin credentials
3. Go to "Dashboards" → "OmniDrop" folder
4. Click on "OmniDrop Overview"
5. Use "Edit" to modify panels or create new ones

## Prometheus Queries

### Useful PromQL Examples

**Request rate by endpoint:**
```promql
sum(rate(omnidrop_http_requests_total[5m])) by (endpoint)
```

**P95 latency:**
```promql
histogram_quantile(0.95, sum(rate(omnidrop_http_request_duration_seconds_bucket[5m])) by (le, endpoint))
```

**Success rate:**
```promql
sum(rate(omnidrop_http_requests_total{status=~"2.."}[5m])) 
/ 
sum(rate(omnidrop_http_requests_total[5m])) * 100
```

**Task creation error rate:**
```promql
sum(rate(omnidrop_task_creations_total{status="failure"}[5m]))
```

**AppleScript errors:**
```promql
sum(rate(omnidrop_applescript_errors_total[5m])) by (error_type)
```

## Configuration

### Prometheus Configuration

Edit `deployments/prometheus/prometheus.yml` to:

- Change scrape interval (default: 10s)
- Add multiple OmniDrop instances
- Configure retention period (default: 15 days)
- Add alerting rules

Example for multiple instances:
```yaml
static_configs:
  - targets: 
      - 'localhost:8787'
    labels:
      instance: 'production'
  - targets:
      - 'localhost:8788'
    labels:
      instance: 'development'
```

### Grafana Configuration

Edit `deployments/docker-compose.yml` environment variables:

```yaml
environment:
  - GF_SECURITY_ADMIN_PASSWORD=your-secure-password
  - GF_SERVER_ROOT_URL=http://your-domain:3000
```

## Production Deployment

### Security Considerations

1. **Change default passwords:**
   ```yaml
   # In docker-compose.yml
   - GF_SECURITY_ADMIN_PASSWORD=strong-random-password
   ```

2. **Restrict metrics endpoint (optional):**
   - Add authentication to `/metrics` endpoint
   - Use Prometheus bearer token authentication
   - Implement IP whitelisting

3. **Use persistent volumes:**
   ```bash
   # Backup Prometheus data
   docker run --rm -v omnidrop-prometheus-data:/data -v $(pwd):/backup \
     alpine tar czf /backup/prometheus-backup.tar.gz -C /data .
   ```

### High Availability

For production setups:

1. **External Prometheus:**
   - Use managed Prometheus service (e.g., AWS Managed Prometheus, GCP Monitoring)
   - Configure remote write in `prometheus.yml`

2. **External Grafana:**
   - Use Grafana Cloud or self-hosted cluster
   - Import dashboard JSON from `deployments/grafana/dashboards/`

## Troubleshooting

### Prometheus Not Scraping

1. Check target status: http://localhost:9090/targets
2. Verify OmniDrop is running: `curl http://localhost:8787/health`
3. Test metrics endpoint: `curl http://localhost:8787/metrics`
4. Check Prometheus logs: `docker-compose logs prometheus`

### Grafana Dashboard Not Loading

1. Verify Prometheus datasource: http://localhost:3000/datasources
2. Test query in Explore: http://localhost:3000/explore
3. Check Grafana logs: `docker-compose logs grafana`
4. Reimport dashboard JSON from `deployments/grafana/dashboards/`

### No Metrics Data

1. Ensure OmniDrop server is receiving traffic
2. Check time range in Grafana (default: last 1 hour)
3. Verify Prometheus retention period
4. Test with simple query: `up{job="omnidrop"}`

## Stopping the Stack

```bash
# Stop services (preserves data)
docker-compose down

# Stop and remove all data
docker-compose down -v
```

## Advanced Topics

### Custom Alerts

Create alert rules in `deployments/prometheus/alerts/omnidrop.yml`:

```yaml
groups:
  - name: omnidrop
    interval: 30s
    rules:
      - alert: HighErrorRate
        expr: |
          sum(rate(omnidrop_http_requests_total{status=~"5.."}[5m])) 
          / 
          sum(rate(omnidrop_http_requests_total[5m])) > 0.05
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value | humanizePercentage }}"
```

### Recording Rules

Optimize frequently-used queries in `deployments/prometheus/recording_rules/omnidrop.yml`:

```yaml
groups:
  - name: omnidrop_aggregations
    interval: 30s
    rules:
      - record: omnidrop:http_requests:rate5m
        expr: sum(rate(omnidrop_http_requests_total[5m])) by (endpoint, status)
      
      - record: omnidrop:http_latency:p95
        expr: histogram_quantile(0.95, sum(rate(omnidrop_http_request_duration_seconds_bucket[5m])) by (le, endpoint))
```

## References

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [PromQL Basics](https://prometheus.io/docs/prometheus/latest/querying/basics/)
- [Grafana Dashboard Best Practices](https://grafana.com/docs/grafana/latest/best-practices/best-practices-for-creating-dashboards/)
