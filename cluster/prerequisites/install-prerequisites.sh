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

echo "📦 Installing cert-manager..."
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.19.2/cert-manager.yaml
echo "✅ cert-manager installed"

echo "📦 Installing Infomaniak cert-manager webhook..."
kubectl apply -f https://github.com/infomaniak/cert-manager-webhook-infomaniak/releases/download/v0.2.0/rendered-manifest.yaml
echo "✅ Infomaniak cert-manager webhook installed"

# Install kustomization
kubectl apply -k "$SCRIPT_DIR/"

echo "📦 Installing SeaweedFS S3"
helm repo add seaweedfs https://seaweedfs.github.io/seaweedfs/helm
helm upgrade --install seaweedfs seaweedfs/seaweedfs \
  --namespace seaweedfs \
  --create-namespace \
  --set master.replicaCount=1 \
  --set volume.replicaCount=1 \
  --set filer.replicaCount=1 \
  --set filer.s3.enabled=true \
  --set filer.s3.enableAuth=true \
  --set filer.s3.existingConfigSecret=s3-config \
  --set filer.s3.createBuckets[0].name=dev \
  --set filer.s3.createBuckets[0].anonymousRead=true
#   --set s3.enabled=true \
#   --set s3.createBuckets[0].name=dev \
#   --set s3.createBuckets[0].anonymousRead=true \
#   --set s3.existingConfigSecret=s3-config \
  
echo "✅ SeaweedFS S3 installed"

echo "📦 Installing AWS Mountpoint S3 CSI Driver"
helm repo add aws-mountpoint-s3-csi-driver https://awslabs.github.io/mountpoint-s3-csi-driver
helm upgrade --install aws-mountpoint-s3-csi-driver \
    --namespace kube-system \
    aws-mountpoint-s3-csi-driver/aws-mountpoint-s3-csi-driver
echo "✅ AWS Mountpoint S3 CSI Driver installed"
