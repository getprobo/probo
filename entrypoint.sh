#!/bin/bash
set -e

# Configuration file path
CONFIG_FILE="${CONFIG_FILE:-/etc/probod/config.yml}"

# Check if config file already exists (e.g., mounted from ConfigMap)
if [ -f "$CONFIG_FILE" ]; then
  echo "Using existing configuration file at: $CONFIG_FILE"
else
  echo "Generating configuration file from environment variables at: $CONFIG_FILE"
  # Generate configuration from environment variables
  probod-bootstrap -output "$CONFIG_FILE"
fi

# Execute probod with the generated config
exec probod -cfg-file "$CONFIG_FILE" "$@"
