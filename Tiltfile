# ============================================================================
# Configuration Generation
# ============================================================================

# Generate configs from templates using config.yml
local_resource(
    'generate-config',
    cmd='go run ./cmd/genconfig .',
    deps=['config.yml', 'external/kratos/kratos-config.yaml.tmpl', 'external/hydra/hydra-config.yaml.tmpl', 'external/postgresql/sync.sh.tmpl', 'external/mailslurper/mailslurper-config.yaml.tmpl', 'routes/routes.yaml.tmpl'],
    labels=['config'],
)

# ============================================================================
# External Dependencies
# ============================================================================

# PostgreSQL
k8s_yaml('external/postgresql/postgresql.yaml')
k8s_resource(
    'postgresql',
    port_forwards=['5432:5432'],
    labels=['external'],
)

# ============================================================================
# Ory Kratos
# ============================================================================

k8s_yaml('external/kratos/kratos.k8s.yaml')
k8s_yaml('external/kratos/kratos-config.generated.yaml')
k8s_yaml('external/mailslurper/mailslurper-config.generated.yaml')
k8s_resource(
    'kratos-migrate',
    resource_deps=['postgresql', 'pg-ready'],
    labels=['ory'],
)
k8s_resource(
    'kratos',
    port_forwards=['4433:4433', '4434:4434'],
    resource_deps=['kratos-migrate', 'generate-config'],
    labels=['ory'],
)
k8s_resource(
    'mailslurper',
    port_forwards=['4436:4436', '4437:4437'],
    resource_deps=['generate-config'],
    labels=['ory'],
)

# ============================================================================
# Ory Hydra
# ============================================================================

k8s_yaml('external/hydra/hydra.k8s.yaml')
k8s_yaml('external/hydra/hydra-config.generated.yaml')
k8s_resource(
    'hydra-migrate',
    resource_deps=['postgresql', 'pg-ready'],
    labels=['ory'],
)
k8s_resource(
    'hydra',
    port_forwards=['4444:4444', '4445:4445'],
    resource_deps=['hydra-migrate', 'generate-config'],
    labels=['ory'],
)

# ============================================================================
# Application Resources (add your app resources here)
# ============================================================================

# Kratos Self-Service UI
k8s_yaml('app/kratos-ui/kratos-ui.k8s.yaml')
k8s_resource(
    'kratos-ui',
    port_forwards=['3000:3000'],
    resource_deps=['kratos'],
    labels=['app'],
)

# ============================================================================
# HTTPRoutes
# ============================================================================

k8s_yaml('routes/routes.generated.yaml')

# ============================================================================
# Local Resources
# ============================================================================

# Health check for PostgreSQL
local_resource(
    'pg-ready',
    cmd='kubectl wait --for=condition=ready pod -l app=postgresql -n external --timeout=60s',
    resource_deps=['postgresql'],
    labels=['health'],
)

# Sync PostgreSQL users and databases from config
local_resource(
    'pg-sync',
    cmd='PGHOST=localhost PGPORT=5432 bash external/postgresql/sync.generated.sh',
    deps=['external/postgresql/sync.generated.sh'],
    resource_deps=['pg-ready', 'generate-config'],
    labels=['external'],
)

# ============================================================================
# Config Reload - Restart pods when their config changes
# ============================================================================

local_resource(
    'kratos-reload',
    cmd='kubectl rollout restart deployment/kratos -n external',
    deps=['external/kratos/kratos-config.generated.yaml'],
    resource_deps=['kratos'],
    labels=['reload'],
)

local_resource(
    'hydra-reload',
    cmd='kubectl rollout restart deployment/hydra -n external',
    deps=['external/hydra/hydra-config.generated.yaml'],
    resource_deps=['hydra'],
    labels=['reload'],
)

local_resource(
    'mailslurper-reload',
    cmd='kubectl rollout restart deployment/mailslurper -n external',
    deps=['external/mailslurper/mailslurper-config.generated.yaml'],
    resource_deps=['mailslurper'],
    labels=['reload'],
)
