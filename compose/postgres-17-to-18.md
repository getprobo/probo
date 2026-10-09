# Upgrade the Compose Postgres volume from 17 to 18

This applies only to the Postgres container in `compose.prod.yaml`. It does not upgrade an external database.

**Back up that database before running the commands below.** With the Postgres 17 container still up, write the dump outside the repository:

```bash
docker compose -f compose.prod.yaml exec postgres pg_dump -U postgres -Fc -f /tmp/probod.dump probod
docker compose -f compose.prod.yaml cp postgres:/tmp/probod.dump "${HOME}/probod.dump"
```

Then run this from the repository root, once `compose.prod.yaml` mounts `postgres-data` at `/var/lib/postgresql`. Postgres 18 will not open a 17 cluster in place. The block stops on the first failure, so Compose is not started when the move or `pg_upgrade` fails.

```bash
set -euo pipefail

docker compose -f compose.prod.yaml stop probo postgres

vol="$(docker compose -f compose.prod.yaml ps -aq postgres | xargs docker inspect -f '{{ range .Mounts }}{{ if or (eq .Destination "/var/lib/postgresql") (eq .Destination "/var/lib/postgresql/data") }}{{ .Name }}{{ end }}{{ end }}')"

docker run --rm --entrypoint bash \
  -v "$vol":/var/lib/postgresql \
  postgres@sha256:52e6ffd11fddd081ae63880b635b2a61c14008c17fc98cdc7ce5472265516dd0 \
  -c 'mkdir -p /var/lib/postgresql/17/data && find /var/lib/postgresql -mindepth 1 -maxdepth 1 ! -name 17 ! -name 18 -exec mv {} /var/lib/postgresql/17/data/ \;'

docker run --rm \
  -v "$vol":/var/lib/postgresql \
  tianon/postgres-upgrade:17-to-18 \
  --link

docker compose -f compose.prod.yaml up -d
```
