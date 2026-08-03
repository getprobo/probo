#!/bin/bash
set -euo pipefail

# ---------------------------------------------------------------------------
# Render-specific glue.
#
# 1. Hostname: Render assigns the public URL at deploy time and injects it as
#    RENDER_EXTERNAL_URL (e.g. https://govrly-keycloak.onrender.com). Keycloak 26
#    accepts a full URL in KC_HOSTNAME, so we hand it straight through instead of
#    hardcoding anything. Nothing here needs to change if the service is renamed.
#
# 2. Database: Render exposes Postgres as postgres:// style values, but Keycloak
#    wants a JDBC URL. We therefore wire the individual DB_* vars in render.yaml
#    (fromDatabase host/port/database/user/password) and assemble the JDBC URL here.
# ---------------------------------------------------------------------------

if [[ -z "${RENDER_EXTERNAL_URL:-}" ]]; then
  echo "WARN: RENDER_EXTERNAL_URL is not set — falling back to KC_HOSTNAME=${KC_HOSTNAME:-<unset>}" >&2
else
  export KC_HOSTNAME="$RENDER_EXTERNAL_URL"
fi

for var in DB_HOST DB_PORT DB_NAME DB_USER DB_PASSWORD; do
  if [[ -z "${!var:-}" ]]; then
    echo "FATAL: required environment variable $var is not set" >&2
    exit 1
  fi
done

export KC_DB_URL="jdbc:postgresql://${DB_HOST}:${DB_PORT}/${DB_NAME}"
export KC_DB_USERNAME="$DB_USER"
export KC_DB_PASSWORD="$DB_PASSWORD"

# ---------------------------------------------------------------------------
# Admin recovery hatch.
#
# KC_BOOTSTRAP_ADMIN_USERNAME/PASSWORD only take effect on the boot that creates
# the master realm. If that first boot happened without them set, master exists
# with no admin user and Keycloak will never retry — you are locked out.
#
# Set KC_BOOTSTRAP_ADMIN_RECOVER=1 on the service and redeploy to force the
# admin to be (re)created, then remove the variable. Safe to leave failing: the
# command errors harmlessly if the user already exists, and we never abort boot.
# ---------------------------------------------------------------------------
if [[ -n "${KC_BOOTSTRAP_ADMIN_RECOVER:-}" ]]; then
  echo "KC_BOOTSTRAP_ADMIN_RECOVER set — creating admin '${KC_BOOTSTRAP_ADMIN_USERNAME:-admin}'"
  /opt/keycloak/bin/kc.sh bootstrap-admin user \
      --username "${KC_BOOTSTRAP_ADMIN_USERNAME:-admin}" \
      --password:env KC_BOOTSTRAP_ADMIN_PASSWORD \
    && echo "bootstrap-admin: OK — remove KC_BOOTSTRAP_ADMIN_RECOVER now" \
    || echo "bootstrap-admin: FAILED (see error above; user may already exist)"
fi

echo "Starting Keycloak"
echo "  KC_HOSTNAME  = ${KC_HOSTNAME:-<unset>}"
echo "  KC_DB_URL    = ${KC_DB_URL}"
echo "  KC_HTTP_PORT = ${KC_HTTP_PORT:-8080}"

exec /opt/keycloak/bin/kc.sh start --optimized --import-realm
