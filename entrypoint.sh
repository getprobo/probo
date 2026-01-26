#!/bin/bash
set -e

# Configuration file path
CONFIG_FILE="${CONFIG_FILE:-/etc/probod/config.yml}"

# Function to generate default SAML certificate and private key if not provided
generate_saml_defaults() {
  if [ -z "$SAML_CERTIFICATE" ] || [ -z "$SAML_PRIVATE_KEY" ]; then
    echo "Generating default SAML certificate and private key..."

    # Generate private key and certificate valid for 10 years
    TEMP_KEY=$(mktemp)
    TEMP_CERT=$(mktemp)

    openssl req -x509 -newkey rsa:2048 -keyout "$TEMP_KEY" -out "$TEMP_CERT" \
      -days 3650 -nodes -subj "/CN=probo-saml/O=Probo/C=US" 2>/dev/null

    # Read generated files and export as environment variables
    export SAML_PRIVATE_KEY=$(cat "$TEMP_KEY")
    export SAML_CERTIFICATE=$(cat "$TEMP_CERT")

    # Clean up temporary files
    rm -f "$TEMP_KEY" "$TEMP_CERT"

    echo "Default SAML certificate and private key generated successfully"
  fi
}

# Function to validate CA bundle path
validate_pg_ca_bundle_path() {
  if [ -n "$PG_CA_BUNDLE_PATH" ]; then
    if [ -f "$PG_CA_BUNDLE_PATH" ]; then
      echo "Loading PostgreSQL CA bundle from: $PG_CA_BUNDLE_PATH"
      export PG_CA_BUNDLE_FILE="$PG_CA_BUNDLE_PATH"
    else
      echo "Warning: PG_CA_BUNDLE_PATH specified but file not found: $PG_CA_BUNDLE_PATH"
    fi
  fi
}

# Check if config file already exists (e.g., mounted from ConfigMap)
if [ -f "$CONFIG_FILE" ]; then
  echo "Using existing configuration file at: $CONFIG_FILE"
else
  echo "Generating configuration file from environment variables at: $CONFIG_FILE"

  # Generate default SAML credentials if not provided
  generate_saml_defaults

  # Validate PostgreSQL CA bundle path if configured
  validate_pg_ca_bundle_path

  # Create directory if it doesn't exist
  mkdir -p "$(dirname "$CONFIG_FILE")"

  cat > "$CONFIG_FILE" <<EOF
unit:
  metrics:
    addr: "${METRICS_ADDR:-localhost:8081}"
  tracing:
    addr: "${TRACING_ADDR:-localhost:4317}"
    max-batch-size: ${TRACING_MAX_BATCH_SIZE:-512}
    batch-timeout: ${TRACING_BATCH_TIMEOUT:-5}
    export-timeout: ${TRACING_EXPORT_TIMEOUT:-30}
    max-queue-size: ${TRACING_MAX_QUEUE_SIZE:-2048}

probod:
  base-url: "${PROBOD_BASE_URL:-http://localhost:8080}"
  encryption-key: "${PROBOD_ENCRYPTION_KEY:?PROBOD_ENCRYPTION_KEY is required}"
  chrome-dp-addr: "${CHROME_DP_ADDR:-localhost:9222}"

  api:
    addr: "${API_ADDR:-:8080}"
    cors:
      allowed-origins: [${API_CORS_ALLOWED_ORIGINS:-"http://localhost:8080"}]
    extra-header-fields: {}

  pg:
    addr: "${PG_ADDR:-localhost:5432}"
    username: "${PG_USERNAME:-postgres}"
    password: "${PG_PASSWORD:-postgres}"
    database: "${PG_DATABASE:-probod}"
    pool-size: ${PG_POOL_SIZE:-100}
EOF

  # Add PostgreSQL CA bundle if configured
  if [ -n "$PG_CA_BUNDLE_FILE" ]; then
    cat >> "$CONFIG_FILE" <<EOF
    ca-cert-bundle: |
$(sed 's/^/      /' "$PG_CA_BUNDLE_FILE")
EOF
  elif [ -n "$PG_CA_BUNDLE" ]; then
    cat >> "$CONFIG_FILE" <<EOF
    ca-cert-bundle: |
$(echo "$PG_CA_BUNDLE" | sed 's/^/      /')
EOF
  fi

  cat >> "$CONFIG_FILE" <<EOF

  auth:
    disable-signup: ${AUTH_DISABLE_SIGNUP:-false}
    invitation-confirmation-token-validity: ${AUTH_INVITATION_TOKEN_VALIDITY:-3600}
    cookie:
      name: "${AUTH_COOKIE_NAME:-SSID}"
      domain: "${AUTH_COOKIE_DOMAIN:-localhost}"
      secret: "${AUTH_COOKIE_SECRET:?AUTH_COOKIE_SECRET is required}"
      duration: ${AUTH_COOKIE_DURATION:-24}
      secure: ${AUTH_COOKIE_SECURE:-true}
    password:
      pepper: "${AUTH_PASSWORD_PEPPER:?AUTH_PASSWORD_PEPPER is required}"
      iterations: ${AUTH_PASSWORD_ITERATIONS:-1000000}
    saml:
      session-duration: ${SAML_SESSION_DURATION:-604800}
      cleanup-interval-seconds: ${SAML_CLEANUP_INTERVAL_SECONDS:-0}
      certificate: |
$(echo "${SAML_CERTIFICATE:-}" | sed 's/^/        /')
      private-key: |
$(echo "${SAML_PRIVATE_KEY:-}" | sed 's/^/        /')

  trust-auth:
    cookie-name: "${TRUST_AUTH_COOKIE_NAME:-TCT}"
    cookie-domain: "${TRUST_AUTH_COOKIE_DOMAIN:-localhost}"
    cookie-duration: ${TRUST_AUTH_COOKIE_DURATION:-24}
    token-duration: ${TRUST_AUTH_TOKEN_DURATION:-168}
    report-url-duration: ${TRUST_AUTH_REPORT_URL_DURATION:-15}
    token-secret: "${TRUST_AUTH_TOKEN_SECRET:?TRUST_AUTH_TOKEN_SECRET is required}"
    scope: "${TRUST_AUTH_SCOPE:-trust_center_readonly}"
    token-type: "${TRUST_AUTH_TOKEN_TYPE:-trust_center_access}"

  aws:
    region: "${AWS_REGION:-us-east-1}"
    bucket: "${AWS_BUCKET:-probod}"
    access-key-id: "${AWS_ACCESS_KEY_ID:-}"
    secret-access-key: "${AWS_SECRET_ACCESS_KEY:-}"
    endpoint: "${AWS_ENDPOINT:-}"
    use-path-style: ${AWS_USE_PATH_STYLE:-false}

  notifications:
    mailer:
      sender-name: "${MAILER_SENDER_NAME:-Probo}"
      sender-email: "${MAILER_SENDER_EMAIL:-no-reply@notification.getprobo.com}"
      smtp:
        addr: "${SMTP_ADDR:-localhost:1025}"
        user: "${SMTP_USER:-}"
        password: "${SMTP_PASSWORD:-}"
        tls-required: ${SMTP_TLS_REQUIRED:-false}
      mailer-interval: ${MAILER_INTERVAL:-60}
    slack:
      sender-interval: ${SLACK_SENDER_INTERVAL:-60}

  openai:
    api-key: "${OPENAI_API_KEY:-}"
    temperature: ${OPENAI_TEMPERATURE:-0.1}
    model-name: "${OPENAI_MODEL_NAME:-gpt-4o}"

  custom-domains:
    renewal-interval: ${CUSTOM_DOMAINS_RENEWAL_INTERVAL:-3600}
    provision-interval: ${CUSTOM_DOMAINS_PROVISION_INTERVAL:-30}
    cname-target: "${CUSTOM_DOMAINS_CNAME_TARGET:-custom.getprobo.com}"
    acme:
      directory: "${ACME_DIRECTORY:-https://acme-v02.api.letsencrypt.org/directory}"
      email: "${ACME_EMAIL:-admin@getprobo.com}"
      key-type: "${ACME_KEY_TYPE:-EC256}"
      root-ca: "${ACME_ROOT_CA:-}"
EOF

  # Add connectors if configured
  if [ -n "$CONNECTOR_SLACK_CLIENT_ID" ]; then
    cat >> "$CONFIG_FILE" <<EOF

  connectors:
    - provider: "slack"
      protocol: "oauth2"
      config:
        client-id: "${CONNECTOR_SLACK_CLIENT_ID}"
        client-secret: "${CONNECTOR_SLACK_CLIENT_SECRET:?CONNECTOR_SLACK_CLIENT_SECRET is required when CONNECTOR_SLACK_CLIENT_ID is set}"
        redirect-uri: "${CONNECTOR_SLACK_REDIRECT_URI:-https://localhost:8080/api/console/v1/connectors/complete}"
        auth-url: "${CONNECTOR_SLACK_AUTH_URL:-https://slack.com/oauth/v2/authorize}"
        token-url: "${CONNECTOR_SLACK_TOKEN_URL:-https://slack.com/api/oauth.v2.access}"
        scopes:
          - "chat:write"
          - "channels:join"
          - "incoming-webhook"
      settings:
        signing-secret: "${CONNECTOR_SLACK_SIGNING_SECRET:?CONNECTOR_SLACK_SIGNING_SECRET is required when CONNECTOR_SLACK_CLIENT_ID is set}"
EOF
  fi

  echo "Configuration file generated at: $CONFIG_FILE"
fi

# Execute probod with the generated config
exec probod -cfg-file "$CONFIG_FILE" "$@"
