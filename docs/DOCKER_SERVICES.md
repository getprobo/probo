# Docker Services Documentation

This document describes all services used in the Docker Compose development environment (`compose.yaml`) and their purposes in the Probo application.

## Overview

The development stack includes 10 services organized into the following categories:

- **Core Services**: PostgreSQL, MinIO, Chrome
- **Observability Stack**: Grafana, Prometheus, Loki, Tempo
- **Development Tools**: Mailpit
- **ACME Testing**: Pebble, Pebble Challenge Test Server

## Core Services

### PostgreSQL

**Image**: `postgres:17.4`
**Port**: `5432`
**Purpose**: Primary database for storing all compliance data

PostgreSQL is the main relational database that stores:
- User accounts and organizations
- Compliance frameworks and controls
- Audit trails and evidence
- Documents and policies
- Risk assessments
- Vendor information

#### Configuration

```yaml
Environment:
  POSTGRES_USER: postgres
  POSTGRES_PASSWORD: postgres

Command:
  postgres -c "shared_buffers=4GB"
           -c "max_connections=200"
           -c "log_statement=all"
```

#### Database Setup

On first startup, initialization scripts from `./compose/postgres/` are executed. The database `probod` is created automatically by the application's migration system.

#### Connection Details

- **Host**: `localhost` (or `postgres` from within containers)
- **Port**: `5432`
- **Database**: `probod`
- **Username**: `postgres` / `probod`
- **Password**: `postgres`

#### Access

```bash
# Using psql from host
psql -h localhost -U postgres -d probod

# Using make command
make psql

# From within Docker
docker compose exec postgres psql -U postgres -d probod
```

#### Performance Settings

- **shared_buffers**: 4GB - Memory for caching data
- **max_connections**: 200 - Maximum concurrent connections
- **log_statement**: all - Logs all SQL statements for debugging

---

### MinIO

**Image**: `quay.io/minio/minio`
**Ports**: `9000` (API), `9001` (Console)
**Purpose**: S3-compatible object storage for files and documents

MinIO provides local S3-compatible storage for development, storing:
- Uploaded evidence files
- Generated PDF reports
- Document attachments
- Policy documents
- Exported data

#### Configuration

```yaml
Environment:
  MINIO_ROOT_USER: probod
  MINIO_ROOT_PASSWORD: thisisnotasecret

Command:
  mkdir -p /var/lib/minio/probod &&
  minio server --json --console-address :9001 /var/lib/minio
```

#### Bucket Setup

The bucket `probod` is created automatically on startup via the startup command.

#### Access

**Console UI**: http://localhost:9001
- **Username**: `probod`
- **Password**: `thisisnotasecret`

**API Endpoint**: http://localhost:9000

#### S3 Configuration in Probo

```yaml
aws:
  region: us-east-1
  bucket: probod
  access-key-id: probod
  secret-access-key: thisisnotasecret
  endpoint: http://127.0.0.1:9000
```

#### CLI Access

```bash
# Using AWS CLI
aws --endpoint-url http://localhost:9000 s3 ls s3://probod/

# Using MinIO Client
mc alias set local http://localhost:9000 probod thisisnotasecret
mc ls local/probod
```

---

### Chrome Headless

**Image**: `chromedp/headless-shell:140.0.7259.2`
**Port**: `9222`
**Purpose**: Headless browser for PDF generation and document rendering

Chrome provides browser automation capabilities via the Chrome DevTools Protocol, used for:
- Generating PDF reports from HTML templates
- Rendering documents for preview
- Converting web content to PDF format

#### Configuration

```yaml
Command:
  --headless
  --disable-gpu
  --disable-dev-shm-usage
  --hide-scrollbars
  --mute-audio
  --no-default-browser-check
  --no-first-run
  --disable-background-networking
  --disable-background-timer-throttling
  --disable-extensions
```

#### Chrome DevTools Protocol

The service exposes the Chrome DevTools Protocol on port 9222, which Probo uses through the `chromedp` Go library.

#### Probo Configuration

```yaml
probod:
  chrome-dp-addr: "localhost:9222"
```

#### Testing

```bash
# Check Chrome version
curl http://localhost:9222/json/version

# List available tabs/targets
curl http://localhost:9222/json/list
```

---

## Observability Stack

The observability stack provides comprehensive monitoring, logging, and tracing for development and debugging.

### Grafana

**Image**: `grafana/grafana:latest`
**Port**: `3001`
**Purpose**: Visualization and dashboards for metrics, logs, and traces

Grafana provides a unified interface for:
- Visualizing Prometheus metrics
- Querying and analyzing logs from Loki
- Viewing distributed traces from Tempo
- Creating custom dashboards

#### Configuration

```yaml
Environment:
  GF_AUTH_ANONYMOUS_ENABLED: true
  GF_AUTH_ANONYMOUS_ORG_ROLE: Admin
  GF_AUTH_DISABLE_LOGIN_FORM: true
  GF_USERS_DEFAULT_THEME: light
```

#### Access

**Web UI**: http://localhost:3001

No login required - anonymous access is enabled with Admin role for development.

#### Data Sources

Grafana is pre-configured with:
- **Prometheus** - Metrics at http://prometheus:9191
- **Loki** - Logs at http://loki:3100
- **Tempo** - Traces at http://tempo:4317

Configuration files are in `./compose/grafana/provisioning/`

#### Common Queries

**Metrics (Prometheus)**:
```promql
# Request rate
rate(http_requests_total[5m])

# Error rate
rate(http_requests_total{status=~"5.."}[5m])
```

**Logs (Loki)**:
```logql
{job="probod"} |= "error"
{job="probod"} | json | level="error"
```

---

### Prometheus

**Image**: `prom/prometheus:latest`
**Port**: `9191`
**Purpose**: Metrics collection and storage

Prometheus scrapes and stores time-series metrics from Probo, including:
- HTTP request metrics (rate, duration, status codes)
- Database query metrics
- Business metrics (users, organizations, controls)
- Go runtime metrics (goroutines, memory, GC)

#### Configuration

```yaml
Command:
  --config.file=/etc/prometheus/prometheus.yml
  --storage.tsdb.path=/prometheus
  --web.console.libraries=/etc/prometheus/console_libraries
  --web.console.templates=/etc/prometheus/consoles
  --web.enable-lifecycle
  --web.enable-remote-write-receiver
  --web.listen-address=:9191
```

#### Scrape Configuration

Configuration file: `./compose/prometheus/prometheus.yaml`

```yaml
scrape_configs:
  - job_name: 'probod'
    scrape_interval: 15s
    static_configs:
      - targets: ['host.docker.internal:8081']
```

#### Access

**Web UI**: http://localhost:9191

#### Metrics Endpoint

Probo exposes metrics at: http://localhost:8081/metrics

#### Useful Queries

```promql
# Total requests
sum(rate(http_requests_total[5m]))

# P95 latency
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# Active database connections
pg_connections_active
```

---

### Loki

**Image**: `grafana/loki:latest`
**Port**: `3100`
**Purpose**: Log aggregation and querying

Loki collects and indexes logs from Probo, providing:
- Centralized log storage
- Efficient log querying
- Label-based log filtering
- Integration with Grafana for visualization

#### Configuration

Uses default configuration from `/etc/loki/local-config.yaml` in the container.

#### Log Ingestion

Probo sends structured JSON logs to stdout, which can be collected by:
- Docker logging drivers
- Promtail (Loki's log shipper)
- Direct HTTP API calls

#### Access

**API**: http://localhost:3100

#### Query Examples

```bash
# Query logs via API
curl -G -s "http://localhost:3100/loki/api/v1/query_range" \
  --data-urlencode 'query={job="probod"}' \
  --data-urlencode 'limit=100'

# Query in Grafana
{job="probod"} |= "error" | json
```

#### Log Format

Probo outputs structured JSON logs:
```json
{
  "level": "info",
  "time": "2024-01-01T12:00:00Z",
  "caller": "server/server.go:123",
  "msg": "request completed",
  "method": "GET",
  "path": "/api/v1/controls",
  "status": 200,
  "duration": 45.2,
  "trace_id": "abc123"
}
```

---

### Tempo

**Image**: `grafana/tempo:latest`
**Port**: `4317` (OTLP gRPC)
**Purpose**: Distributed tracing backend

Tempo stores and queries distributed traces, providing:
- End-to-end request tracing
- Service dependency visualization
- Performance bottleneck identification
- Trace-to-log correlation

#### Configuration

Configuration file: `./compose/tempo/tempo.yaml`

#### OpenTelemetry Integration

Probo uses OpenTelemetry instrumentation to send traces to Tempo via OTLP/gRPC protocol.

#### Probo Configuration

```yaml
unit:
  tracing:
    addr: "localhost:4317"
    max-batch-size: 512
    batch-timeout: 5
    export-timeout: 30
    max-queue-size: 2048
```

#### Access

Traces are viewed through Grafana's Explore interface at http://localhost:3001

#### Trace Context

Each trace includes:
- **Trace ID**: Unique identifier for the entire request
- **Span ID**: Identifier for each operation
- **Duration**: How long each operation took
- **Tags**: Metadata (HTTP method, status, user ID, etc.)

#### Correlation

Traces are automatically correlated with:
- **Logs**: Trace ID is included in log entries
- **Metrics**: Exemplars link metrics to traces

---

## Development Tools

### Mailpit

**Image**: `axllent/mailpit:latest`
**Ports**: `1025` (SMTP), `8025` (Web UI)
**Purpose**: Email testing and debugging

Mailpit is an email testing tool that captures all outgoing emails without actually sending them, allowing you to:
- Test email functionality locally
- Preview email templates
- Debug email content and formatting
- Verify email delivery logic

#### Configuration

```yaml
Environment:
  MP_DISABLE_VERSION_CHECK: true
  MP_VERBOSE: false
  MP_SMTP_AUTH_ACCEPT_ANY: true
  MP_ENABLE_PROMETHEUS: true
  MP_SMTP_AUTH_ALLOW_INSECURE: true
```

#### Probo Configuration

```yaml
mailer:
  sender-name: "Probo"
  sender-email: "no-reply@notification.getprobo.com"
  smtp:
    addr: "localhost:1025"
    tls-required: false
```

#### Access

**Web UI**: http://localhost:8025

#### Features

- View all captured emails in a web interface
- Search and filter emails
- View HTML and plain text versions
- Check email headers and attachments
- API access for automated testing

#### Email Types Sent by Probo

- User invitation emails
- Password reset emails
- Audit notification emails
- Report delivery emails
- Task assignment notifications

#### API Access

```bash
# List all messages
curl http://localhost:8025/api/v1/messages

# Get specific message
curl http://localhost:8025/api/v1/message/{id}
```

---

## ACME Testing Services

These services enable local testing of Let's Encrypt certificate provisioning for custom domains.

### Pebble

**Image**: `letsencrypt/pebble:latest`
**Ports**: `14000` (ACME), `15000` (Management)
**Purpose**: ACME protocol test server for Let's Encrypt simulation

Pebble is a small ACME test server that mimics Let's Encrypt, allowing:
- Local testing of ACME certificate provisioning
- Validation of custom domain SSL setup
- Testing certificate renewal logic
- Fast iteration without rate limits

#### Configuration

```yaml
Environment:
  PEBBLE_VA_NOSLEEP: "1"         # Fast validation
  PEBBLE_WFE_NONCEREJECT: "0"    # Allow reused nonces
  PEBBLE_VA_ALWAYS_VALID: "1"    # Skip actual validation

Command:
  pebble -config /test/config/pebble-config.json
         -dnsserver 127.0.0.1:8053
```

Configuration file: `./compose/pebble/pebble-config.json`

#### Probo Configuration

```yaml
custom-domains:
  acme:
    directory: "https://localhost:14000/dir"
    email: "admin@getprobo.com"
    key-type: "EC256"
    insecure-tls: true
```

#### ACME Endpoints

- **Directory**: https://localhost:14000/dir
- **Management API**: http://localhost:15000

#### Certificates

Pebble issues certificates signed by its own CA. The root CA certificate is available at:
`./compose/pebble/certs/rootCA.pem`

This certificate is generated by `mkcert` and must be trusted locally for HTTPS to work.

#### Testing Custom Domains

1. Request certificate from Pebble
2. Complete HTTP-01 or DNS-01 challenge
3. Receive certificate (valid for 90 days)
4. Test certificate renewal

---

### Pebble Challenge Test Server

**Image**: `letsencrypt/pebble-challtestsrv:latest`
**Ports**: `8055` (HTTP-01), `8053` (DNS), `8056` (Management)
**Purpose**: Challenge validation server for ACME testing

This service handles ACME challenge validation:
- **HTTP-01**: Serves challenge responses at `/.well-known/acme-challenge/`
- **DNS-01**: Responds to DNS TXT record queries
- **Management API**: Control challenge responses

#### Configuration

```yaml
Command:
  pebble-challtestsrv -dns01 ":8053"
                      -http01 ":8055"
                      -management ":8056"
```

#### How It Works

1. Probo requests certificate from Pebble
2. Pebble creates a challenge
3. Probo provisions the challenge response
4. Pebble validates by querying this server
5. Certificate is issued if validation succeeds

#### Management API

```bash
# Add HTTP-01 challenge response
curl -X POST http://localhost:8056/add-http01 \
  -d '{"token":"abc", "content":"xyz"}'

# Add DNS-01 TXT record
curl -X POST http://localhost:8056/set-txt \
  -d '{"host":"_acme-challenge.example.com", "value":"abc123"}'
```

---

## Service Dependencies

```
probo (application)
├── depends on: postgres (database)
├── depends on: minio (file storage)
├── depends on: chrome (PDF generation)
├── sends metrics to: prometheus
├── sends logs to: loki
├── sends traces to: tempo
├── sends emails to: mailpit
└── uses for ACME: pebble + pebble-challtestsrv

grafana
├── queries: prometheus (metrics)
├── queries: loki (logs)
└── queries: tempo (traces)
```

## Starting and Stopping Services

### Start All Services

```bash
make stack-up
# or
docker compose up -d
```

### Stop All Services

```bash
make stack-down
# or
docker compose down
```

### View Running Services

```bash
make stack-ps
# or
docker compose ps
```

### View Logs

```bash
# All services
docker compose logs -f

# Specific service
docker compose logs -f postgres
docker compose logs -f minio
```

### Restart Service

```bash
docker compose restart postgres
```

## Volumes and Data Persistence

The following volumes persist data across container restarts:

- `postgres-data` - PostgreSQL database files
- `minio-data` - MinIO object storage
- `grafana-data` - Grafana dashboards and settings
- `prometheus-data` - Prometheus metrics database
- `tempo-data` - Tempo trace storage

### Clearing Data

```bash
# Remove all volumes (WARNING: deletes all data)
docker compose down -v

# Remove specific volume
docker volume rm probo_postgres-data
```

## Network Configuration

All services run on a default Docker Compose network and can communicate using service names:

- From Probo: `postgres:5432`, `minio:9000`, `chrome:9222`
- From Grafana: `prometheus:9191`, `loki:3100`, `tempo:4317`

## Port Summary

| Service | Port(s) | Purpose |
|---------|---------|---------|
| PostgreSQL | 5432 | Database access |
| MinIO | 9000, 9001 | S3 API, Console UI |
| Chrome | 9222 | DevTools Protocol |
| Grafana | 3001 | Web UI |
| Prometheus | 9191 | Metrics API, Web UI |
| Loki | 3100 | Log ingestion API |
| Tempo | 4317 | OTLP trace ingestion |
| Mailpit | 1025, 8025 | SMTP, Web UI |
| Pebble | 14000, 15000 | ACME API, Management |
| Pebble ChalTest | 8053, 8055, 8056 | DNS, HTTP-01, Management |

## Resource Requirements

Minimum recommended resources for development:

- **CPU**: 4 cores
- **Memory**: 8GB RAM
- **Disk**: 20GB free space

Individual service resources:
- PostgreSQL: 2GB RAM (shared_buffers=4GB)
- MinIO: 512MB RAM
- Chrome: 1GB RAM
- Observability stack: ~2GB RAM combined

## Troubleshooting

### PostgreSQL Connection Issues

```bash
# Check if PostgreSQL is running
docker compose ps postgres

# View PostgreSQL logs
docker compose logs postgres

# Test connection
psql -h localhost -U postgres -c "SELECT 1"
```

### MinIO Access Issues

```bash
# Check MinIO health
curl http://localhost:9000/minio/health/live

# List buckets
aws --endpoint-url http://localhost:9000 s3 ls
```

### Chrome Not Responding

```bash
# Check Chrome status
curl http://localhost:9222/json/version

# Restart Chrome
docker compose restart chrome
```

### Observability Stack Issues

```bash
# Check Prometheus targets
curl http://localhost:9191/api/v1/targets

# Check Loki ready status
curl http://localhost:3100/ready

# Check Tempo ready status
curl http://localhost:4317
```

## Security Notes

⚠️ **Development Only**: This Docker Compose setup is for development and should NOT be used in production.

- Default passwords are used (change in production)
- Anonymous access is enabled in Grafana
- TLS is disabled for many services
- No network isolation
- Insecure ACME validation

For production deployment, use:
- Managed database services (AWS RDS, GCP Cloud SQL)
- Managed object storage (AWS S3, GCS)
- Proper authentication and TLS everywhere
- Network policies and firewalls

## Additional Resources

- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [MinIO Documentation](https://min.io/docs/)
- [Chrome DevTools Protocol](https://chromedevtools.github.io/devtools-protocol/)
- [Grafana Documentation](https://grafana.com/docs/)
- [Prometheus Documentation](https://prometheus.io/docs/)
- [Loki Documentation](https://grafana.com/docs/loki/)
- [Tempo Documentation](https://grafana.com/docs/tempo/)
- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [Pebble GitHub](https://github.com/letsencrypt/pebble)
