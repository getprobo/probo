# Keycloak 26 for the Render MFA / OIDC-IdP test deployment (see README.md).
# Image tag and digest are kept in sync with the local dev Keycloak in compose.yaml.
#
# ---------------------------------------------------------------------------
# Stage 1 — optimized build
# `kc.sh build` bakes the build-time options (db vendor, health) into the image
# so the container can start with `--optimized` and skip re-augmentation.
# ---------------------------------------------------------------------------
FROM quay.io/keycloak/keycloak:26.6.1@sha256:26ae26445475f7fac5f90ee138b1bdb64324f5815fb16133ffdbdb122d97c4d8 AS builder

WORKDIR /opt/keycloak

RUN /opt/keycloak/bin/kc.sh build \
      --db=postgres \
      --health-enabled=true

# ---------------------------------------------------------------------------
# Stage 2 — runtime
# ---------------------------------------------------------------------------
FROM quay.io/keycloak/keycloak:26.6.1@sha256:26ae26445475f7fac5f90ee138b1bdb64324f5815fb16133ffdbdb122d97c4d8

COPY --from=builder /opt/keycloak/ /opt/keycloak/

# Imported on first boot via `--import-realm` (no-op once the realm exists).
COPY realm-export/govrly-test-realm.json /opt/keycloak/data/import/govrly-test-realm.json

COPY entrypoint.sh /opt/keycloak/entrypoint.sh

USER root
RUN chown 1000:0 /opt/keycloak/entrypoint.sh /opt/keycloak/data/import/govrly-test-realm.json \
 && chmod 755 /opt/keycloak/entrypoint.sh
USER 1000

# Render routes to this port; also set via KC_HTTP_PORT in render.yaml.
EXPOSE 10000

ENTRYPOINT ["/opt/keycloak/entrypoint.sh"]
