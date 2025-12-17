# kaloupile

Local development environment for kaloupile with Kubernetes (KinD), Tilt, and Ory authentication stack.

## Local Development

### Prerequisites

Make sure you have the following tools installed:

- [Docker](https://docs.docker.com/get-docker/)
- [KinD](https://kind.sigs.k8s.io/docs/user/quick-start/#installation) (Kubernetes in Docker)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Helm](https://helm.sh/docs/intro/install/)
- [Tilt](https://docs.tilt.dev/install.html)
- [Go 1.23+](https://golang.org/doc/install)

### Getting Started

1. **Copy and configure** `config.yml`:
   ```bash
   cp config.example.yml config.yml
   # Edit config.yml with your settings (domain, secrets, etc.)
   ```

2. **First-time setup** - Creates the KinD cluster and installs prerequisites:
   ```bash
   make setup
   # or
   go run ./cmd/setup .
   ```

3. **Start development environment**:
   ```bash
   make dev
   # or
   tilt up
   ```

4. Open the Tilt dashboard at http://localhost:10350

### Configuration

All configuration is centralized in `config.yml`. This file is used to generate:
- Kratos configuration (`external/kratos/kratos-config.generated.yaml`)
- Hydra configuration (`external/hydra/hydra-config.generated.yaml`)
- Mailslurper configuration (`external/mailslurper/mailslurper-config.generated.yaml`)
- PostgreSQL user sync script (`external/postgresql/sync.generated.sh`)
- HTTPRoutes (`routes/routes.generated.yaml`)

Key configuration options:

| Setting | Description |
|---------|-------------|
| `scheme` | `http` or `https` |
| `domain` | Base domain for services (e.g., `tristan.dev.uctf.io`) |
| `postgres.users` | List of PostgreSQL users and their databases |
| `kratos.*` | Kratos ports, secrets, and webhook settings |
| `hydra.*` | Hydra ports and secrets |

### What Gets Installed

#### Prerequisites (via `make setup`)

1. **Gateway API CRDs** - Experimental channel for HTTPRoute support
2. **kGateway** - Kubernetes Gateway implementation
3. **Main Gateway** - In `gateway` namespace, allowing routes from all namespaces
4. **Namespaces** - `external`, `app`, `gateway`

#### External Services (via Tilt)

| Service | Description | Ports |
|---------|-------------|-------|
| **PostgreSQL 16** | Database for Kratos and Hydra | `5432` |
| **Ory Kratos** | Identity & User Management | `4433` (public), `4434` (admin) |
| **Ory Hydra** | OAuth2 & OpenID Connect | `4444` (public), `4445` (admin) |
| **Mailslurper** | Development SMTP server | `4436` (web), `4437` (API), `1025` (SMTP) |

#### Applications

| Service | Description | Port |
|---------|-------------|------|
| **Kratos UI** | Self-service UI for login/registration | `3000` |

#### HTTPRoutes

Routes are automatically configured for:
- `kratos.<domain>` → Kratos public API
- `<domain>/auth/*` → Kratos UI (login, registration, settings, etc.)

### Port Forwards

When Tilt is running, these ports are forwarded to localhost:

| Port | Service |
|------|---------|
| `3000` | Kratos UI |
| `4433` | Kratos Public API |
| `4434` | Kratos Admin API |
| `4436` | Mailslurper Web UI |
| `4437` | Mailslurper API |
| `4444` | Hydra Public API |
| `4445` | Hydra Admin API |
| `5432` | PostgreSQL |

### Project Structure

```
kaloupile/
├── cmd/
│   ├── setup/              # First-time setup (KinD + prerequisites)
│   └── genconfig/          # Config template generator
├── prerequisite/           # Pre-Tilt cluster setup
│   ├── namespaces/         # Namespace definitions
│   ├── gateway-api/        # Gateway API CRDs installation
│   ├── kgateway/           # kGateway installation
│   └── gateway/            # Main Gateway deployment
├── external/               # External dependencies (managed by Tilt)
│   ├── postgresql/         # PostgreSQL manifests + sync script
│   ├── kratos/             # Kratos manifests + config template
│   ├── hydra/              # Hydra manifests + config template
│   └── mailslurper/        # Mailslurper config template
├── app/                    # Application deployments
│   └── kratos-ui/          # Kratos self-service UI
├── routes/                 # HTTPRoute templates
├── config.yml              # Main configuration file (gitignored)
├── config.example.yml      # Example configuration
├── kind-config.yaml        # KinD cluster configuration
├── Tiltfile                # Tilt development orchestration
└── Makefile                # Development commands
```

### Commands

| Command | Description |
|---------|-------------|
| `make setup` | Run first-time setup (creates KinD cluster + prerequisites) |
| `make dev` | Start Tilt development environment |
| `make clean` | Delete the KinD cluster |

### Automatic Config Reload

When you modify `config.yml`, Tilt will:
1. Regenerate all configuration files
2. Apply updated ConfigMaps to the cluster
3. Automatically restart affected pods (Kratos, Hydra, Mailslurper)
