#!/bin/bash
set -e

echo "📦 Installing kGateway..."

# Install kGateway CRDs
helm upgrade -i kgateway-crds oci://cr.kgateway.dev/kgateway-dev/charts/kgateway-crds \
    --create-namespace \
    --namespace kgateway-system \
    --version v2.1.2 

# Install kGateway controller
helm upgrade -i kgateway oci://cr.kgateway.dev/kgateway-dev/charts/kgateway \
    --namespace kgateway-system \
    --version v2.1.2 \
    --set controller.image.pullPolicy=Always \
    --set controller.extraEnv.KGW_ENABLE_GATEWAY_API_EXPERIMENTAL_FEATURES="true"

echo "✅ kGateway installed"
