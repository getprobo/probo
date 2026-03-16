# Lima Sandbox Environments

Run isolated dev environments using [Lima](https://lima-vm.io/) VMs. Each git
worktree gets its own VM with Docker, standard ports, and a unique IP — no port
conflicts between sandboxes.

## Prerequisites

```bash
brew install lima jq
```

## Quickstart

```bash
# Create and start a sandbox for the current worktree
./contrib/lima/sandbox.sh create

# Check status and get the VM IP
./contrib/lima/sandbox.sh status

# Start the Docker stack inside the VM
./contrib/lima/sandbox.sh exec -- make stack-up

# Build and run the dev server
./contrib/lima/sandbox.sh exec -- make build
./contrib/lima/sandbox.sh exec -- make dev

# Access services from your host browser using the VM IP
# e.g. http://192.168.105.2:5173
```

## Commands

| Command | Description |
|---|---|
| `./contrib/lima/sandbox.sh create [--cpus C] [--memory M] [--disk D]` | Create and start a new VM |
| `./contrib/lima/sandbox.sh start` | Start a stopped VM |
| `./contrib/lima/sandbox.sh stop` | Shut down the VM (preserves disk and Docker images) |
| `./contrib/lima/sandbox.sh restart` | Stop + start |
| `./contrib/lima/sandbox.sh delete` | Remove the VM entirely |
| `./contrib/lima/sandbox.sh ssh` | Interactive shell at `/workspace` |
| `./contrib/lima/sandbox.sh exec -- CMD` | Run a command in the VM |
| `./contrib/lima/sandbox.sh status` | Show VM state, IP, and service URLs |
| `./contrib/lima/sandbox.sh list` | List all `probo-*` VMs |

## Architecture

```
Host (macOS)
├── worktree: ~/Developer/probo/delhi      → VM "probo-delhi"      (192.168.105.x)
├── worktree: ~/Developer/probo/feature-a  → VM "probo-feature-a"  (192.168.105.y)
└── worktree: ~/Developer/probo/feature-b  → VM "probo-feature-b"  (192.168.105.z)
```

Each VM:
- Mounts the worktree at `/workspace` via virtiofs (read-write)
- Runs Docker + docker-compose inside the VM
- Forwards the host SSH agent
- Gets its own IP via vzNAT — all services on standard ports

## Makefile targets

Convenience wrappers are available:

```bash
make sandbox-create
make sandbox-start
make sandbox-stop
make sandbox-delete
make sandbox-ssh
make sandbox-status
```

## Troubleshooting

**VM won't start**: Check `limactl list` for stale entries. Delete with
`./contrib/lima/sandbox.sh delete` and recreate.

**Slow file I/O**: The worktree is mounted via virtiofs which is fast for most
operations. If `node_modules` is slow, consider symlinking it to VM-local disk.

**Docker permission denied**: The provision script adds the lima user to the
docker group. If you see permission errors, restart the VM with
`./contrib/lima/sandbox.sh restart`.

**Can't reach VM IP from host**: Ensure vzNAT networking is working. Run
`./contrib/lima/sandbox.sh status` to verify the IP is assigned.
