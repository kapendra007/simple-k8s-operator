# ConfigMapSync Operator

A Kubernetes operator that synchronizes ConfigMaps across namespaces, built with Kubebuilder and Go.

## 🚀 Features

- **Cross-Namespace Sync**: Synchronize ConfigMaps from source to destination namespaces
- **Smart Conflict Resolution**: Source-always-wins strategy with hash-based change detection
- **Exponential Backoff Retries**: Intelligent retry mechanism for transient failures
- **Comprehensive Status Tracking**: Rich status reporting with conditions and retry counts
- **Finalizer-Based Cleanup**: Automatic cleanup of destination ConfigMaps on deletion
- **Production Ready**: Full RBAC, error handling, and observability

## 📋 Prerequisites

- Kubernetes cluster (v1.19+)
- kubectl configured
- Go 1.24+ (for development)
- Docker (for building images)

## 🚀 Quick Start (New Users)

### ⚡ One-Command Installation

**The fastest way to get started:**

```bash
curl -s https://raw.githubusercontent.com/kapendra007/k8s-operator/main/operators/src/ConfigMapSync/install.sh | bash
```

### Option 1: Use Pre-built Image (Recommended)

**Perfect for users who just want to use the operator without building:**

1. **Clone the repository**:
   ```bash
   git clone https://github.com/kapendra007/k8s-operator.git
   cd k8s-operator/operators/src/ConfigMapSync
   ```

2. **Deploy to your cluster**:
   ```bash
   # Install the Custom Resource Definitions
   kubectl apply -f config/crd/bases/

   # Create the operator namespace and RBAC
   kubectl apply -f config/rbac/

   # Deploy the operator with pre-built image
   kubectl apply -f - <<EOF
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: configmapsync-controller-manager
     namespace: configmapsync-system
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
         containers:
         - name: manager
           image: docker.io/kapendra007/configmapsync:v1.0.0
           ports:
           - containerPort: 8081
             name: metrics
           resources:
             limits:
               cpu: 500m
               memory: 128Mi
             requests:
               cpu: 10m
               memory: 64Mi
         serviceAccountName: configmapsync-controller-manager
   EOF
   ```

3. **Verify the installation**:
   ```bash
   kubectl get pods -n configmapsync-system
   kubectl logs -f deployment/configmapsync-controller-manager -n configmapsync-system
   ```

4. **Create your first ConfigMapSync**:
   ```bash
   # First, create a source ConfigMap
   kubectl create configmap my-app-config --from-literal=database.url=postgres://localhost:5432/myapp
   
   # Then sync it to another namespace
   kubectl apply -f - <<EOF
   apiVersion: apps.kapendra.com/v1
   kind: ConfigMapSync
   metadata:
     name: my-first-sync
     namespace: default
   spec:
     sourceNamespace: default
     destinationNamespace: kube-system  
     configMapName: my-app-config
   EOF
   ```

5. **Check the sync worked**:
   ```bash
   kubectl get configmap my-app-config -n kube-system
   kubectl get configmapsync my-first-sync -o yaml
   ```

### Option 2: Deploy from Source

**For developers who want to build and customize:**

1. **Clone the repository**:
   ```bash
   git clone https://github.com/kapendra007/k8s-operator.git
   cd k8s-operator/operators/src/ConfigMapSync
   ```

2. **Install CRDs**:
   ```bash
   make install
   ```

3. **Run the operator locally**:
   ```bash
   make run
   ```

### Option 3: Deploy Your Own Image

**For users who want to build and deploy their own image:**

1. **Build and push image**:
   ```bash
   make docker-build docker-push IMG=<your-registry>/configmapsync:tag
   ```

2. **Deploy**:
   ```bash
   make deploy IMG=<your-registry>/configmapsync:tag
   ```

## 🏭 Production Deployment

### Step 1: Build and Push Container Image

**1. Choose your container registry:**
```bash
# Examples:
# Docker Hub: docker.io/your-username/configmapsync:v1.0.0
# GitHub Container Registry: ghcr.io/your-username/configmapsync:v1.0.0
# Google Container Registry: gcr.io/your-project/configmapsync:v1.0.0
# AWS ECR: your-account.dkr.ecr.region.amazonaws.com/configmapsync:v1.0.0

export IMG=docker.io/your-username/configmapsync:v1.0.0
```

**2. Build and push:**
```bash
# Build the Docker image
make docker-build IMG=$IMG

# Login to your registry
docker login docker.io  # or your registry

# Push the image
make docker-push IMG=$IMG
```

### Step 2: Deploy to Production Cluster

**1. Connect to your production cluster:**
```bash
kubectl config current-context
kubectl cluster-info
```

**2. Install CRDs and deploy:**
```bash
# Install the Custom Resource Definitions
make install

# Deploy the operator
make deploy IMG=$IMG
```

### Step 3: Verify Production Deployment

**1. Check operator status:**
```bash
kubectl get deployment -n configmapsync-system
kubectl get pods -n configmapsync-system
```

**2. View operator logs:**
```bash
kubectl logs -f deployment/configmapsync-controller-manager -n configmapsync-system
```

**3. Test with a sample ConfigMapSync:**
```bash
kubectl apply -f config/samples/
kubectl get configmapsync
```

### Step 4: Production Configuration

**Resource Limits (Recommended):**
```bash
# Edit the deployment to add resource limits
kubectl patch deployment configmapsync-controller-manager -n configmapsync-system -p '
{
  "spec": {
    "template": {
      "spec": {
        "containers": [{
          "name": "manager",
          "resources": {
            "limits": {"cpu": "500m", "memory": "128Mi"},
            "requests": {"cpu": "10m", "memory": "64Mi"}
          }
        }]
      }
    }
  }
}'
```

**High Availability (Optional):**
```bash  
# Scale to multiple replicas
kubectl scale deployment configmapsync-controller-manager --replicas=2 -n configmapsync-system
```

### Step 5: Production Monitoring

**Health Check:**
```bash
# Check if operator is healthy
kubectl get pods -n configmapsync-system
kubectl describe deployment configmapsync-controller-manager -n configmapsync-system
```

**Monitor ConfigMapSync Resources:**
```bash
# List all ConfigMapSync resources across namespaces
kubectl get configmapsync --all-namespaces

# Check specific resource status
kubectl describe configmapsync <name> -n <namespace>
```

## 🚀 Quick Production Deployment Script

Create this script for automated deployment:

```bash
#!/bin/bash
# deploy-prod.sh

set -e

IMAGE_REGISTRY=${1:-"docker.io/your-username"}
IMAGE_TAG=${2:-"v1.0.0"}
IMG="$IMAGE_REGISTRY/configmapsync:$IMAGE_TAG"

echo "🚀 Deploying ConfigMapSync Operator to Production"
echo "📦 Image: $IMG"

# Build and push
echo "🔨 Building image..."
make docker-build IMG=$IMG

echo "📤 Pushing image..."
make docker-push IMG=$IMG

# Deploy
echo "⚙️  Installing CRDs..."
make install

echo "🚀 Deploying operator..."
make deploy IMG=$IMG

# Wait for deployment
echo "⏳ Waiting for deployment to be ready..."
kubectl wait --for=condition=Available deployment/configmapsync-controller-manager -n configmapsync-system --timeout=300s

# Verify
echo "✅ Deployment complete!"
echo "📋 Operator Status:"
kubectl get pods -n configmapsync-system
echo ""
echo "🔍 To view logs: kubectl logs -f deployment/configmapsync-controller-manager -n configmapsync-system"
echo "🧪 To test: kubectl apply -f config/samples/"
```

**Usage:**
```bash
chmod +x deploy-prod.sh
./deploy-prod.sh docker.io/your-username v1.0.0
```

## 📖 Usage

### Basic Example

Create a ConfigMapSync resource to sync a ConfigMap from one namespace to another:

```yaml
apiVersion: apps.kapendra.com/v1
kind: ConfigMapSync
metadata:
  name: my-config-sync
  namespace: default
spec:
  sourceNamespace: source-ns
  destinationNamespace: target-ns  
  configMapName: my-config
```

### Complete Example

```yaml
# First, create the source ConfigMap
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: production
data:
  database.url: "postgres://prod-db:5432/myapp"
  app.name: "MyApp"
  config.yaml: |
    port: 8080
    debug: false

---
# Create the ConfigMapSync to replicate it
apiVersion: apps.kapendra.com/v1
kind: ConfigMapSync
metadata:
  name: app-config-sync
  namespace: default
spec:
  sourceNamespace: production
  destinationNamespace: staging
  configMapName: app-config
```

## 🔍 Monitoring and Status

### Check Sync Status

```bash
kubectl get configmapsync my-config-sync -o yaml
```

### Example Status Output

```yaml
status:
  conditions:
  - type: Synced
    status: "True"
    reason: SyncSucceeded
    message: ConfigMap synced successfully
  - type: SourceAvailable  
    status: "True"
    reason: SourceFound
    message: Source ConfigMap exists and accessible
  - type: Ready
    status: "True" 
    reason: AllComponentsReady
    message: All sync components are functioning properly
  lastSyncTime: "2025-01-23T10:30:45Z"
  syncStatus: Success
  message: ConfigMap synced successfully
  sourceExists: true
  destinationExists: true
  retryCount: 0
```

### View Logs

```bash
kubectl logs -f deployment/configmapsync-controller-manager -n configmapsync-system
```

## 🎯 How It Works

### Sync Process

1. **Watch**: Controller watches ConfigMapSync resources
2. **Fetch**: Retrieves source ConfigMap from specified namespace  
3. **Sync**: Creates or updates destination ConfigMap
4. **Track**: Updates hash and timestamp annotations for change detection
5. **Status**: Reports comprehensive status with conditions

### Conflict Resolution

- **Strategy**: Source Always Wins
- **Detection**: SHA256 hash tracking of source ConfigMap data
- **Resolution**: Destination ConfigMap is always overwritten with source data
- **Tracking**: Annotations track sync history and source hash

### Error Handling

- **Exponential Backoff**: Retry delays increase with each failure (30s → 1m → 2m → 4m → 8m → 10m max)
- **Different Intervals**: 
  - Source not found: 5 minutes
  - API errors: 30 seconds  
  - Create/Update failures: 1 minute base
- **Status Updates**: All errors reflected in status conditions

### Cleanup

- **Finalizers**: Prevent deletion until cleanup completes
- **Automatic**: Destination ConfigMaps deleted when ConfigMapSync is removed
- **Safe**: Handles edge cases and concurrent operations

## 🏗️ Architecture

### Custom Resource Definition

The operator defines a `ConfigMapSync` CRD with the following structure:

```yaml
apiVersion: apps.kapendra.com/v1
kind: ConfigMapSync
spec:
  sourceNamespace: string      # Source namespace containing the ConfigMap
  destinationNamespace: string # Target namespace for ConfigMap replication  
  configMapName: string        # Name of the ConfigMap to sync
status:
  lastSyncTime: string         # RFC3339 timestamp of last successful sync
  syncStatus: string           # Current sync status (Success/Failed/InProgress)
  message: string              # Human-readable status message
  sourceExists: boolean        # Whether source ConfigMap exists
  destinationExists: boolean   # Whether destination ConfigMap exists  
  retryCount: integer          # Number of retry attempts for current operation
  conditions: []Condition      # Kubernetes-standard status conditions
```

### Controller Logic

- **Reconciliation Loop**: Event-driven processing of ConfigMapSync resources
- **Cross-Namespace**: Handles permissions and security across namespace boundaries
- **State Management**: Tracks sync state, retry attempts, and error conditions
- **Observability**: Comprehensive logging and status reporting

## 🔒 Security & RBAC

The operator requires the following permissions:

```yaml
# ConfigMapSync CRD permissions
- apiGroups: ["apps.kapendra.com"]
  resources: ["configmapsyncs", "configmapsyncs/status", "configmapsyncs/finalizers"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]

# ConfigMap permissions  
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]

# Leader election and events
- apiGroups: [""]
  resources: ["configmaps", "events"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["coordination.k8s.io"]
  resources: ["leases"]  
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
```

## 🧪 Development

### Setup

1. **Clone and setup**:
   ```bash
   git clone <repo-url>
   cd ConfigMapSync
   go mod tidy
   ```

2. **Generate code**:
   ```bash
   make generate
   make manifests
   ```

3. **Run tests**:
   ```bash
   make test
   ```

4. **Run locally**:
   ```bash
   make install  # Install CRDs
   make run      # Run controller locally
   ```

### Code Structure

```
.
├── api/v1/                    # CRD definitions
│   ├── configmapsync_types.go
│   └── zz_generated.deepcopy.go
├── internal/controller/       # Controller logic  
│   └── configmapsync_controller.go
├── config/                    # Kubernetes manifests
│   ├── crd/bases/
│   ├── rbac/
│   └── manager/
├── cmd/                       # Main application
└── Makefile                   # Build automation
```

### Key Components

- **`ConfigMapSyncReconciler`**: Main controller with reconciliation logic
- **`setCondition()`**: Helper for managing Kubernetes status conditions  
- **`calculateBackoffDuration()`**: Exponential backoff calculation for retries
- **`calculateSourceHash()`**: SHA256-based change detection for ConfigMap data

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go conventions and formatting (`go fmt`, `go vet`)
- Add tests for new functionality
- Update documentation for API changes
- Ensure RBAC permissions are minimal and correct
- Test cross-namespace scenarios thoroughly

## 📝 Advanced Usage

### Monitoring Multiple ConfigMaps

Create multiple ConfigMapSync resources to sync different ConfigMaps:

```yaml
apiVersion: apps.kapendra.com/v1
kind: ConfigMapSync
metadata:
  name: app-config-sync
spec:
  sourceNamespace: production
  destinationNamespace: staging
  configMapName: app-config
---
apiVersion: apps.kapendra.com/v1
kind: ConfigMapSync  
metadata:
  name: db-config-sync
spec:
  sourceNamespace: production
  destinationNamespace: staging
  configMapName: database-config
```

### Cleanup and Uninstall

```bash
# Delete all ConfigMapSync resources
kubectl delete configmapsyncs --all

# Uninstall CRDs
make uninstall

# Remove controller deployment
make undeploy
```

### Troubleshooting

**Common Issues for New Users**:

1. **"No resources found" when getting ConfigMapSync**:
   ```bash
   # Install CRDs first
   kubectl apply -f config/crd/bases/
   
   # Or if using make
   make install
   ```

2. **"Permission denied" errors**:
   ```bash
   # Apply RBAC permissions
   kubectl apply -f config/rbac/
   
   # Check if service account exists
   kubectl get serviceaccount configmapsync-controller-manager -n configmapsync-system
   ```

3. **Operator pod not starting**:
   ```bash
   # Check if namespace exists
   kubectl create namespace configmapsync-system --dry-run=client -o yaml | kubectl apply -f -
   
   # Check image pull issues
   kubectl describe pod -l control-plane=controller-manager -n configmapsync-system
   ```

4. **ConfigMapSync not syncing**:
   ```bash
   # Check if source ConfigMap exists
   kubectl get configmap <configmap-name> -n <source-namespace>
   
   # Check if destination namespace exists
   kubectl get namespace <destination-namespace>
   
   # Check ConfigMapSync status
   kubectl describe configmapsync <name>
   ```

5. **"ImagePullBackOff" error**:
   - The pre-built image is publicly available at `docker.io/kapendra007/configmapsync:v1.0.0`
   - No authentication required
   - Check your cluster's internet connectivity

**Debug Commands**:
```bash
# Check operator logs
kubectl logs -f deployment/configmapsync-controller-manager -n configmapsync-system

# Check if CRDs are installed
kubectl get crd configmapsyncs.apps.kapendra.com

# List all ConfigMapSync resources
kubectl get configmapsync --all-namespaces

# Describe ConfigMapSync resource  
kubectl describe configmapsync <name> -n <namespace>

# Check destination ConfigMap
kubectl get configmap <name> -n <destination-namespace> -o yaml

# Check operator deployment
kubectl get deployment configmapsync-controller-manager -n configmapsync-system
```

**Complete Fresh Installation (If Everything Fails)**:
```bash
# 1. Clean up any existing installation
kubectl delete namespace configmapsync-system --ignore-not-found
kubectl delete crd configmapsyncs.apps.kapendra.com --ignore-not-found

# 2. Fresh installation
git clone https://github.com/kapendra007/k8s-operator.git
cd k8s-operator/operators/src/ConfigMapSync

# 3. Apply everything step by step
kubectl apply -f config/crd/bases/
kubectl create namespace configmapsync-system
kubectl apply -f config/rbac/

# 4. Deploy operator
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: configmapsync-controller-manager
  namespace: configmapsync-system
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
      containers:
      - name: manager
        image: docker.io/kapendra007/configmapsync:v1.0.0
        resources:
          limits:
            cpu: 500m
            memory: 128Mi
          requests:
            cpu: 10m
            memory: 64Mi
      serviceAccountName: configmapsync-controller-manager
EOF

# 5. Wait and verify
kubectl wait --for=condition=Available deployment/configmapsync-controller-manager -n configmapsync-system --timeout=300s
kubectl get pods -n configmapsync-system
```

## 📊 Metrics and Observability

The operator provides comprehensive observability through:

- **Structured Logging**: JSON-formatted logs with contextual information
- **Status Conditions**: Kubernetes-standard condition reporting
- **Retry Tracking**: Visible retry counts and backoff strategies
- **Change Detection**: Hash-based tracking of source ConfigMap modifications

## 🔄 Operational Patterns

### Blue-Green Deployments
Use ConfigMapSync to replicate configuration from production to staging environments for testing.

### Multi-Tenant Configuration
Sync common configuration from a central namespace to multiple tenant namespaces.

### Configuration Promotion
Promote tested configurations from development → staging → production namespaces.

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/your-username/configmapsync/issues)
- **Documentation**: This README and inline code comments
- **Community**: Follow Kubernetes operator best practices

## 🙏 Acknowledgments

- Built with [Kubebuilder](https://kubebuilder.io/)
- Inspired by Kubernetes community best practices
- Thanks to the [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime) project

## 📄 License

Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

---

**Happy ConfigMap syncing!** 🎉