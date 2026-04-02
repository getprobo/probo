# Govrly - Cybersecurity Governance & Compliance Platform

Govrly is a cybersecurity Governance, Risk, and Compliance (GRC) platform designed to help organizations manage regulatory requirements, monitor security posture, and maintain continuous compliance.

The platform centralizes governance processes, risk tracking, compliance frameworks, and audit evidence in a single system.

---

# Platform Capabilities

## Governance
- Policy and procedure management
- Control libraries and security baselines
- Compliance framework mapping
- Governance workflow automation
- Audit readiness management

## Risk Management
- Risk identification and classification
- Risk register management
- Risk treatment planning
- Residual risk tracking
- Continuous risk monitoring

## Compliance Management
- Compliance framework implementation tracking
- Evidence collection workflows
- Compliance dashboards and reporting
- Control maturity tracking
- Compliance gap identification

## Audit Management
- Internal audit preparation
- Evidence repository
- Control testing documentation
- Audit findings tracking
- Audit reporting

---

# Supported Compliance Frameworks

- ISO 27001
- NCA ECC
- SAMA Cybersecurity Framework
- PDPL
- SOC 2
- Custom frameworks

---

# Architecture Overview

## Backend
- Go
- GraphQL
- PostgreSQL

## Frontend
- React
- TypeScript
- Relay
- TailwindCSS

## Infrastructure
- Docker
- OpenTelemetry
- CI/CD

## Observability
- Prometheus
- Grafana
- Loki
- Tempo

---

# 🚀 Getting Started

## Prerequisites

- Go 1.21+
- Node.js 22+
- Docker
- mkcert

---

# ⚡ Quick Start

## Clone the repository

git clone --recurse-submodules https://github.com/govrly/govrly.git
cd govrly

## Install dependencies

# Install Go dependencies
go mod download

# Install Node.js dependencies
npm ci

## Start the development environment

# Start infrastructure services
make stack-up

# Build the project
make build

# Start the application using development settings
bin/govrlyd -cfg-file cfg/dev.yaml

The application will be available at:

http://localhost:8080

---

# Testing Custom Domains

Add to hosts file:

127.0.0.1 custom.govrly.local

---

# Security

- RBAC
- Secure authentication
- Audit logging
- Encryption support

---

# License

Proprietary software developed for Govrly.
All rights reserved.
