# Probo Installation Guide

This document provides comprehensive installation instructions for the Probo compliance management platform daemon (`probod`).

## Installation Methods

Probo can be deployed using either the official Docker image or pre-compiled binaries available through GitHub releases.

## Docker Installation

The official Docker images are available on GitHub Container Registry and support multiple architectures:

- **Multi-architecture image**: `ghcr.io/getprobo/probo:latest`
- **AMD64 (x86_64)**: `ghcr.io/getprobo/probo:latest-amd64`
- **ARM64**: `ghcr.io/getprobo/probo:latest-arm64`

### Basic Docker Setup

To run Probo using Docker:

```bash
docker run -d \
  --name probod \
  -p 8080:8080 \
  -v /path/to/config.yaml:/etc/probod/config.yaml \
  ghcr.io/getprobo/probo:latest
```

### Docker Compose Setup

For a complete setup with dependencies, you can use Docker Compose:

```yaml
version: "3.8"

services:
  probod:
    image: ghcr.io/getprobo/probo:latest
    ports:
      - "8080:8080"
    volumes:
      - ./config.yaml:/etc/probod/config.yaml
      - ./data:/data
    environment:
      - PROBOD_CONFIG=/etc/probod/config.yaml
    depends_on:
      - postgres
      - minio

  postgres:
    image: postgres:17.4
    shm_size: 1g
    command: >
      postgres -c "shared_buffers=4GB"
               -c "max_connections=200"
               -c "log_statement=all"
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  minio:
    image: quay.io/minio/minio
    entrypoint: sh
    command: |
      -c 'mkdir -p /var/lib/minio/probod && minio server --json --console-address :9001 /var/lib/minio'
    environment:
      MINIO_ROOT_USER: probod
      MINIO_ROOT_PASSWORD: thisisnotasecret
    volumes:
      - minio_data:/var/lib/minio
    ports:
      - "9000:9000"
      - "9001:9001"

volumes:
  postgres_data:
  minio_data:
```

### Docker Architecture Support

The Docker images support the following architectures:

- **linux/amd64** (x86_64) - Standard 64-bit Intel/AMD processors
- **linux/arm64** - ARM 64-bit processors (Apple Silicon, AWS Graviton, etc.)

Multi-architecture images automatically select the appropriate variant for your platform.

## Binary Installation

Pre-compiled binaries are available for download from the [GitHub releases page](https://github.com/getprobo/probo/releases).

### Supported Platforms

The following platforms are officially supported:

- **Windows**: AMD64 (x86_64)
- **macOS**: AMD64 (x86_64) and ARM64 (Apple Silicon)
- **Linux**: AMD64 (x86_64) and ARM64 (via Docker)

If your specific platform is not available, please contact us, and we may be able to add support for additional architectures.

### Installation Steps

1. **Download the Binary**

   Visit the [GitHub releases page](https://github.com/getprobo/probo/releases) and download the appropriate archive for your platform:
   - **Windows**: `probod_Windows_x86_64.zip`
   - **macOS (Intel)**: `probod_Darwin_x86_64.tar.gz`
   - **macOS (Apple Silicon)**: `probod_Darwin_arm64.tar.gz`

2. **Extract the Archive**

   ```bash
   # For tar.gz files (macOS/Linux)
   tar -xzf probod_Darwin_x86_64.tar.gz

   # For zip files (Windows)
   # Use your preferred extraction tool
   ```

3. **Install the Binary**

   **macOS/Linux:**

   ```bash
   # Move to a directory in your PATH
   sudo mv probod /usr/local/bin/

   # Make executable (if not already)
   sudo chmod +x /usr/local/bin/probod
   ```

   **Windows:**

   ```cmd
   # Move probod.exe to a directory in your PATH
   # Or add the current directory to your PATH environment variable
   ```

4. **Verify Installation**

   ```bash
   probod --version
   ```

### Running the Binary

Once installed, you can run Probo with a configuration file:

```bash
# Using the default configuration location
probod --config /etc/probod/config.yaml

# Or specify a custom configuration file
probod --config ./my-config.yaml
```

## System Requirements

### Minimum Requirements

- **CPU**: 1 core, 2 GHz
- **Memory**: 1 GB RAM
- **Storage**: 10 GB available space
- **Network**: Internet connectivity for external integrations

### Recommended Requirements

- **CPU**: 2+ cores, 2.4 GHz
- **Memory**: 4 GB RAM
- **Storage**: 50 GB available space (SSD preferred)
- **Network**: Stable internet connection

### Dependencies

Probo requires the following external services:

1. **PostgreSQL Database** (version 12 or higher)
2. **S3-Compatible Storage** (AWS S3, MinIO, etc.)
3. **Chrome/Chromium** (for PDF generation via Chrome DevTools Protocol)

Optional dependencies:

- **SMTP Server** (for email notifications)
- **OpenAI API** (for AI-powered features)

## Quick Start

### 1. Database Setup

Create a PostgreSQL database for Probo:

```sql
CREATE DATABASE probod;
CREATE USER postgres WITH PASSWORD 'your_secure_password';
GRANT ALL PRIVILEGES ON DATABASE probod TO postgres;
```

### 2. Configuration

Create a basic configuration file (`config.yaml`):

```yaml
probod:
  hostname: "localhost:8080"
  encryption-key: "your-base64-encoded-encryption-key"

  pg:
    addr: "localhost:5432"
    username: "postgres"
    password: "your_secure_password"
    database: "probod"

  aws:
    region: "us-east-1"
    bucket: "probod"
    access-key-id: "your-access-key"
    secret-access-key: "your-secret-key"
    endpoint: "http://localhost:9000" # For MinIO
```

### 3. Start the Service

```bash
# Using Docker
docker run -d \
  --name probod \
  -p 8080:8080 \
  -v ./config.yaml:/etc/probod/config.yaml \
  ghcr.io/getprobo/probo:latest

# Using Binary
probod --config config.yaml
```

### 4. Access the Application

Open your web browser and navigate to `http://localhost:8080` to access the Probo web interface.

## Production Deployment

### Security Considerations

1. **Use strong, unique secrets** for all authentication components
2. **Enable TLS** for all external communications
3. **Use managed database services** with encryption at rest
4. **Implement proper monitoring** and logging
5. **Regular security updates** and vulnerability assessments

### Load Balancing

For high-availability deployments, consider using:

- **Reverse Proxy**: Nginx, HAProxy, or cloud load balancers
- **Database Clustering**: PostgreSQL with read replicas
- **File Storage**: Distributed S3-compatible storage

### Monitoring

Probo provides metrics and health checks:

- **Health Check**: `GET /health`
- **Metrics**: Prometheus-compatible metrics endpoint
- **Logging**: Structured JSON logging with configurable levels

## Troubleshooting

### Common Installation Issues

1. **Permission Denied (Binary)**

   ```bash
   chmod +x probod
   ```

2. **Database Connection Failed**
   - Verify database credentials and network connectivity
   - Check PostgreSQL is running and accepting connections

3. **Docker Image Pull Failed**

   ```bash
   docker login ghcr.io
   docker pull ghcr.io/getprobo/probo:latest
   ```

4. **Port Already in Use**

   ```bash
   # Find process using port 8080
   lsof -i :8080

   # Or use a different port in configuration
   ```

### Getting Help

- **Documentation**: Check the [configuration reference](./CONFIGURATION.md)
- **GitHub Issues**: Report bugs and request features
- **Community**: Join our community discussions

### Log Analysis

Enable debug logging to troubleshoot issues:

```yaml
probod:
  # ... other configuration
  log-level: debug
```

Check startup logs for configuration validation errors and service initialization issues.
