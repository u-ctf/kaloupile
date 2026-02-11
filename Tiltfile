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
)

local_resource(
    name = "setup",
    cmd = "%s setup" % KALOUPILE_BIN,
    deps = [
        "cluster/prerequisites",
        "kind-config.yml",
    ],
    resource_deps = ["build-kaloupile"],
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
)

local_resource(
    name = "sync",
    cmd = "sh -c \"until %s sync; do echo [tilt] sync failed, retrying in %s seconds; sleep %s; done\"" % (KALOUPILE_BIN, SYNC_RETRY_SECONDS, SYNC_RETRY_SECONDS),
    deps = [
        "config.yml",
        "pkg/postgresql",
    ],
    resource_deps = ["routes", "build-kaloupile"],
)
