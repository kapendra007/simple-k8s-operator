# Kubernetes Operators Collection

A collection of production-ready Kubernetes operators built with Kubebuilder and Go.

## Available Operators

### 🔄 ConfigMapSync Operator
**Location:** `operators/src/ConfigMapSync/`

A Kubernetes operator that synchronizes ConfigMaps across namespaces with intelligent conflict resolution and retry mechanisms.

**Features:**
- Cross-namespace ConfigMap synchronization
- Smart conflict resolution with source-always-wins strategy
- Exponential backoff retry logic
- Comprehensive status tracking and conditions
- Finalizer-based cleanup
- Production-ready with full RBAC

**Quick Start:**
```bash
cd operators/src/ConfigMapSync
# One-command installation
curl -s https://raw.githubusercontent.com/kapendra007/k8s-operator/main/operators/src/ConfigMapSync/install.sh | bash
```

**Documentation:** [ConfigMapSync README](operators/src/ConfigMapSync/README.md)

## Getting Started

Each operator in this collection has its own directory under `operators/src/` with complete documentation, installation instructions, and examples.

### Prerequisites
- Kubernetes cluster (v1.19+)
- kubectl configured
- Go 1.24+ (for development)
- Docker (for building images)

### Development Structure
```
operators/src/
├── ConfigMapSync/          # ConfigMap synchronization operator
└── [future-operators]/     # Additional operators will be added here
```

## Contributing

1. Fork the repository
2. Create a feature branch for your operator
3. Follow the existing operator structure
4. Add comprehensive documentation
5. Submit a pull request

Each operator should follow the established patterns for consistency and maintainability.