# Contributing

Thank you for your interest in contributing to our project!

## How to Contribute

1. Fork the repository
2. Create a new branch for your changes
3. Make your changes
4. Sign your commits with DCO:
   ```sh
   git commit -s -m "Your commit message"
   ```
5. Push to your fork and submit a pull request

## Developer Certificate of Origin

We use the Developer Certificate of Origin (DCO) in lieu of a
Contributor License Agreement (CLA) for all contributions to this
project. DCO is a legally binding statement that asserts that you are
the creator of your contribution and that you wish to allow us to use
it in this project.

When you contribute to this repository with a pull request, you need
to sign-off that you agree to the DCO. You do this by adding a
`Signed-off-by` line to your commit messages containing your name and
email:

```
git commit -s -m "Your commit message"
```

This will automatically add a sign-off message to your
commit. Alternatively, you can manually add:

```sh
Signed-off-by: John Doe <john.doe@example.org>
```

## Code Guidelines

- Keep code clean and simple
- Follow existing code style
- Update documentation if needed

## Development Setup

### Prerequisites

- Go 1.21+
- Node.js 22+
- Docker
- mkcert (for SSL certificates)
- parallel (GNU Parallel for process management)
- gow (Go file watcher for auto-reload)

### Installation Steps

1. Clone the repository:

   ```bash
   git clone https://github.com/getprobo/probo.git
   cd probo
   ```

2. Install Go dependencies:

   ```bash
   go mod download
   ```

3. Install JavaScript dependencies:

   ```bash
   npm ci
   ```

4. Start the Docker infrastructure stack:

   ```bash
   make stack-up
   ```

### Running the Development Environment

The fastest way to develop is to use the `make dev` command, which starts the Go backend and frontend dev servers with hot module replacement (HMR):

```bash
# Start Docker services first
make stack-up

# Then start all dev servers with one command
make dev
```

This starts 3 processes:
- **Go backend** with auto-reload (gow) on `http://localhost:8080`
- **Console frontend dev server** (Vite) on `http://localhost:5173`
- **Trust center dev server** (Vite) on `http://localhost:5174`

Backend automatically proxies to Vite servers, so you get:
- ⚡ **No TypeScript builds needed** - Skip the long build step
- 🔄 **Hot Module Replacement** - Changes appear instantly in browser
- 🚀 **Fast iteration** - 3-5 second backend rebuild vs 58+ second full build
- 📝 **One command** - All services managed together

### Alternative: Manual Development

If you prefer to run services separately:

```bash
# Terminal 1 - Start the API server
bin/probod -cfg-file cfg/dev.yaml

# Terminal 2 - Start the console frontend dev server
npm -w @probo/console run dev

# Terminal 3 - Start the trust frontend dev server (optional)
npm -w @probo/trust run dev
```

If running services separately, set environment variables to enable dev mode:

```bash
# Terminal 1 - Start with dev proxies
VITE_DEV_SERVER_CONSOLE=http://localhost:5173 \
VITE_DEV_SERVER_TRUST=http://localhost:5174 \
bin/probod -cfg-file cfg/dev.yaml
```

### Building for Production

To build the project with optimized frontend bundles:

```bash
make build
```

For detailed information about all Docker services in the development stack, see [Docker Services Documentation](docs/DOCKER_SERVICES.md).

## Need Help?

Create an issue if you:

- Found a bug
- Have a feature request
- Need help with something

Thank you for contributing!
