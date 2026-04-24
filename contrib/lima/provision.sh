#!/bin/bash
# Copyright (c) 2025, 2026 Probo Inc.
# SPDX-License-Identifier: ISC

set -euo pipefail

# Read Lima user from cidata (values are unquoted so we cannot source the file
# directly — LIMA_CIDATA_COMMENT may contain spaces).
LIMA_CIDATA_USER="$(sed -n 's/^LIMA_CIDATA_USER=//p' /mnt/lima-cidata/lima.env)"
export LIMA_CIDATA_USER

export DEBIAN_FRONTEND=noninteractive

GO_VERSION="1.26.1"
NODE_MAJOR=24
NPM_VERSION="11.8.0"

GOTESTSUM_VERSION="v1.13.0"
GOLANGCI_LINT_VERSION="v2.11.3"
GOW_VERSION="v0.0.0-20260225145757-ff0f6779ab4c"
MKCERT_VERSION="v1.4.4"

apt-get update -qq
apt-get install -y -qq \
    build-essential \
    git \
    curl \
    jq \
    parallel \
    ca-certificates \
    gnupg \
    lsb-release \
    unzip

if ! command -v docker &>/dev/null; then
    install -m 0755 -d /etc/apt/keyrings
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg \
        | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
    chmod a+r /etc/apt/keyrings/docker.gpg

    echo \
        "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] \
        https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" \
        | tee /etc/apt/sources.list.d/docker.list > /dev/null

    apt-get update -qq
    apt-get install -y -qq \
        docker-ce \
        docker-ce-cli \
        containerd.io \
        docker-buildx-plugin \
        docker-compose-plugin

    systemctl enable --now docker
fi

usermod -aG docker "${LIMA_CIDATA_USER:-lima}" 2>/dev/null || true

if [ ! -d "/usr/local/go" ] || ! /usr/local/go/bin/go version | grep -q "go${GO_VERSION}"; then
    rm -rf /usr/local/go
    ARCH=$(dpkg --print-architecture)
    curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${ARCH}.tar.gz" \
        | tar -C /usr/local -xzf -
fi

cat > /etc/profile.d/go.sh << 'GOEOF'
export PATH="/usr/local/go/bin:$HOME/go/bin:$PATH"
GOEOF
chmod +x /etc/profile.d/go.sh

export PATH="/usr/local/go/bin:$PATH"
export HOME="${HOME:-/root}"

GOBIN=/usr/local/bin /usr/local/go/bin/go install "gotest.tools/gotestsum@${GOTESTSUM_VERSION}"
GOBIN=/usr/local/bin /usr/local/go/bin/go install "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}"
GOBIN=/usr/local/bin /usr/local/go/bin/go install "github.com/mitranim/gow@${GOW_VERSION}"

if ! command -v node &>/dev/null || ! node --version | grep -q "v${NODE_MAJOR}"; then
    curl -fsSL "https://deb.nodesource.com/setup_${NODE_MAJOR}.x" | bash -
    apt-get install -y -qq nodejs
fi

npm install -g "npm@${NPM_VERSION}"

if ! command -v mkcert &>/dev/null; then
    GOBIN=/usr/local/bin /usr/local/go/bin/go install "filippo.io/mkcert@${MKCERT_VERSION}"
fi
mkcert -install 2>/dev/null || true

LIMA_USER="${LIMA_CIDATA_USER:-lima}"
LIMA_HOME=$(eval echo "~${LIMA_USER}")
mkdir -p /root/.parallel "${LIMA_HOME}/.parallel"
touch /root/.parallel/will-cite "${LIMA_HOME}/.parallel/will-cite"
chown -R "${LIMA_USER}:${LIMA_USER}" "${LIMA_HOME}/.parallel"

# Generate sandbox-specific probod config and frontend .env files
VM_IP=$(ip -4 -j addr show dev lima0 | jq -r '.[0].addr_info[0].local')

su - "${LIMA_USER}" -c "export PATH=/usr/local/go/bin:\$HOME/go/bin:\$PATH && cd /workspace && make bin/probod-bootstrap"

mkdir -p /etc/probod

OAUTH2_SIGNING_KEY_PATH=/etc/probod/oauth2-signing-key.pem
if [ ! -f "${OAUTH2_SIGNING_KEY_PATH}" ]; then
    openssl genrsa -out "${OAUTH2_SIGNING_KEY_PATH}" 2048
    chmod 600 "${OAUTH2_SIGNING_KEY_PATH}"
fi

# Load developer-specific overrides (not committed to repo).
if [ -f /workspace/.sandbox.env ]; then
    set -a
    . /workspace/.sandbox.env
    set +a
fi

PROBOD_BASE_URL="http://${VM_IP}:8080" \
AUTH_COOKIE_DOMAIN="${VM_IP}" \
AUTH_COOKIE_SECURE=false \
AUTH_COOKIE_SECRET="this-is-a-secure-secret-for-cookie-signing-at-least-32-bytes" \
AUTH_PASSWORD_PEPPER="this-is-a-secure-pepper-for-password-hashing-at-least-32-bytes" \
PROBOD_ENCRYPTION_KEY="thisisnotasecretAAAAAAAAAAAAAAAAAAAAAAAAAAA=" \
OAUTH2_SERVER_SIGNING_KEY="$(cat "${OAUTH2_SIGNING_KEY_PATH}")" \
API_CORS_ALLOWED_ORIGINS="http://${VM_IP}:8080,http://${VM_IP}:5173,http://${VM_IP}:5174" \
AWS_ENDPOINT="http://127.0.0.1:8333" \
AWS_ACCESS_KEY_ID="probod" \
AWS_SECRET_ACCESS_KEY="thisisnotasecret" \
AWS_USE_PATH_STYLE=true \
    /workspace/bin/probod-bootstrap -output /etc/probod/config.yml

# probod runs as ${LIMA_USER} but bootstrap writes config.yml as root with 0600
# because it contains secrets. Transfer ownership so probod can read it.
chown "${LIMA_USER}:${LIMA_USER}" /etc/probod/config.yml "${OAUTH2_SIGNING_KEY_PATH}"

echo "VITE_API_URL=http://${VM_IP}:8080" > /workspace/apps/console/.env
echo "VITE_API_URL=http://${VM_IP}:8080" > /workspace/apps/trust/.env

# Bind-mount VM-local node_modules over the shared workspace to avoid
# platform conflicts between macOS host and Linux VM native binaries.
cat > /etc/systemd/system/probo-node-modules.service << EOF
[Unit]
Description=Bind-mount VM-local node_modules over workspace
DefaultDependencies=no
Before=probo-console.service probo-trust.service

[Service]
Type=oneshot
RemainAfterExit=yes
ExecStartPre=/bin/mkdir -p /var/lib/probo/node_modules /workspace/node_modules
ExecStart=/bin/mount --bind /var/lib/probo/node_modules /workspace/node_modules
ExecStartPost=/bin/chown ${LIMA_USER}:${LIMA_USER} /var/lib/probo/node_modules

[Install]
WantedBy=multi-user.target
EOF

# Install systemd services for the sandbox
cat > /etc/systemd/system/probo-stack.service << 'EOF'
[Unit]
Description=Probo Docker Compose Stack
Requires=docker.service
After=docker.service

[Service]
Type=simple
User=root
WorkingDirectory=/workspace
ExecStart=/usr/bin/docker compose -f compose.yaml up
ExecStop=/usr/bin/docker compose -f compose.yaml down
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF

cat > /etc/systemd/system/probod.service << EOF
[Unit]
Description=Probo API Server
Requires=probo-stack.service
After=probo-stack.service

[Service]
Type=simple
User=${LIMA_USER}
WorkingDirectory=/workspace
ExecStart=/usr/local/bin/gow -r=false run ./cmd/probod -cfg-file /etc/probod/config.yml
Restart=on-failure
RestartSec=3s
Environment=PATH=/usr/local/go/bin:/usr/local/bin:/usr/bin:/bin

[Install]
WantedBy=multi-user.target
EOF

cat > /etc/systemd/system/probo-console.service << EOF
[Unit]
Description=Probo Console Dev Server
Requires=probo-node-modules.service
After=probod.service probo-node-modules.service

[Service]
Type=simple
User=${LIMA_USER}
WorkingDirectory=/workspace
ExecStart=/usr/bin/npm --workspace @probo/console run dev -- --host 0.0.0.0
Restart=on-failure
RestartSec=3s

[Install]
WantedBy=multi-user.target
EOF

cat > /etc/systemd/system/probo-trust.service << EOF
[Unit]
Description=Probo Trust Dev Server
Requires=probo-node-modules.service
After=probod.service probo-node-modules.service

[Service]
Type=simple
User=${LIMA_USER}
WorkingDirectory=/workspace
ExecStart=/usr/bin/npm --workspace @probo/trust run dev -- --host 0.0.0.0
Restart=on-failure
RestartSec=3s

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable --now probo-node-modules.service
systemctl enable --now probo-stack.service
systemctl enable probod.service probo-console.service probo-trust.service

# Populate VM-local node_modules with Linux-native binaries (esbuild, etc.).
# The host's node_modules is macOS; `probo-node-modules.service` bind-mounts an
# empty tree over /workspace/node_modules, and we install into it once here so
# `make build` and the dev servers can run without cross-platform mismatches.
if [ -z "$(ls -A /var/lib/probo/node_modules 2>/dev/null)" ]; then
    su - "${LIMA_USER}" -c "cd /workspace && npm ci"
fi
