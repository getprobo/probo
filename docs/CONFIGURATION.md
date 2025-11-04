# Probod Configuration Reference

This document provides a comprehensive reference for configuring the Probo compliance management platform daemon (`probod`).

For installation instructions, please refer to the [Installation Guide](./INSTALLATION.md).

## Configuration File Format

Probod uses YAML format for configuration files. The configuration is structured hierarchically with the root key `probod` containing all service-specific settings.

### Basic Configuration Structure

```yaml
unit:
  metrics:
    addr: "localhost:8081"

  tracing:
    addr: "localhost:4317"
    max-batch-size: 1000
    batch-timeout: 10
    export-timeout: 10
    max-queue-size: 10000

probod:
  base-url: "http://localhost:8080"
  encryption-key: "base64-encoded-encryption-key"
  chrome-dp-addr: "localhost:9222"

  api:
    addr: "localhost:8080"
    cors:
      allowed-origins: ["http://localhost:3000"]
    extra-header-fields:
      "Custom-Header": "value"

  pg:
    addr: "localhost:5432"
    username: "postgres"
    password: "postgres"
    database: "probod"
    pool-size: 100
    ca-cert-bundle: |
      -----BEGIN CERTIFICATE-----
      ...certificate content...
      -----END CERTIFICATE-----

  auth:
    disable-signup: false
    invitation-confirmation-token-validity: 3600
    cookie:
      name: "SSID"
      domain: "localhost"
      secret: "cookie-signing-secret"
      duration: 24
    password:
      pepper: "password-hashing-pepper"
      iterations: 1000000

  trust-auth:
    cookie-name: "TCT"
    cookie-domain: "localhost"
    cookie-duration: 24
    token-duration: 720
    report-url-duration: 15
    token-secret: "trust-token-signing-secret"
    scope: "trust_center_readonly"
    token-type: "trust_center_access"

  trust-center:
    http-addr: ":80"
    https-addr: ":443"

  aws:
    region: "us-east-1"
    bucket: "probod"
    access-key-id: "access-key"
    secret-access-key: "secret-key"
    endpoint: "http://127.0.0.1:9000"

  notifications:
    mailer:
      sender-name: "Probo"
      sender-email: "no-reply@notification.getprobo.com"
      mailer-interval: 60
      smtp:
        addr: "localhost:1025"
        user: "smtp-username"
        password: "smtp-password"
        tls-required: false
    slack:
      sender-interval: 60

  openai:
    api-key: "openai-api-key"
    temperature: 0.1
    model-name: "gpt-4o"

  custom-domains:
    renewal-interval: 3600
    provision-interval: 30
    cname-target: "custom.getprobo.com"
    acme:
      directory: "https://localhost:14000/dir"
      email: "admin@getprobo.com"
      key-type: "EC256"
      root-ca: ""
      account-key: ""
      root-ca: ""

  connectors:
    - provider: "slack"
      protocol: "oauth2"
      config:
        client-id: "slack-client-id"
        client-secret: "slack-client-secret"
        redirect-uri: "https://localhost:8080/api/console/v1/connectors/complete"
        auth-url: "https://slack.com/oauth/v2/authorize"
        token-url: "https://slack.com/api/oauth.v2.access"
        scopes:
          - "chat:write"
          - "channels:join"
          - "incoming-webhook"
      settings:
        signing-secret: "slack-signing-secret"
```

## Telemetry and Observability

Probod includes built-in support for Prometheus metrics and OpenTelemetry tracing. The telemetry configuration is part of the main configuration file structure.

### Telemetry Configuration

The telemetry configuration is specified at the top level of the configuration file under the `unit` section:

```yaml
unit:
  metrics:
    addr: "localhost:8081"

  tracing:
    addr: "localhost:8082"
    max-batch-size: 1000
    batch-timeout: 10
    export-timeout: 10
    max-queue-size: 10000

probod:
  # ... rest of probod configuration
```

### Prometheus Metrics

#### `unit.metrics.addr` (string)

**Default**: Same as API address

Network address where the Prometheus metrics endpoint will be exposed. The metrics are available at `/metrics` on this address.

#### Example Configuration

```yaml
unit:
  metrics:
    addr: "0.0.0.0:8081" # Expose metrics on port 8081
```

### OpenTelemetry Tracing

#### `unit.tracing.addr` (string)

**Default**: Not configured (tracing disabled)

Network address for the OpenTelemetry trace exporter endpoint.

#### `unit.tracing.max-batch-size` (integer)

**Default**: `512`

Maximum number of spans to batch before exporting.

#### `unit.tracing.batch-timeout` (integer)

**Default**: `5` (seconds)

Maximum time to wait before exporting a batch of spans.

#### `unit.tracing.export-timeout` (integer)

**Default**: `30` (seconds)

Maximum time to wait for span export to complete.

#### `unit.tracing.max-queue-size` (integer)

**Default**: `2048`

Maximum queue size for spans waiting to be exported.

#### Example Configuration

```yaml
unit:
  tracing:
    addr: "tempo:4317" # OTLP gRPC endpoint
    max-batch-size: 1000
    batch-timeout: 10
    export-timeout: 10
    max-queue-size: 10000
```

#### Built-in Instrumentation

Probod automatically instruments:

- Database operations (PostgreSQL queries)
- HTTP requests and responses
- GraphQL operations and resolvers
- S3 storage operations
- PDF generation processes

#### Prometheus Scrape Configuration

Configure Prometheus to scrape metrics from probod:

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: "probod"
    scrape_interval: 15s
    static_configs:
      - targets: ["probod:8081"] # Use metrics addr from config
    metrics_path: "/metrics"
```

### Logging

Probod provides automatic structured JSON logging with:

- Request correlation IDs
- Integration with OpenTelemetry trace and span IDs
- Component-specific loggers
- Consistent formatting across all services

## Configuration Sections

### General Settings

#### `base-url` (string)

**Default**: `"http://localhost:8080"`

The base URL where the Probod service will be accessible externally. This should include the scheme (http or https), hostname, and optionally port. This setting affects URL generation for emails, redirects, and API responses. For production deployments, use the full HTTPS URL (e.g., `"https://app.example.com"`).

#### `encryption-key` (string)

**Required**

Base64-encoded encryption key used for encrypting sensitive data at rest. Must be provided for production deployments.

#### `chrome-dp-addr` (string)

**Default**: `"localhost:9222"`

Address of the Chrome DevTools Protocol endpoint used for PDF generation and document processing.

### API Configuration

#### `api.addr` (string)

**Default**: `"localhost:8080"`

The network address and port where the Probod API server will listen for incoming connections.

#### `api.cors.allowed-origins` (array of strings)

**Default**: `[]`

List of origins allowed for Cross-Origin Resource Sharing (CORS) requests. Required for web applications accessing the API from different domains.

#### `api.extra-header-fields` (map of string to string)

**Default**: `{}`

Additional HTTP headers to include in API responses. Useful for custom security headers or integration requirements.

### Database Configuration

#### `pg.addr` (string)

**Default**: `"localhost:5432"`

PostgreSQL database server address and port.

#### `pg.username` (string)

**Default**: `"postgres"`

Database username for authentication.

#### `pg.password` (string)

**Default**: `"postgres"`

Database password for authentication.

#### `pg.database` (string)

**Default**: `"probod"`

Name of the PostgreSQL database to connect to.

#### `pg.pool-size` (integer)

**Default**: `100`

Maximum number of database connections in the connection pool.

#### `pg.ca-cert-bundle` (string)

**Optional**

PEM-encoded CA certificate bundle for TLS database connections. Required when connecting to databases with custom or self-signed certificates.

### Authentication Configuration

#### `auth.disable-signup` (boolean)

**Default**: `false`

When set to `true`, disables new user registration through the web interface.

#### `auth.invitation-confirmation-token-validity` (integer)

**Default**: `3600`

Validity period (in seconds) for email invitation confirmation tokens.

#### `auth.cookie.name` (string)

**Default**: `"SSID"`

Name of the HTTP cookie used for session management.

#### `auth.cookie.domain` (string)

**Default**: `"localhost"`

Domain scope for session cookies.

#### `auth.cookie.secret` (string)

**Default**: Auto-generated

Secret key used for signing session cookies. Should be at least 32 bytes for security.

#### `auth.cookie.duration` (integer)

**Default**: `24`

Session cookie lifetime in hours.

#### `auth.cookie.secure` (boolean)

**Default**: `true`

Controls whether the Secure flag is set on session cookies. When true, cookies are only sent over HTTPS connections.

**Important**: This must be set to `true` for SAML authentication to work properly. SAML requires `SameSite=None` cookies for cross-site POST requests from identity providers, and modern browsers require the `Secure` flag to be set when using `SameSite=None`. Setting this to `false` will cause SAML authentication to fail as session cookies will be rejected by browsers.

#### `auth.password.pepper` (string)

**Default**: Auto-generated

Additional secret value used in password hashing. Should be at least 32 bytes and kept confidential.

#### `auth.password.iterations` (integer)

**Default**: `1000000`

Number of iterations for password hashing algorithm (PBKDF2). Higher values increase security but require more computational resources.

### Trust Center Authentication

#### `trust-auth.cookie-name` (string)

**Default**: `"TCT"`

Name of the HTTP cookie used for trust center access tokens.

#### `trust-auth.cookie-domain` (string)

**Default**: `"localhost"`

Domain scope for trust center cookies.

#### `trust-auth.cookie-duration` (integer)

**Default**: `24`

Trust center cookie lifetime in hours.

#### `trust-auth.token-duration` (integer)

**Default**: `720`

Trust center access token lifetime in hours.

#### `trust-auth.report-url-duration` (integer)

**Default**: `15`

Validity period for generated report URLs in minutes.

#### `trust-auth.token-secret` (string)

**Default**: Auto-generated

Secret key used for signing trust center tokens. Should be at least 32 bytes.

#### `trust-auth.scope` (string)

**Default**: `"trust_center_readonly"`

OAuth2 scope for trust center access.

#### `trust-auth.token-type` (string)

**Default**: `"trust_center_access"`

Type identifier for trust center access tokens.

### Trust Center Configuration

#### `trust-center.http-addr` (string)

**Default**: `":80"`

Network address and port where the trust center HTTP server will listen for ACME HTTP-01 challenges and HTTP to HTTPS redirects.

#### `trust-center.https-addr` (string)

**Default**: `":443"`

Network address and port where the trust center HTTPS server will listen for secure connections. This server handles custom domain routing with automatic TLS certificate management.

### AWS Configuration

#### `aws.region` (string)

**Default**: `"us-east-1"`

AWS region for S3 bucket operations.

#### `aws.bucket` (string)

**Default**: `"probod"`

S3 bucket name for file storage.

#### `aws.access-key-id` (string)

**Required**

AWS access key ID for authentication.

#### `aws.secret-access-key` (string)

**Required**

AWS secret access key for authentication.

#### `aws.endpoint` (string)

**Optional**

Custom S3-compatible endpoint URL. Useful for local development with MinIO or other S3-compatible services.

### Notifications Configuration

The `notifications` section configures how Probod sends notifications through various channels.

#### `notifications.mailer.sender-name` (string)

**Default**: `"Probo"`

Display name for outgoing emails.

#### `notifications.mailer.sender-email` (string)

**Default**: `"no-reply@notification.getprobo.com"`

Email address used as the sender for outgoing emails.

#### `notifications.mailer.mailer-interval` (integer)

**Default**: `60`

Interval in seconds between checking for pending email notifications to send.

#### `notifications.mailer.smtp.addr` (string)

**Default**: `"localhost:1025"`

SMTP server address and port.

#### `notifications.mailer.smtp.user` (string)

**Optional**

Username for SMTP authentication.

#### `notifications.mailer.smtp.password` (string)

**Optional**

Password for SMTP authentication.

#### `notifications.mailer.smtp.tls-required` (boolean)

**Default**: `false`

Whether TLS encryption is required for SMTP connections.

#### `notifications.slack.sender-interval` (integer)

**Default**: `60`

Interval in seconds between checking for pending Slack notifications to send.

### OpenAI Integration

#### `openai.api-key` (string)

**Required for AI features**

API key for OpenAI services integration.

#### `openai.temperature` (float)

**Default**: `0.1`

Temperature parameter for AI model responses (0.0 to 1.0). Lower values produce more deterministic outputs.

#### `openai.model-name` (string)

**Default**: `"gpt-4o"`

OpenAI model identifier to use for AI-powered features.

### Custom Domains Configuration

The `custom-domains` section configures automatic TLS certificate management for custom trust center domains using ACME (Let's Encrypt).

#### `custom-domains.renewal-interval` (integer)

**Default**: `3600`

Interval in seconds between checking for certificates that need renewal.

#### `custom-domains.provision-interval` (integer)

**Default**: `30`

Interval in seconds between checking for domains that need certificate provisioning.

#### `custom-domains.cname-target` (string)

**Default**: `"custom.getprobo.com"`

The CNAME target that custom domains should point to. This is used for domain validation and documentation.

#### `custom-domains.acme.directory` (string)

**Default**: `"https://localhost:14000/dir"`

ACME directory URL. For local development, use Pebble at `"https://localhost:14000/dir"`.

#### `custom-domains.acme.email` (string)

**Default**: `"admin@getprobo.com"`

Contact email address for ACME account registration and certificate expiration notifications.

#### `custom-domains.acme.key-type` (string)

**Default**: `"EC256"`

Type of cryptographic key to use for certificates. Supported values: `"EC256"`, `"EC384"`, `"RSA2048"`, `"RSA4096"`.

#### `custom-domains.acme.root-ca` (string)

**Optional**

PEM-encoded root CA certificate for ACME server validation. Required when using private ACME servers or testing with Pebble.

#### `custom-domains.acme.account-key` (string)

**Optional**

PEM-encoded ACME account private key. If not provided, a new account key will be generated automatically.

### External Connectors

The `connectors` section defines external service integrations.

#### OAuth2 Connector Configuration

```yaml
connectors:
  - provider: "slack"
    protocol: "oauth2"
    config:
      client-id: "oauth2-client-id"
      client-secret: "oauth2-client-secret"
      redirect-uri: "https://your-domain/api/console/v1/connectors/complete"
      auth-url: "https://slack.com/oauth/v2/authorize"
      token-url: "https://slack.com/api/oauth.v2.access"
      scopes:
        - "chat:write"
        - "channels:join"
        - "incoming-webhook"
    settings:
      signing-secret: "slack-signing-secret"
```

##### `connectors[].provider` (string)

**Required**

Service provider name. Supported values: `"slack"`.

##### `connectors[].protocol` (string)

**Required**

Connector protocol type. Currently supported: `"oauth2"`.

##### `connectors[].config.client-id` (string)

**Required**

OAuth2 client identifier provided by the external service.

##### `connectors[].config.client-secret` (string)

**Required**

OAuth2 client secret provided by the external service.

##### `connectors[].config.redirect-uri` (string)

**Required**

OAuth2 redirect URI. Must match the URI registered with the external service.

##### `connectors[].config.auth-url` (string)

**Required**

OAuth2 authorization endpoint URL.

##### `connectors[].config.token-url` (string)

**Required**

OAuth2 token exchange endpoint URL.

##### `connectors[].config.scopes` (array of strings)

**Optional**

List of OAuth2 scopes to request during authorization.

##### `connectors[].settings` (object)

**Optional**

Additional provider-specific settings. For Slack connectors, this includes:

- `signing-secret` (string): Slack signing secret for webhook verification.

## Troubleshooting

### Common Configuration Issues

1. **Database Connection Failures**: Verify database credentials, network connectivity, and certificate configuration.
2. **Authentication Problems**: Check cookie domain settings and secret key configuration.
3. **External Connector Issues**: Verify OAuth2 client credentials and redirect URI configuration.
4. **File Upload Problems**: Ensure AWS credentials and S3 bucket configuration are correct.

### Logging

Probod provides structured logging that can help diagnose configuration issues. Enable debug logging by setting appropriate log levels in your deployment environment.

## Configuration Validation

Probod validates configuration on startup and will report specific errors for:

- Missing required fields
- Invalid data formats
- Unreachable external services
- Invalid secrets or keys

Review startup logs carefully to identify and resolve configuration issues.
