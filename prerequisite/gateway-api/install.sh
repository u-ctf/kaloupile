#!/bin/bash
set -e

echo "📦 Installing Gateway API CRDs..."

# Install Gateway API CRDs (standard channel with experimental features for ListenerSet)
kubectl apply --server-side -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.4.1/experimental-install.yaml

echo "✅ Gateway API CRDs installed"
