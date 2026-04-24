#!/bin/bash
# Copyright (c) 2025, 2026 Probo Inc.
# SPDX-License-Identifier: ISC

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
TEMPLATE="${SCRIPT_DIR}/probo.yaml"

REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
WORKTREE_NAME="$(basename "${REPO_ROOT}")"
VM_NAME="probo-${WORKTREE_NAME}"

usage() {
    cat <<EOF
Usage: $(basename "$0") <command> [options]

Commands:
  create [--cpus C] [--memory M] [--disk D]   Create a new sandbox VM
  start                                        Start a stopped sandbox
  stop                                         Stop the sandbox
  restart                                      Stop + start the sandbox
  delete                                       Delete the sandbox entirely
  ssh                                          Open interactive shell in /workspace
  exec -- CMD                                  Run a command in the sandbox
  status                                       Show sandbox state, IP, and services
  list                                         List all probo-* VMs

VM name: ${VM_NAME} (derived from worktree directory)
EOF
    exit 1
}

get_vm_ip() {
    limactl shell "${VM_NAME}" ip -4 -j addr show dev lima0 2>/dev/null \
        | jq -r '.[0].addr_info[0].local // empty' 2>/dev/null || true
}

get_vm_status() {
    local status
    status=$(limactl list --json 2>/dev/null \
        | jq -r "select(.name == \"${VM_NAME}\") | .status" 2>/dev/null) || true
    echo "${status:-NotFound}"
}

cmd_create() {
    local cpus="" memory="" disk=""

    while [[ $# -gt 0 ]]; do
        case "$1" in
            --cpus)   cpus="$2"; shift 2 ;;
            --memory) memory="$2"; shift 2 ;;
            --disk)   disk="$2"; shift 2 ;;
            *)        echo "Unknown option: $1"; usage ;;
        esac
    done

    echo "Creating sandbox: ${VM_NAME}"
    echo "Worktree: ${REPO_ROOT}"

    local -a create_args=(
        --name "${VM_NAME}"
        --tty=false
        --set ".mounts = [{\"location\": \"${REPO_ROOT}\", \"mountPoint\": \"/workspace\", \"writable\": true}]"
        --mount-type virtiofs
    )

    if [[ -n "${cpus}" ]]; then
        create_args+=(--cpus "${cpus}")
    fi
    if [[ -n "${memory}" ]]; then
        create_args+=(--memory "${memory}")
    fi
    if [[ -n "${disk}" ]]; then
        create_args+=(--disk "${disk}")
    fi

    limactl create "${create_args[@]}" "${TEMPLATE}"
    limactl start "${VM_NAME}"

    echo ""
    cmd_status
}

cmd_start() {
    echo "Starting sandbox: ${VM_NAME}"
    limactl start "${VM_NAME}"
    echo ""
    cmd_status
}

cmd_stop() {
    echo "Stopping sandbox: ${VM_NAME}"
    limactl stop "${VM_NAME}"
    echo "Sandbox stopped."
}

cmd_restart() {
    cmd_stop
    echo ""
    cmd_start
}

cmd_delete() {
    echo "Deleting sandbox: ${VM_NAME}"
    limactl delete --force "${VM_NAME}"
    echo "Sandbox deleted."
}

cmd_ssh() {
    exec limactl shell --workdir /workspace "${VM_NAME}"
}

cmd_exec() {
    limactl shell --workdir /workspace "${VM_NAME}" "$@"
}

cmd_status() {
    local status ip
    status="$(get_vm_status)"
    ip="$(get_vm_ip)"

    echo "Sandbox: ${VM_NAME}"
    echo "State:   ${status}"
    echo "IP:      ${ip:-"-"}"

    if [[ "${status}" == "Running" && -n "${ip}" ]]; then
        echo ""
        echo "Services (use VM IP to access from host):"
        echo "  Console:    http://${ip}:5173"
        echo "  Trust:      http://${ip}:5174"
        echo "  API:        http://${ip}:8080"
        echo "  Grafana:    http://${ip}:3001"
        echo "  Mailpit:    http://${ip}:8025"
        echo "  Keycloak:   http://${ip}:8082"
        echo "  PostgreSQL: psql -h ${ip} -U probod"
    fi
}

cmd_list() {
    printf "%-25s %-12s %s\n" "NAME" "STATE" "IP"

    limactl list --json 2>/dev/null | jq -r '
        select(.name | startswith("probo-")) |
        [.name, .status] | @tsv
    ' | while IFS=$'\t' read -r name status; do
        local ip="-"
        if [[ "${status}" == "Running" ]]; then
            ip=$(limactl shell "${name}" ip -4 -j addr show dev lima0 2>/dev/null \
                | jq -r '.[0].addr_info[0].local // "-"' 2>/dev/null || echo "-")
        fi
        printf "%-25s %-12s %s\n" "${name}" "${status}" "${ip}"
    done
}

if [[ $# -lt 1 ]]; then
    usage
fi

command="$1"
shift

case "${command}" in
    create)  cmd_create "$@" ;;
    start)   cmd_start ;;
    stop)    cmd_stop ;;
    restart) cmd_restart ;;
    delete)  cmd_delete ;;
    ssh)     cmd_ssh ;;
    exec)
        if [[ "${1:-}" == "--" ]]; then shift; fi
        cmd_exec "$@"
        ;;
    status)  cmd_status ;;
    list)    cmd_list ;;
    *)       echo "Unknown command: ${command}"; usage ;;
esac
