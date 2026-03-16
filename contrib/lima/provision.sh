#!/bin/bash
# Copyright (c) 2025, 2026 Probo Inc.
# SPDX-License-Identifier: ISC

set -euo pipefail

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
