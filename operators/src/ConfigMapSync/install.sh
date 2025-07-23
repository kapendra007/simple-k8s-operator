#!/bin/bash

# ConfigMapSync Operator Installation Script
# Usage: curl -s https://raw.githubusercontent.com/kapendra007/k8s-operator/main/operators/src/ConfigMapSync/install.sh | bash

set -e

echo "🚀 Installing ConfigMapSync Operator..."

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl is not installed. Please install kubectl first."
    exit 1
fi

# Check if cluster is accessible
if ! kubectl cluster-info &> /dev/null; then
    echo "❌ Cannot connect to Kubernetes cluster. Please check your kubeconfig."
    exit 1
fi

echo "✅ Kubernetes cluster connection verified"

# Create temporary directory for manifests
TEMP_DIR=$(mktemp -d)
cd "$TEMP_DIR"

echo "📦 Downloading manifests..."

# Download CRD
curl -s -o configmapsync-crd.yaml https://raw.githubusercontent.com/kapendra007/k8s-operator/main/operators/src/ConfigMapSync/config/crd/bases/apps.kapendra.com_configmapsyncs.yaml

# Create RBAC manifest
cat <<EOF > configmapsync-rbac.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: configmapsync-system
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: configmapsync-controller-manager
  namespace: configmapsync-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: configmapsync-manager-role
rules:
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["create", "delete", "get", "list", "patch", "update", "watch"]
- apiGroups: ["apps.kapendra.com"]
  resources: ["configmapsyncs"]
  verbs: ["create", "delete", "get", "list", "patch", "update", "watch"]
- apiGroups: ["apps.kapendra.com"]
  resources: ["configmapsyncs/finalizers"]
  verbs: ["update"]
- apiGroups: ["apps.kapendra.com"]
  resources: ["configmapsyncs/status"]
  verbs: ["get", "patch", "update"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: configmapsync-manager-rolebinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: configmapsync-manager-role
subjects:
- kind: ServiceAccount
  name: configmapsync-controller-manager
  namespace: configmapsync-system
EOF

# Create deployment manifest
cat <<EOF > configmapsync-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: configmapsync-controller-manager
  namespace: configmapsync-system
  labels:
    app: configmapsync-controller
spec:
  replicas: 1
  selector:
    matchLabels:
      control-plane: controller-manager
  template:
    metadata:
      labels:
        control-plane: controller-manager
    spec:
      serviceAccountName: configmapsync-controller-manager
      containers:
      - name: manager
        image: docker.io/kapendra007/configmapsync:v1.0.0
        ports:
        - containerPort: 8081
          name: metrics
          protocol: TCP
        resources:
          limits:
            cpu: 500m
            memory: 128Mi
          requests:
            cpu: 10m
            memory: 64Mi
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          runAsNonRoot: true
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8081
          initialDelaySeconds: 15
          periodSeconds: 20
        readinessProbe:
          httpGet:
            path: /readyz
            port: 8081
          initialDelaySeconds: 5
          periodSeconds: 10
EOF

echo "📋 Installing CRDs..."
kubectl apply -f configmapsync-crd.yaml

echo "🔐 Setting up RBAC..."
kubectl apply -f configmapsync-rbac.yaml

echo "🚀 Deploying operator..."
kubectl apply -f configmapsync-deployment.yaml

echo "⏳ Waiting for operator to be ready..."
kubectl wait --for=condition=Available deployment/configmapsync-controller-manager -n configmapsync-system --timeout=300s

# Cleanup
cd /
rm -rf "$TEMP_DIR"

echo ""
echo "✅ ConfigMapSync Operator installed successfully!"
echo ""
echo "🎯 Quick Test:"
echo "1. Create a source ConfigMap:"
echo "   kubectl create configmap test-config --from-literal=key=value"
echo ""
echo "2. Create a ConfigMapSync:"
echo "   kubectl apply -f - <<EOF"
echo "   apiVersion: apps.kapendra.com/v1"
echo "   kind: ConfigMapSync"
echo "   metadata:"
echo "     name: test-sync"
echo "   spec:"
echo "     sourceNamespace: default"
echo "     destinationNamespace: kube-system"
echo "     configMapName: test-config"
echo "   EOF"
echo ""
echo "3. Verify sync:"
echo "   kubectl get configmap test-config -n kube-system"
echo "   kubectl get configmapsync test-sync -o yaml"
echo ""
echo "📖 Documentation: https://github.com/kapendra007/k8s-operator"
echo "🐛 Issues: https://github.com/kapendra007/k8s-operator/issues"
echo ""
echo "Happy ConfigMap syncing! 🎉"