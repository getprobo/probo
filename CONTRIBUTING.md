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

4. Build the project:

   ```bash
   make build
   ```

5. Start docker containers:

   ```bash
   make stack-up
   ```

6. Generate the local dev config (writes `cfg/dev.yaml`):

   ```bash
   # Optional: override any dev default (secrets, OAuth clients, LLM keys).
   # cp .env.example .env && $EDITOR .env

   make dev-config
   ```

   The target stashes a dev-only RSA signing key under `cfg/.dev-oauth2-signing-key.pem` so tokens survive probod restarts, and sources `.env` if present so you can override defaults without editing the Makefile. `cfg/dev.yaml`, `.env`, and the signing key are all gitignored. Re-run the target to regenerate.

7. Start the development servers:

   ```bash
   # In one terminal - start the API server
   bin/probod -cfg-file cfg/dev.yaml

   # In another terminal - start the frontend
   npm -w @probo/console run dev
   ```

The application should now be running at `http://localhost:3000`

For detailed information about all Docker services in the development stack, see [Docker Services Documentation](docs/DOCKER_SERVICES.md).

## Need Help?

Create an issue if you:

- Found a bug
- Have a feature request
- Need help with something

Thank you for contributing!
