# Kaloupile CLI (dev)

This CLI is a small dev helper meant to be used alongside a Tiltfile. It focuses on bootstrapping a local Kind cluster, installing K8s dependencies, applying routes, and running sync tasks.

## Commands

- `setup`
  - Runs the prerequisites script and ensures a Kind cluster named `kaloupile-dev` exists.
  - Idempotent: once it runs, it writes a marker file at `.kaloupile/setup.done` and exits early on subsequent runs.
  - Uses `cluster/prerequisites/install-prerequisites.sh`.

- `dependencies`
  - Loads `config.yml` and installs dependencies.
  - Applies the PostgreSQL manifest template at `cluster/dependencies/postgresql/postgresql.yaml`.
  - Applies the Fake SMTP manifest at `cluster/dependencies/fake-smtp/fake-smtp.yaml`.

- `routes`
  - Loads `config.yml` and applies route templates.
  - Uses `cluster/routes/routes.yaml`.

- `sync`
  - Loads `config.yml` and runs sync tasks.
  - Currently syncs PostgreSQL users and databases.

- `cleanup`
  - Deletes the Kind cluster (the CLI is configured to use `kaloupile-dev`).

## Flags

- `--config` (default: `config.yml`)
  - Path to the config file used by `dependencies`, `routes`, and `sync`.

## Notes

- `kubectl` and `kind` must be available in PATH.
- The CLI assumes an `external` namespace for dependencies.
- The setup marker is not removed by `cleanup`.

---

# Memories

This section stores key architectural decisions, patterns, and lessons learned. Future AI agents should review these memories before making changes to avoid repeating past mistakes or deviating from established conventions.

## Architectural Decisions

*No entries yet. Add decisions here as they are made.*

## Patterns & Conventions

*No entries yet. Document recurring patterns here.*

## Lessons Learned / Mistakes to Avoid

*No entries yet. Record pitfalls and their solutions here.*
