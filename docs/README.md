# Probod Operations Documentation

This documentation provides operational guidance for deploying, configuring, and managing the Probod compliance management daemon.

## Service Components

Probod requires the following components for operation:

- **probod**: Main service daemon
- **PostgreSQL**: Database backend (version 17+)
- **S3-Compatible Storage**: Object storage (AWS S3, MinIO, etc.)
- **Chrome DevTools Protocol**: PDF generation service

## Deployment Guide

- **[Installation Guide](./INSTALLATION.md)** - Docker and binary deployment procedures
- **[Configuration Reference](./CONFIGURATION.md)** - Complete configuration options and examples
- **[Docker Services](./DOCKER_SERVICES.md)** - Development environment services documentation

## Quick Deployment

### Docker Deployment

```bash
docker run -d \
  --name probod \
  -p 8080:8080 \
  -v /path/to/config.yaml:/etc/probod/config.yaml \
  ghcr.io/getprobo/probo:latest
```

### Service Verification

```bash
# Health check
curl http://localhost:8080/health

# Service logs
docker logs probod
```

## System Requirements

### Minimum Requirements

- **CPU**: 1 core, 2 GHz
- **Memory**: 1 GB RAM
- **Storage**: 10 GB available space
- **Database**: PostgreSQL 12+
- **Storage Backend**: S3-compatible object storage

### Recommended for Production

- **CPU**: 2+ cores, 2.4 GHz
- **Memory**: 4 GB RAM
- **Storage**: 50 GB SSD
- **Database**: Managed PostgreSQL with backups
- **Load Balancer**: For high availability deployments

## External Dependencies

### Required Services

- **PostgreSQL Database**: Primary data storage
- **S3-Compatible Storage**: File and document storage (AWS S3, MinIO, etc.)
- **Chrome/Chromium**: PDF generation via DevTools Protocol

### Optional Integrations

- **SMTP Server**: Email notifications
- **OpenAI API**: AI-powered features
- **OAuth2 Providers**: External service integrations (GitHub, Slack, etc.)

## Production Deployment

### High Availability

- Deploy multiple probod instances behind a load balancer
- Use managed PostgreSQL with read replicas
- Configure distributed object storage
- Implement monitoring and alerting

### Security Configuration

- Enable TLS for all external communications
- Use strong, unique secrets for authentication components
- Configure network security groups and firewalls
- Regular security updates and vulnerability scanning

## Operations

### Monitoring

- **Health Check**: `GET /health`
- **Metrics**: Prometheus-compatible metrics at `/metrics` endpoint
- **Tracing**: OpenTelemetry distributed tracing support
- **Logging**: Structured JSON logging with configurable levels

#### Configuration-Based Observability

Probod provides built-in observability configured through the main configuration file:

- **Prometheus Metrics**: Configure `unit.metrics.addr` for metrics endpoint
- **OpenTelemetry Tracing**: Configure `unit.tracing` section for trace collection
- **Structured Logging**: Automatic JSON logging with correlation IDs

#### Configuration Example

```yaml
unit:
  metrics:
    addr: "0.0.0.0:8081"
  tracing:
    addr: "tempo:4317"
```

See [Configuration Reference](./CONFIGURATION.md#telemetry-and-observability) for complete details.

### Backup and Recovery

- **Database**: Regular PostgreSQL backups with point-in-time recovery
- **Object Storage**: S3 versioning and cross-region replication
- **Configuration**: Version control all configuration files

### Troubleshooting

Common operational issues:

- Verify database connectivity and credentials
- Check S3 storage access permissions
- Validate configuration file syntax
- Review service logs for startup errors

## Support

- **Repository**: [getprobo/probo](https://github.com/getprobo/probo)
- **Issues**: Report operational problems on GitHub
- **Releases**: [GitHub releases](https://github.com/getprobo/probo/releases)
