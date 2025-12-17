# Probo Kubernetes Deployment

This directory contains the Helm chart for deploying Probo on Kubernetes with external managed services.

## Quick Links

- [Helm Chart Documentation](charts/probo/README.md)
- [Values Reference](charts/probo/values.yaml)

## Prerequisites

Before deploying Probo, ensure you have:

1. **Kubernetes Cluster** - Version 1.23+
2. **Helm** - Version 3.8+
3. **PostgreSQL Database** - Managed service (AWS RDS, GCP Cloud SQL, Azure Database, etc.)
4. **S3 Storage** - AWS S3 or S3-compatible storage (GCS, DigitalOcean Spaces, MinIO, etc.)

## Install

#### Using Local Chart

##### Generate Secrets

```bash
export ENCRYPTION_KEY=$(openssl rand -base64 32)
export COOKIE_SECRET=$(openssl rand -base64 32)
export PASSWORD_PEPPER=$(openssl rand -base64 32)
export TRUST_TOKEN_SECRET=$(openssl rand -base64 32)
```

#### Download remote dependencies

```bash
helm dependency update ./charts/probo
```

##### Install using Chart and set values

```bash
helm install my-probo ./charts/probo \
  --set probo.baseUrl="probo.example.com" \
  --set probo.encryptionKey="$ENCRYPTION_KEY" \
  --set probo.auth.cookieSecret="$COOKIE_SECRET" \
  --set probo.auth.passwordPepper="$PASSWORD_PEPPER" \
  --set probo.trustAuth.tokenSecret="$TRUST_TOKEN_SECRET" \
  --set postgresql.enabled=true \
  --set postgresql.auth.postgresUser="probod" \
  --set postgresql.auth.postgresPassword="your-db-password" \
  --set postgresql.auth.database="probod" \
  --set minio.enabled=true \
  --set s3.bucket="your-bucket-name" \
  --set s3.accessKeyId="your-access-key" \
  --set s3.secretAccessKey="your-secret-key"
```

##### Install using Chart and values file


```bash
helm install my-probo ./charts/probo \
  --set probo.encryptionKey="$ENCRYPTION_KEY" \
  --set probo.auth.cookieSecret="$COOKIE_SECRET" \
  --set probo.auth.passwordPepper="$PASSWORD_PEPPER" \
  --set probo.trustAuth.tokenSecret="$TRUST_TOKEN_SECRET" \
  --set probo.mailer.smtp.password="smtp-password" \
  --set postgresql.enabled=true \
  --set postgresql.auth.postgresPassword="probod" \
  --set s3.accessKeyId="your-access-key" \
  --set s3.secretAccessKey="your-secret-key" \
  -f ./charts/probo/values.yaml
```

### 3. Access

```bash
kubectl port-forward svc/probo 8080:8080
# Visit http://localhost:8080
```

## Production Deployment

For production deployments, we recommend:

1. **Copy the production template:**
   ```bash
   cp charts/probo/values.yaml charts/probo/values-k8s-production.yaml
   ```

2. **Edit the configuration:**
   - Set your domain name
   - Configure external PostgreSQL connection
   - Configure S3 storage credentials
   - Add SMTP settings for email
   - Enable ingress with TLS
   - Configure autoscaling

3. **Install:**
   ```bash
   helm install probo ./charts/probo -f ./charts/probo/values-k8s-production.yaml
   ```

## Architecture

### What Gets Deployed

- **Probo Application** - Main Go binary serving GraphQL APIs and React frontends
- **Chrome Headless** - For PDF generation (optional, can use external service)
- **LoadBalancer Service** - For external access via TCP on ports 80, 443, and 8080
- **Ingress** - Alternative to LoadBalancer for HTTP routing (optional)

### External Dependencies (Required)

- **PostgreSQL** - Managed database for compliance data
- **S3 Storage** - Object storage for files and documents

The chart is designed to work with managed cloud services, ensuring reliability and scalability.

### Deployment Mode: HAProxy Ingress (Default)

The default configuration uses HAProxy Ingress controller which provides both Layer 4 (TCP passthrough) and Layer 7 (HTTP routing) capabilities.

**Architecture:**

```
Client
  ↓
HAProxy Ingress LoadBalancer
  ├─ Port 80 (TCP Layer 4 passthrough) → probo:80
  ├─ Port 443 (TCP Layer 4 passthrough) → probo:443
  └─ HTTP probo.example.com (Layer 7 routing) → probo:8080
```

**How it works:**

1. **HAProxy Ingress Controller** runs in the release namespace with a LoadBalancer service
2. **TCP ConfigMap** defines Layer 4 TCP passthrough rules:
   - Port 80 → Probo service port 80
   - Port 443 → Probo service port 443
3. **HTTP Ingress** defines Layer 7 HTTP routing:
   - Host `probo.example.com` → Probo service port 8080

**Benefits:**
- Single LoadBalancer for both TCP and HTTP traffic
- TCP passthrough for ports 80/443 (Probo handles TLS directly)
- HTTP routing for backoffice on port 8080
- Supports ACME/Let's Encrypt integration

**Port Configuration:**
- **Port 80** - TCP passthrough to Probo:80 (HTTP service, ACME challenges)
- **Port 443** - TCP passthrough to Probo:443 (HTTPS service with TLS)
- **Port 8080** - HTTP routing to Probo:8080 (Backoffice, via host-based routing)

## Configuration

### Required Configuration

All deployments require:

- `probo.encryptionKey` - For data encryption at rest
- `probo.auth.cookieSecret` - For session management
- `probo.auth.passwordPepper` - For password hashing
- `probo.trustAuth.tokenSecret` - For trust center tokens
- `postgresql.host` - PostgreSQL server hostname
- `postgresql.password` - Database password
- `s3.accessKeyId` - S3 access credentials
- `s3.secretAccessKey` - S3 secret key

See [values.yaml](charts/probo/values.yaml) for all available options.


## Cloud Provider Examples

### AWS
- PostgreSQL: Amazon RDS for PostgreSQL
- Storage: Amazon S3
- Kubernetes: Amazon EKS

#### Exemple
```bash
# Prerequisites:
# - Amazon RDS PostgreSQL instance
# - S3 bucket created

helm install my-probo ././charts/probo \
  --set probo.encryptionKey="$ENCRYPTION_KEY" \
  --set probo.auth.cookieSecret="$COOKIE_SECRET" \
  --set probo.auth.passwordPepper="$PASSWORD_PEPPER" \
  --set probo.trustAuth.tokenSecret="$TRUST_TOKEN_SECRET" \
  --set postgresql.host="mydb.abc123.us-east-1.rds.amazonaws.com" \
  --set postgresql.password="<rds-password>" \
  --set s3.region="us-east-1" \
  --set s3.bucket="my-probo-bucket" \
  --set s3.accessKeyId="<aws-access-key>" \
  --set s3.secretAccessKey="<aws-secret-key>"
```

### GCP
- PostgreSQL: Cloud SQL for PostgreSQL
- Storage: Cloud Storage with S3 compatibility
- Kubernetes: Google Kubernetes Engine (GKE)

#### Exemple

```bash
# Prerequisites:
# - Cloud SQL PostgreSQL instance
# - Cloud Storage bucket with HMAC keys

helm install my-probo ././charts/probo \
  --set probo.encryptionKey="$ENCRYPTION_KEY" \
  --set probo.auth.cookieSecret="$COOKIE_SECRET" \
  --set probo.auth.passwordPepper="$PASSWORD_PEPPER" \
  --set probo.trustAuth.tokenSecret="$TRUST_TOKEN_SECRET" \
  --set postgresql.host="10.0.0.5" \
  --set postgresql.password="<cloudsql-password>" \
  --set s3.endpoint="https://storage.googleapis.com" \
  --set s3.bucket="my-probo-bucket" \
  --set s3.accessKeyId="<hmac-access-key>" \
  --set s3.secretAccessKey="<hmac-secret>"
```

### Azure
- PostgreSQL: Azure Database for PostgreSQL
- Storage: Azure Blob Storage (with S3 compatibility)
- Kubernetes: Azure Kubernetes Service (AKS)

#### Example


```bash
# Prerequisites:
# - Azure Database for PostgreSQL instance
# - Azure Blob Storage container with S3 compatibility
helm install my-probo ././charts/probo \
  --set probo.encryptionKey="$ENCRYPTION_KEY" \
  --set probo.auth.cookieSecret="$COOKIE_SECRET" \
  --set probo.auth.passwordPepper="$PASSWORD_PEPPER" \
  --set probo.trustAuth.tokenSecret="$TRUST_TOKEN_SECRET" \
  --set postgresql.host="mydb.postgres.database.azure.com" \
  --set postgresql.password="<azure-db-password>" \
  --set s3.endpoint="https://<your-storage-account>.blob.core.windows.net" \
  --set s3.bucket="my-probo-bucket" \
  --set s3.accessKeyId="<azure-access-key>" \
  --set s3.secretAccessKey="<azure-secret-key>" \
  --set s3.usePathStyle=true
```

> **Note:** Azure Blob Storage requires `s3.usePathStyle=true` to construct URLs correctly (path-style: `https://account.blob.core.windows.net/container/...` instead of virtual-hosted style: `https://bucket.account.blob.core.windows.net/...`).

### DigitalOcean
- PostgreSQL: Managed PostgreSQL Database
- Storage: DigitalOcean Spaces
- Kubernetes: DigitalOcean Kubernetes (DOKS)

#### Exemple
```bash
# Prerequisites:
# - Managed PostgreSQL Database
# - Spaces bucket

helm install my-probo ././charts/probo \
  --set probo.encryptionKey="$ENCRYPTION_KEY" \
  --set probo.auth.cookieSecret="$COOKIE_SECRET" \
  --set probo.auth.passwordPepper="$PASSWORD_PEPPER" \
  --set probo.trustAuth.tokenSecret="$TRUST_TOKEN_SECRET" \
  --set postgresql.host="db-postgresql-nyc1-12345.ondigitalocean.com" \
  --set postgresql.password="<db-password>" \
  --set s3.region="nyc3" \
  --set s3.endpoint="https://nyc3.digitaloceanspaces.com" \
  --set s3.bucket="my-probo-bucket" \
  --set s3.accessKeyId="<spaces-key>" \
  --set s3.secretAccessKey="<spaces-secret>"
```

### Using External Secrets Operator

Example with AWS Secrets Manager:

```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: probo-secrets
spec:
  secretStoreRef:
    name: aws-secrets-manager
    kind: SecretStore
  target:
    name: probo
  data:
    - secretKey: encryption-key
      remoteRef:
        key: probo/encryption-key
    - secretKey: db-password
      remoteRef:
        key: probo/db-password
```

## Full Values

| Key                                                     | Type    | Default                                            | Description                                                                                         |
|---------------------------------------------------------| ------- |----------------------------------------------------|-----------------------------------------------------------------------------------------------------|
| replicaCount                                            | int     | `1`                                                | Number of Probo application replicas                                                                |
| image.repository                                        | string  | `"ghcr.io/getprobo/probo"`                         | Container image repository                                                                          |
| image.pullPolicy                                        | string  | `"IfNotPresent"`                                   | Image pull policy                                                                                   |
| image.tag                                               | string  | `"latest"`                                         | Overrides the image tag whose default is the chart appVersion                                       |
| imagePullSecrets                                        | list    | `[]`                                               | Image pull secrets for private registries                                                           |
| nameOverride                                            | string  | `""`                                               | Override the name of the chart                                                                      |
| fullnameOverride                                        | string  | `""`                                               | Override the fullname of the chart                                                                  |
| serviceAccount.create                                   | bool    | `true`                                             | Specifies whether a service account should be created                                               |
| serviceAccount.automount                                | bool    | `true`                                             | Automatically mount a ServiceAccount's API credentials                                              |
| serviceAccount.annotations                              | object  | `{}`                                               | Annotations to add to the service account                                                           |
| serviceAccount.name                                     | string  | `""`                                               | The name of the service account to use                                                              |
| podAnnotations                                          | object  | `{}`                                               | Annotations to add to pods                                                                          |
| podLabels                                               | object  | `{}`                                               | Labels to add to pods                                                                               |
| podSecurityContext.runAsUser                            | int     | `0`                                                | User ID to run the container as (0 = root, change to 1000 for production)                           |
| podSecurityContext.runAsGroup                           | int     | `0`                                                | Group ID to run the container as (0 = root, change to 1000 for production)                          |
| securityContext.runAsUser                               | int     | `0`                                                | User ID for the security context                                                                    |
| securityContext.runAsGroup                              | int     | `0`                                                | Group ID for the security context                                                                   |
| securityContext.privileged                              | bool    | `true`                                             | Enable privileged mode for the container (set to false for production)                              |
| securityContext.capabilities.drop                       | list    | `["ALL"]`                                          | Linux capabilities to drop from the container                                                       |
| securityContext.readOnlyRootFilesystem                  | bool    | `false`                                            | Mount root filesystem as read-only (/data directory requires write access)                          |
| service.type                                            | string  | `"ClusterIP"`                                      | Kubernetes service type                                                                             |
| service.port                                            | int     | `8080`                                             | Service port                                                                                        |
| service.httpPort                                        | int     | `80`                                               | HTTP port for TCP passthrough                                                                       |
| service.httpsPort                                       | int     | `443`                                              | HTTPS port for TCP passthrough                                                                      |
| service.annotations                                     | object  | `{}`                                               | Annotations for the service                                                                         |
| haproxy-ingress.enabled                                 | bool    | `true`                                             | Enable included HAProxy Ingress controller                                                          |
| haproxy-ingress.controller.ingressClass                 | string  | `"haproxy"`                                        | Ingress class name for the controller                                                               |
| haproxy-ingress.controller.ingressClassResource.enabled | bool | `true`                                      | Enable IngressClass resource creation                                                               |
| haproxy-ingress.controller.ingressClassResource.name    | string | `"haproxy"`                                  | Name of the IngressClass resource                                                                   |
| haproxy-ingress.controller.ingressClassResource.default | bool | `false`                                     | Set as default IngressClass                                                                         |
| haproxy-ingress.controller.service.type                 | string  | `"LoadBalancer"`                                   | HAProxy Ingress controller service type                                                             |
| haproxy-ingress.controller.extraArgs                    | list | `["--tcp-services-configmap=$(POD_NAMESPACE)/haproxy-tcp-services"]` | Extra arguments for HAProxy controller |
| haproxy-ingress.controller.config.ssl-redirect          | string | `"false"`                                         | Disable automatic SSL redirect                                                                      |
| ingress.enabled                                         | bool    | `true`                                             | Enable ingress resource                                                                             |
| ingress.className                                       | string  | `"haproxy"`                                        | Ingress class name                                                                                  |
| ingress.annotations                                     | object  | `{"kubernetes.io/ingress.class": "haproxy"}`       | Annotations for the ingress resource                                                                |
| ingress.hosts[0].host                                   | string  | `"probo.example.com"`                              | Ingress hostname                                                                                    |
| ingress.hosts[0].paths[0].path                          | string  | `"/"`                                              | Path to match                                                                                       |
| ingress.hosts[0].paths[0].pathType                      | string  | `"Prefix"`                                         | Path matching type                                                                                  |
| ingress.tls                                             | list    | `[]`                                               | TLS configuration for ingress                                                                       |
| resources.limits.cpu                                    | string  | `"2000m"`                                          | CPU limit for the container                                                                         |
| resources.limits.memory                                 | string  | `"2Gi"`                                            | Memory limit for the container                                                                      |
| resources.requests.cpu                                  | string  | `"500m"`                                           | CPU request for the container                                                                       |
| resources.requests.memory                               | string  | `"512Mi"`                                          | Memory request for the container                                                                    |
| livenessProbe.httpGet.path                              | string  | `"/"`                                              | Path for the liveness probe                                                                         |
| livenessProbe.httpGet.port                              | string  | `"http"`                                           | Port for the liveness probe                                                                         |
| livenessProbe.initialDelaySeconds                       | int     | `30`                                               | Initial delay for liveness probe                                                                    |
| livenessProbe.periodSeconds                             | int     | `10`                                               | Period for liveness probe                                                                           |
| livenessProbe.timeoutSeconds                            | int     | `5`                                                | Timeout for liveness probe                                                                          |
| livenessProbe.failureThreshold                          | int     | `6`                                                | Failure threshold for liveness probe                                                                |
| readinessProbe.httpGet.path                             | string  | `"/"`                                              | Path for the readiness probe                                                                        |
| readinessProbe.httpGet.port                             | string  | `"http"`                                           | Port for the readiness probe                                                                        |
| readinessProbe.initialDelaySeconds                      | int     | `10`                                               | Initial delay for readiness probe                                                                   |
| readinessProbe.periodSeconds                            | int     | `5`                                                | Period for readiness probe                                                                          |
| readinessProbe.timeoutSeconds                           | int     | `3`                                                | Timeout for readiness probe                                                                         |
| readinessProbe.failureThreshold                         | int     | `3`                                                | Failure threshold for readiness probe                                                               |
| autoscaling.enabled                                     | bool    | `false`                                            | Enable horizontal pod autoscaling                                                                   |
| autoscaling.minReplicas                                 | int     | `1`                                                | Minimum number of replicas                                                                          |
| autoscaling.maxReplicas                                 | int     | `10`                                               | Maximum number of replicas                                                                          |
| autoscaling.targetCPUUtilizationPercentage              | int     | `80`                                               | Target CPU utilization percentage                                                                   |
| volumes                                                 | list    | `[]`                                               | Additional volumes on the output Deployment definition                                              |
| volumeMounts                                            | list    | `[]`                                               | Additional volumeMounts on the output Deployment definition                                         |
| nodeSelector                                            | object  | `{}`                                               | Node selector for pod assignment                                                                    |
| tolerations                                             | list    | `[]`                                               | Tolerations for pod assignment                                                                      |
| affinity                                                | object  | `{}`                                               | Affinity rules for pod assignment                                                                   |
| persistence.enabled                                     | bool    | `false`                                            | Enable persistent volume for data storage                                                           |
| persistence.storageClass                                | string  | `""`                                               | Storage class for the persistent volume                                                             |
| persistence.accessMode                                  | string  | `"ReadWriteOnce"`                                  | Access mode for the persistent volume                                                               |
| persistence.size                                        | string  | `"10Gi"`                                           | Size of the persistent volume                                                                       |
| persistence.annotations                                 | object  | `{}`                                               | Annotations for the persistent volume claim                                                         |
| metrics.serviceMonitor.enabled                          | bool    | `false`                                            | Enable Prometheus Operator ServiceMonitor                                                           |
| metrics.serviceMonitor.interval                         | string  | `"30s"`                                            | Scrape interval for metrics                                                                         |
| metrics.serviceMonitor.scrapeTimeout                    | string  | `"10s"`                                            | Scrape timeout for metrics                                                                          |
| metrics.serviceMonitor.labels                           | object  | `{}`                                               | Labels for the ServiceMonitor                                                                       |
| metrics.serviceMonitor.relabelings                      | list    | `[]`                                               | Relabeling configs for the ServiceMonitor                                                           |
| probo.baseUrl                                           | string  | `"probo.example.com"`                              | Public hostname where Probo will be accessible                                                      |
| probo.encryptionKey                                     | string  | `""`                                               | **REQUIRED** Base64-encoded encryption key (generate with: openssl rand -base64 32)                 |
| probo.service.port                                      | int     | `8080`                                             | Probo application service port                                                                      |
| probo.metrics.port                                      | int     | `8081`                                             | Probo metrics service port                                                                          |
| probo.tracing.enabled                                   | bool    | `false`                                            | Enable OpenTelemetry tracing                                                                        |
| probo.tracing.addr                                      | string  | `""`                                               | OTLP gRPC endpoint (e.g., tempo:4317)                                                               |
| probo.tracing.maxBatchSize                              | int     | `512`                                              | Maximum batch size for trace exports                                                                |
| probo.tracing.batchTimeout                              | int     | `5`                                                | Batch timeout in seconds                                                                            |
| probo.tracing.exportTimeout                             | int     | `30`                                               | Export timeout in seconds                                                                           |
| probo.tracing.maxQueueSize                              | int     | `2048`                                             | Maximum queue size for traces                                                                       |
| probo.cors.allowedOrigins                               | list    | `["https://probo.example.com", "http://..."]`      | CORS allowed origins                                                                                |
| probo.extraHeaderFields                                 | object  | `{}`                                               | Extra HTTP headers to add to responses                                                              |
| probo.auth.disableSignup                                | bool    | `false`                                            | Disable user signup                                                                                 |
| probo.auth.invitationTokenValidity                      | int     | `3600`                                             | Invitation token validity in seconds                                                                |
| probo.auth.cookieName                                   | string  | `"SSID"`                                           | Authentication cookie name                                                                          |
| probo.auth.cookieDomain                                 | string  | `"probo.example.com"`                              | Authentication cookie domain                                                                        |
| probo.auth.cookieSecret                                 | string  | `""`                                               | **REQUIRED** Cookie signing secret (at least 32 bytes, generate with: openssl rand -base64 32)      |
| probo.auth.cookieDuration                               | int     | `24`                                               | Cookie duration in hours                                                                            |
| probo.auth.passwordPepper                               | string  | `""`                                               | **REQUIRED** Password hashing pepper (at least 32 bytes, generate with: openssl rand -base64 32)    |
| probo.auth.passwordIterations                           | int     | `1000000`                                          | Password hashing iterations                                                                         |
| probo.trustAuth.cookieName                              | string  | `"TCT"`                                            | Trust center cookie name                                                                            |
| probo.trustAuth.cookieDomain                            | string  | `"probo.example.com"`                              | Trust center cookie domain                                                                          |
| probo.trustAuth.cookieDuration                          | int     | `24`                                               | Trust center cookie duration in hours                                                               |
| probo.trustAuth.tokenDuration                           | int     | `168`                                              | Trust token duration in hours                                                                       |
| probo.trustAuth.reportUrlDuration                       | int     | `15`                                               | Report URL duration in minutes                                                                      |
| probo.trustAuth.tokenSecret                             | string  | `""`                                               | **REQUIRED** Trust token signing secret (at least 32 bytes, generate with: openssl rand -base64 32) |
| probo.trustAuth.scope                                   | string  | `"trust_center_readonly"`                          | Trust token scope                                                                                   |
| probo.trustAuth.tokenType                               | string  | `"trust_center_access"`                            | Trust token type                                                                                    |
| probo.mailer.senderName                                 | string  | `"Probo"`                                          | Email sender name                                                                                   |
| probo.mailer.senderEmail                                | string  | `"no-reply@notification.getprobo.com"`             | Email sender address                                                                                |
| probo.mailer.smtp.addr                                  | string  | `"sandbox.smtp.mailtrap.io:2525"`                  | SMTP server address                                                                                 |
| probo.mailer.smtp.user                                  | string  | `"2d1b1d0e8b3d0b"`                                 | SMTP username                                                                                       |
| probo.mailer.smtp.password                              | string  | `"25a8eb11e75e8d"`                                 | SMTP password                                                                                       |
| probo.mailer.smtp.tlsRequired                           | bool    | `true`                                             | Require TLS for SMTP connection                                                                     |
| probo.openai.apiKey                                     | string  | `""`                                               | OpenAI API key for AI features (optional)                                                           |
| probo.openai.temperature                                | float   | `0.1`                                              | OpenAI temperature setting                                                                          |
| probo.openai.modelName                                  | string  | `"gpt-4o"`                                         | OpenAI model name                                                                                   |
| probo.customDomains.enabled                             | bool    | `false`                                            | Enable custom domains feature                                                                       |
| probo.customDomains.renewalInterval                     | int     | `3600`                                             | Certificate renewal interval in seconds                                                             |
| probo.customDomains.provisionInterval                   | int     | `30`                                               | Domain provision interval in seconds                                                                |
| probo.customDomains.cnameTarget                         | string  | `"probo.example.com"`                              | CNAME target for custom domains                                                                     |
| probo.customDomains.acme.directory                      | string  | `"https://acme-v02.api.letsencrypt.org/directory"` | ACME directory URL                                                                                  |
| probo.customDomains.acme.email                          | string  | `"admin@example.com"`                              | ACME registration email                                                                             |
| probo.customDomains.acme.keyType                        | string  | `"EC256"`                                          | ACME key type                                                                                       |
| probo.customDomains.acme.insecureTls                    | bool    | `false`                                            | Allow insecure TLS for ACME                                                                         |
| probo.connectors                                        | list    | `[]`                                               | External OAuth2 connectors configuration                                                            |
| postgresql.enabled                                      | bool    | `true`                                             | Enable included PostgreSQL container for demo purposes using CloudNativePG                          |
| postgresql.resources.limits.memory                      | string  | `"2Gi"`                                            | PostgreSQL memory limit                                                                             |
| postgresql.resources.limits.ephemeral-storage           | string  | `"5Gi"`                                            | PostgreSQL ephemeral storage limit                                                                  |
| postgresql.auth.postgresUser                            | string  | `"probod"`                                         | PostgreSQL username                                                                                 |
| postgresql.auth.postgresPassword                        | string  | `"probod"`                                         | PostgreSQL password                                                                                 |
| postgresql.auth.database                                | string  | `"probod"`                                         | PostgreSQL database name                                                                            |
| s3.region                                               | string  | `"us-east-1"`                                      | S3 region                                                                                           |
| s3.bucket                                               | string  | `"probod"`                                         | S3 bucket name                                                                                      |
| s3.endpoint                                             | string  | `""`                                               | S3 endpoint (leave empty for AWS S3, set for S3-compatible storage)                                 |
| s3.accessKeyId                                          | string  | `""`                                               | **REQUIRED** (when minio.enabled=false) S3 access key                                               |
| s3.secretAccessKey                                      | string  | `""`                                               | **REQUIRED** (when minio.enabled=false) S3 secret key                                               |
| minio.enabled                                           | bool    | `true`                                             | Enable included MinIO for demo purposes (NOT for production)                                        |
| minio.mode                                              | string  | `"standalone"`                                     | MinIO mode                                                                                          |
| minio.persistence.enabled                               | bool    | `false`                                            | Enable MinIO persistence                                                                            |
| minio.resources.requests.memory                         | string  | `"128Mi"`                                          | MinIO memory request                                                                                |
| minio.primary.enabled                                   | bool    | `true`                                             | Enable MinIO primary node                                                                           |
| minio.auth.rootUser                                     | string  | `"probod"`                                         | MinIO root username                                                                                 |
| minio.auth.rootPassword                                 | string  | `"probod"`                                         | MinIO root password                                                                                 |
| chrome.enabled                                          | bool    | `true`                                             | Deploy Chrome headless in the cluster for PDF generation                                            |
| chrome.replicaCount                                     | int     | `1`                                                | Number of Chrome replicas                                                                           |
| chrome.image.repository                                 | string  | `"chromedp/headless-shell"`                        | Chrome container image repository                                                                   |
| chrome.image.tag                                        | string  | `"140.0.7259.2"`                                   | Chrome image tag                                                                                    |
| chrome.image.pullPolicy                                 | string  | `"IfNotPresent"`                                   | Chrome image pull policy                                                                            |
| chrome.resources.limits.cpu                             | string  | `"1000m"`                                          | Chrome CPU limit                                                                                    |
| chrome.resources.limits.memory                          | string  | `"1Gi"`                                            | Chrome memory limit                                                                                 |
| chrome.resources.requests.cpu                           | string  | `"100m"`                                           | Chrome CPU request                                                                                  |
| chrome.resources.requests.memory                        | string  | `"256Mi"`                                          | Chrome memory request                                                                               |
| chrome.external.addr                                    | string  | `""`                                               | External Chrome address (used when chrome.enabled=false)                                            |
