# Configuration Note

## Environment Variable Substitution

The Helm chart templates use a `-from-env` suffix pattern in the configuration file (e.g., `encryption-key-from-env: "ENCRYPTION_KEY"`). This pattern assumes the Probo application supports reading secrets from environment variables instead of directly from the config file.

### Current Implementation

The chart currently:
1. Stores sensitive values in a Kubernetes Secret
2. Mounts them as environment variables in the pod
3. References them in the config with `-from-env` suffix

### If Probo Doesn't Support `-from-env` Pattern

If the current Probo application doesn't support the `-from-env` pattern, you have three options:

#### Option 1: Implement in Probo (Recommended)

Update the Probo configuration parser to recognize `-from-env` suffixes and read from environment variables. This is the most secure approach.

Example in Go:
```go
// In config loading code
if strings.HasSuffix(key, "-from-env") {
    envVar := value.(string)
    actualValue := os.Getenv(envVar)
    // Use actualValue instead of value
}
```

#### Option 2: Use initContainer with envsubst

Modify the deployment to use an init container that substitutes environment variables:

```yaml
initContainers:
  - name: config-init
    image: alpine:3.19
    command:
      - sh
      - -c
      - |
        apk add --no-cache gettext
        envsubst < /config-template/config.yaml > /config/config.yaml
    env:
      # All environment variables from secrets
      - name: ENCRYPTION_KEY
        valueFrom: ...
    volumeMounts:
      - name: config-template
        mountPath: /config-template
      - name: config
        mountPath: /config
```

Then update the ConfigMap to use `${ENCRYPTION_KEY}` instead of the `-from-env` pattern.

#### Option 3: Mount Everything as Secret

Create a complete config.yaml as a Secret (not ConfigMap) with actual values:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: probo-config
stringData:
  config.yaml: |
    probod:
      encryption-key: {{ .Values.probo.encryptionKey }}
      # ... all config with actual values
```

This is simpler but less secure as the entire config is stored as a Secret.

## Recommendation

We recommend implementing Option 1 in the Probo codebase as it:
- Follows Kubernetes best practices
- Keeps secrets separate from configuration
- Works well with secret management tools (Vault, External Secrets, etc.)
- Supports secret rotation without config changes

## Current Status

The Helm chart is built with Option 1 in mind. If the Probo application doesn't yet support `-from-env`, please implement it or use one of the alternative approaches above.
