# step-ca local CA

This directory holds the persistent [step-ca](https://github.com/smallstep/certificates)
state for local custom-domain TLS. It is initialized on first `make stack-up`.

After the first run, install the root CA once so browsers and server-side TLS
clients (e.g. CIMD OAuth) trust issued certificates across restarts:

```bash
step certificate install compose/step-ca/certs/root_ca.crt
```

The ACME directory URL is `https://localhost:9000/acme/acme/directory`.

step-ca shares the `acme-http-01-proxy` container network so HTTP-01 validation
to `http://<hostname>/.well-known/acme-challenge/...` reaches Caddy on port 80,
which forwards to probod's trust-center HTTP listener on the host.

## Custom domain DNS (optional)

Managed compliance-page domains (`*.probopage.localhost`) resolve via the
`.localhost` TLD and skip DNS checks.

For customer custom domains in local dev, point DNS at the host via
`/etc/hosts` and ensure HTTP-01 reaches probod through the
`acme-http-01-proxy` service on port 80.
