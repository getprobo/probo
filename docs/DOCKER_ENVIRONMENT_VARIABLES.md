# Environment Variables Reference

This document provides a comprehensive reference for all environment variables used by the Probo entrypoint script to generate the configuration file.

## Configuration File

| Variable      | Description                    | Default Value            | Required |
| ------------- | ------------------------------ | ------------------------ | -------- |
| `CONFIG_FILE` | Path to the configuration file | `/etc/probod/config.yml` | No       |

## Observability

### Metrics

| Variable       | Description                             | Default Value    | Required |
| -------------- | --------------------------------------- | ---------------- | -------- |
| `METRICS_ADDR` | Address for Prometheus metrics endpoint | `localhost:8081` | No       |

### Tracing

| Variable                 | Description                                             | Default Value    | Required |
| ------------------------ | ------------------------------------------------------- | ---------------- | -------- |
| `TRACING_ADDR`           | OpenTelemetry collector address for distributed tracing | `localhost:4317` | No       |
| `TRACING_MAX_BATCH_SIZE` | Maximum number of spans to batch before export          | `512`            | No       |
| `TRACING_BATCH_TIMEOUT`  | Timeout in seconds for batching spans                   | `5`              | No       |
| `TRACING_EXPORT_TIMEOUT` | Timeout in seconds for exporting traces                 | `30`             | No       |
| `TRACING_MAX_QUEUE_SIZE` | Maximum queue size for spans waiting to be exported     | `2048`           | No       |

## Application Configuration

| Variable                | Description                                                      | Default Value           | Required |
| ----------------------- | ---------------------------------------------------------------- | ----------------------- | -------- |
| `PROBOD_BASE_URL`       | Public hostname for the Probo instance (used for URL generation) | `http://localhost:8080` | No       |
| `PROBOD_ENCRYPTION_KEY` | Base64-encoded encryption key for sensitive data (32+ bytes)     | -                       | **Yes**  |
| `CHROME_DP_ADDR`        | Chrome DevTools Protocol address for PDF generation              | `localhost:9222`        | No       |

## API Configuration

| Variable                   | Description                                    | Default Value           | Required |
| -------------------------- | ---------------------------------------------- | ----------------------- | -------- |
| `API_ADDR`                 | Address and port for the API server to bind to | `:8080`                 | No       |
| `API_CORS_ALLOWED_ORIGINS` | Comma-separated list of allowed CORS origins   | `http://localhost:8080` | No       |

## PostgreSQL Database

| Variable       | Description                                        | Default Value    | Required |
| -------------- | -------------------------------------------------- | ---------------- | -------- |
| `PG_ADDR`      | PostgreSQL server address and port                 | `localhost:5432` | No       |
| `PG_USERNAME`  | PostgreSQL username                                | `postgres`       | No       |
| `PG_PASSWORD`  | PostgreSQL password                                | `postgres`       | No       |
| `PG_DATABASE`  | PostgreSQL database name                           | `probod`         | No       |
| `PG_POOL_SIZE` | Maximum number of connections in the database pool | `100`            | No       |

## Authentication

### User Authentication

| Variable                         | Description                                   | Default Value   | Required |
| -------------------------------- | --------------------------------------------- | --------------- | -------- |
| `AUTH_DISABLE_SIGNUP`            | Disable user self-registration                | `false`         | No       |
| `AUTH_INVITATION_TOKEN_VALIDITY` | Invitation token validity duration in seconds | `3600` (1 hour) | No       |

### Authentication Cookies

| Variable               | Description                                        | Default Value | Required |
| ---------------------- | -------------------------------------------------- | ------------- | -------- |
| `AUTH_COOKIE_NAME`     | Name of the session cookie                         | `SSID`        | No       |
| `AUTH_COOKIE_DOMAIN`   | Domain for the session cookie                      | `localhost`   | No       |
| `AUTH_COOKIE_SECRET`   | Secret key for signing session cookies (32+ bytes) | -             | **Yes**  |
| `AUTH_COOKIE_DURATION` | Session cookie validity duration in hours          | `24`          | No       |
| `AUTH_COOKIE_SECURE`   | Set Secure flag on cookies (use false for HTTP)    | `true`        | No       |

### Password Security

| Variable                   | Description                                          | Default Value | Required |
| -------------------------- | ---------------------------------------------------- | ------------- | -------- |
| `AUTH_PASSWORD_PEPPER`     | Secret pepper value for password hashing (32+ bytes) | -             | **Yes**  |
| `AUTH_PASSWORD_ITERATIONS` | Number of PBKDF2 iterations for password hashing     | `1000000`     | No       |

## Trust Center Authentication

| Variable                         | Description                                            | Default Value           | Required |
| -------------------------------- | ------------------------------------------------------ | ----------------------- | -------- |
| `TRUST_AUTH_COOKIE_NAME`         | Name of the trust center token cookie                  | `TCT`                   | No       |
| `TRUST_AUTH_COOKIE_DOMAIN`       | Domain for the trust center cookie                     | `localhost`             | No       |
| `TRUST_AUTH_COOKIE_DURATION`     | Trust center cookie validity duration in hours         | `24`                    | No       |
| `TRUST_AUTH_TOKEN_DURATION`      | Trust center access token validity duration in hours   | `168` (7 days)          | No       |
| `TRUST_AUTH_REPORT_URL_DURATION` | Validity duration for report URLs in minutes           | `15`                    | No       |
| `TRUST_AUTH_TOKEN_SECRET`        | Secret key for signing trust center tokens (32+ bytes) | -                       | **Yes**  |
| `TRUST_AUTH_SCOPE`               | OAuth2 scope for trust center access                   | `trust_center_readonly` | No       |
| `TRUST_AUTH_TOKEN_TYPE`          | Token type identifier for trust center tokens          | `trust_center_access`   | No       |

## AWS / S3 Storage

| Variable                | Description                                              | Default Value | Required |
| ----------------------- | -------------------------------------------------------- | ------------- | -------- |
| `AWS_REGION`            | AWS region for S3 storage                                | `us-east-1`   | No       |
| `AWS_BUCKET`            | S3 bucket name for file storage                          | `probod`      | No       |
| `AWS_ACCESS_KEY_ID`     | AWS access key ID (leave empty for IAM role)             | -             | No       |
| `AWS_SECRET_ACCESS_KEY` | AWS secret access key (leave empty for IAM role)         | -             | No       |
| `AWS_ENDPOINT`          | Custom S3 endpoint (for MinIO or S3-compatible services) | -             | No       |

## Notifications

### Email (SMTP)

| Variable              | Description                                    | Default Value                        | Required |
| --------------------- | ---------------------------------------------- | ------------------------------------ | -------- |
| `MAILER_SENDER_NAME`  | Display name for outgoing emails               | `Probo`                              | No       |
| `MAILER_SENDER_EMAIL` | Email address for outgoing emails              | `no-reply@notification.getprobo.com` | No       |
| `SMTP_ADDR`           | SMTP server address and port                   | `localhost:1025`                     | No       |
| `SMTP_TLS_REQUIRED`   | Require TLS for SMTP connections               | `false`                              | No       |
| `MAILER_INTERVAL`     | Interval in seconds for processing email queue | `60`                                 | No       |

### Slack

| Variable                | Description                                                 | Default Value | Required |
| ----------------------- | ----------------------------------------------------------- | ------------- | -------- |
| `SLACK_SENDER_INTERVAL` | Interval in seconds for processing Slack notification queue | `60`          | No       |

## OpenAI Integration

| Variable             | Description                                            | Default Value | Required |
| -------------------- | ------------------------------------------------------ | ------------- | -------- |
| `OPENAI_API_KEY`     | OpenAI API key for AI-powered features                 | -             | No       |
| `OPENAI_TEMPERATURE` | Temperature parameter for OpenAI completions (0.0-2.0) | `0.1`         | No       |
| `OPENAI_MODEL_NAME`  | OpenAI model name to use                               | `gpt-4o`      | No       |

## Custom Domains

| Variable                            | Description                                           | Default Value         | Required |
| ----------------------------------- | ----------------------------------------------------- | --------------------- | -------- |
| `CUSTOM_DOMAINS_RENEWAL_INTERVAL`   | Interval in seconds for checking certificate renewals | `3600` (1 hour)       | No       |
| `CUSTOM_DOMAINS_PROVISION_INTERVAL` | Interval in seconds for provisioning new domains      | `30`                  | No       |
| `CUSTOM_DOMAINS_CNAME_TARGET`       | CNAME target for custom domains                       | `custom.getprobo.com` | No       |

### ACME / Let's Encrypt

| Variable         | Description                                                     | Default Value                                    | Required |
| ---------------- | --------------------------------------------------------------- | ------------------------------------------------ | -------- |
| `ACME_DIRECTORY` | ACME directory URL for certificate issuance                     | `https://acme-v02.api.letsencrypt.org/directory` | No       |
| `ACME_EMAIL`     | Email address for ACME account registration                     | `admin@getprobo.com`                             | No       |
| `ACME_KEY_TYPE`  | Key type for ACME certificates (RSA2048, RSA4096, EC256, EC384) | `EC256`                                          | No       |
| `ACME_ROOT_CA`   | Custom root CA certificate (PEM format)                         | -                                                | No       |

## Connectors

### Slack Connector (OAuth2)

These variables are only used if `CONNECTOR_SLACK_CLIENT_ID` is set.

| Variable                         | Description                                       | Default Value                                               | Required                   |
| -------------------------------- | ------------------------------------------------- | ----------------------------------------------------------- | -------------------------- |
| `CONNECTOR_SLACK_CLIENT_ID`      | Slack OAuth2 app client ID                        | -                                                           | No                         |
| `CONNECTOR_SLACK_CLIENT_SECRET`  | Slack OAuth2 app client secret                    | -                                                           | **Yes** (if client ID set) |
| `CONNECTOR_SLACK_REDIRECT_URI`   | OAuth2 redirect URI for Slack connector           | `https://localhost:8080/api/console/v1/connectors/complete` | No                         |
| `CONNECTOR_SLACK_AUTH_URL`       | Slack OAuth2 authorization endpoint               | `https://slack.com/oauth/v2/authorize`                      | No                         |
| `CONNECTOR_SLACK_TOKEN_URL`      | Slack OAuth2 token endpoint                       | `https://slack.com/api/oauth.v2.access`                     | No                         |
| `CONNECTOR_SLACK_SIGNING_SECRET` | Slack app signing secret for webhook verification | -                                                           | **Yes** (if client ID set) |

## Security Best Practices

### Required Secrets

The following environment variables are **required** and must be set to secure random values in production:

1. `PROBOD_ENCRYPTION_KEY` - Generate with: `openssl rand -base64 32`
2. `AUTH_COOKIE_SECRET` - Generate with: `openssl rand -base64 32`
3. `AUTH_PASSWORD_PEPPER` - Generate with: `openssl rand -base64 32`
4. `TRUST_AUTH_TOKEN_SECRET` - Generate with: `openssl rand -base64 32`

### Secret Generation Example

```bash
# Generate all required secrets
export PROBOD_ENCRYPTION_KEY=$(openssl rand -base64 32)
export AUTH_COOKIE_SECRET=$(openssl rand -base64 32)
export AUTH_PASSWORD_PEPPER=$(openssl rand -base64 32)
export TRUST_AUTH_TOKEN_SECRET=$(openssl rand -base64 32)

echo "PROBOD_ENCRYPTION_KEY=$PROBOD_ENCRYPTION_KEY"
echo "AUTH_COOKIE_SECRET=$AUTH_COOKIE_SECRET"
echo "AUTH_PASSWORD_PEPPER=$AUTH_PASSWORD_PEPPER"
echo "TRUST_AUTH_TOKEN_SECRET=$TRUST_AUTH_TOKEN_SECRET"
```

## Configuration Priority

The entrypoint script follows this priority order:

1. If `CONFIG_FILE` exists (e.g., mounted from ConfigMap/volume), use it as-is
2. Otherwise, generate config file from environment variables
3. Environment variables use provided values or fall back to defaults
4. Script fails if required variables are missing (marked with `:?` in bash)
