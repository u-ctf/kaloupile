# Kaloupile (dev helper)

Kaloupile is a small development CLI meant to run alongside a Tiltfile. It bootstraps a local Kind cluster, installs K8s dependencies, applies routes, and runs sync tasks.

## Quick start

1. `go run ./cmd setup`
2. `go run ./cmd dependencies`
3. `go run ./cmd routes`
4. `go run ./cmd sync`

## Requirements

- `kind` and `kubectl` in PATH
- A working Go toolchain

## Commands

- `setup`
  - Installs prerequisites from [cluster/prerequisites/install-prerequisites.sh](cluster/prerequisites/install-prerequisites.sh)
  - Ensures a Kind cluster named `kaloupile-dev` exists
  - Installs cert-manager from [cert-manager.yaml](https://github.com/cert-manager/cert-manager/releases/download/v1.19.2/cert-manager.yaml)
  - Installs Infomaniak webhook from [rendered-manifest.yaml](https://github.com/infomaniak/cert-manager-webhook-infomaniak/releases/download/v0.2.0/rendered-manifest.yaml)

- `dependencies`
  - Loads [config.yml](config.yml)
  - Creates Infomaniak API Secret from [cluster/dependencies/cert-manager/infomaniak-api-credentials.yaml](cluster/dependencies/cert-manager/infomaniak-api-credentials.yaml)
  - Creates Certificate from [cluster/dependencies/cert-manager/certificate.yaml](cluster/dependencies/cert-manager/certificate.yaml)
  - Installs PostgreSQL from [cluster/dependencies/postgresql/postgresql.yaml](cluster/dependencies/postgresql/postgresql.yaml)
  - Installs Fake SMTP from [cluster/dependencies/fake-smtp/fake-smtp.yaml](cluster/dependencies/fake-smtp/fake-smtp.yaml)

- `routes`
  - Loads [config.yml](config.yml)
  - Applies routes from [cluster/routes/routes.yaml](cluster/routes/routes.yaml)

- `sync`
  - Loads [config.yml](config.yml)
  - Syncs PostgreSQL users and databases

- `cleanup`
  - Deletes the Kind cluster (kaloupile-dev)

## Configuration

- Default config: [config.yml](config.yml)
- Override with `--config <path>`

The PostgreSQL sync supports the standard Postgres env overrides:
- `PGHOST`
- `PGPORT`
- `PGSSLMODE`

DNS settings require an Infomaniak API token:

```yaml
dns:
  infomaniak:
    token: "your-infomaniak-api-token"
```

## Kind configuration

The Kind cluster is created using [kind-config.yml](kind-config.yml). It includes port mappings for HTTP, HTTPS, and PostgreSQL.

## Notes

- The CLI assumes an `external` namespace for dependencies.
