# Kaloupile dev Tiltfile

KALOUPILE_BIN = "bin/kaloupile"
SYNC_RETRY_SECONDS = 7

local_resource(
    name = "build-kaloupile",
    cmd = "mkdir -p bin && go build -o %s ./cmd" % KALOUPILE_BIN,
    deps = [
        "cmd",
        "pkg",
        "go.mod",
        "go.sum",
    ],
    labels = ["cluster"]
)

local_resource(
    name = "setup",
    cmd = "%s setup" % KALOUPILE_BIN,
    deps = [
        "cluster/prerequisites",
        "kind-config.yml",
    ],
    resource_deps = ["build-kaloupile"],
    labels = ["cluster"]
)

local_resource(
    name = "dependencies",
    cmd = "%s dependencies" % KALOUPILE_BIN,
    deps = [
        "cluster/dependencies",
        "config.yml",
        "pkg/dependencies",
    ],
    resource_deps = ["setup", "build-kaloupile"],
    labels = ["cluster"]
)

local_resource(
    name = "routes",
    cmd = "%s routes" % KALOUPILE_BIN,
    deps = [
        "cluster/routes",
        "config.yml",
        "pkg/routes",
    ],
    resource_deps = ["dependencies", "build-kaloupile"],
    labels = ["cluster"]
)

local_resource(
    name = "sync",
    cmd = "sh -c \"until %s sync; do echo [tilt] sync failed, retrying in %s seconds; sleep %s; done\"" % (KALOUPILE_BIN, SYNC_RETRY_SECONDS, SYNC_RETRY_SECONDS),
    deps = [
        "config.yml",
        "pkg/postgresql",
    ],
    resource_deps = ["routes", "build-kaloupile"],
    labels = ["cluster"]
)

config = read_yaml("config.yml")
services = config.get("services", []) if config else []

for service in services:
    service_path = service.get("path") if service else None
    if not service_path:
        continue
    
    service_name = service.get("name", "<unknown>") if service else "<unknown>"
    service_tiltfile = os.path.join(service_path, ".dev", "Tiltfile")
    if os.path.exists(service_tiltfile):
        entrypoint = None
        symbols = load_dynamic(service_tiltfile)
        if not symbols or 'entrypoint' not in symbols:
            fail("[tilt] service Tiltfile is missing an entrypoint: %s" % service_name)
        entrypoint = symbols['entrypoint']
        if type(entrypoint) != "function":
            fail("[tilt] service Tiltfile entrypoint is not a function: %s" % service_name)
        entrypoint(config, service)
        entrypoint = None
        print("[tilt] loaded service Tiltfile: %s" % service_name)
    else:
        print("[tilt] skipping missing service Tiltfile: %s" % service_name)
