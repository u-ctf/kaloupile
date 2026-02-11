#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "📦 Installing kGateway..."

# Install kGateway CRDs
helm upgrade -i kgateway-crds oci://cr.kgateway.dev/kgateway-dev/charts/kgateway-crds \
    --create-namespace \
    --namespace kgateway-system \
    --version v2.2.0 

# Install kGateway controller
helm upgrade -i kgateway oci://cr.kgateway.dev/kgateway-dev/charts/kgateway \
    --namespace kgateway-system \
    --version v2.2.0 \
    --set controller.image.pullPolicy=Always \
    --set controller.extraEnv.KGW_ENABLE_GATEWAY_API_EXPERIMENTAL_FEATURES="true"

echo "✅ kGateway installed"

echo "📦 Installing Gateway API CRDs..."

# Install Gateway API CRDs (standard channel with experimental features for ListenerSet)
kubectl apply --server-side -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.4.1/experimental-install.yaml

echo "✅ Gateway API CRDs installed"

# Install kustomization
kubectl apply -k "$SCRIPT_DIR/"
