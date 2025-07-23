# ConfigMapSync Controller Implementation Plan

## Overview
This document outlines the step-by-step implementation plan for the ConfigMapSync Kubernetes operator that synchronizes ConfigMaps between namespaces.

## Current State Analysis

### Existing CRD Structure
- **Source**: `sourceNamespace` + `configMapName`
- **Target**: `destinationNamespace` + `configMapName`
- **Status**: `lastSyncTime` for tracking sync operations

### Current Controller State
- Basic scaffolding in place (`internal/controller/configmapsync_controller.go`)
- CRD types defined (`api/v1/configmapsync_types.go`)
- RBAC permissions for ConfigMapSync resources configured
- Basic Reconcile function prints spec values but no sync logic implemented

## Implementation Strategy

### Core Sync Logic Flow
```
1. Fetch source ConfigMap from sourceNamespace
2. Create/update target ConfigMap in destinationNamespace
3. Handle deletion scenarios (using finalizers)
4. Update status with sync timestamp
5. Set up watches for source ConfigMap changes
```

### Key Design Decisions

#### Sync Strategy Options
- **One-way sync**: Source → Destination (recommended for initial implementation)
- **Bidirectional**: Complex conflict resolution needed
- **Multi-target**: One source → multiple destinations (future enhancement)

#### Conflict Resolution Strategy
- **Source always wins**: Overwrite destination ConfigMap (recommended)
- **Skip sync**: If destination modified externally
- **Merge strategies**: For specific keys (advanced feature)

#### Deletion Handling Options
- **Cascade deletion**: Delete destination when source deleted
- **Orphan mode**: Keep destination when ConfigMapSync deleted
- **Configurable**: Let user choose deletion behavior

## Step-by-Step Implementation Plan

### Step 1: Add ConfigMap RBAC Permissions
**File**: `internal/controller/configmapsync_controller.go`
- Add RBAC marker for ConfigMap resources
- Required verbs: `get;list;watch;create;update;patch;delete`

### Step 2: Implement Core Sync Logic
**File**: `internal/controller/configmapsync_controller.go`
- Fetch source ConfigMap from source namespace
- Create or update destination ConfigMap in target namespace
- Handle ConfigMap not found scenarios
- Add proper error handling and logging

### Step 3: Add Finalizer Handling
**File**: `internal/controller/configmapsync_controller.go`
- Add finalizer to ConfigMapSync resource
- Implement cleanup logic when ConfigMapSync is deleted
- Remove finalizer after successful cleanup

### Step 4: Set Up ConfigMap Watching
**File**: `internal/controller/configmapsync_controller.go`
- Configure controller to watch ConfigMap changes
- Trigger reconciliation when source ConfigMap changes
- Handle cross-namespace watching

### Step 5: Enhance Status Tracking
**File**: `api/v1/configmapsync_types.go` and controller
- Add more status fields (sync status, error messages, etc.)
- Update status after each sync operation
- Add status conditions for better observability

### Step 6: Add Error Handling and Retries
**File**: `internal/controller/configmapsync_controller.go`
- Implement proper error handling
- Add retry logic for transient failures
- Set appropriate requeue intervals

### Step 7: Implement Conflict Detection
**File**: `internal/controller/configmapsync_controller.go`
- Detect if destination ConfigMap was modified externally
- Add annotations/labels to track managed ConfigMaps
- Implement conflict resolution strategy

## RBAC Requirements

### Additional Permissions Needed
```yaml
# ConfigMap permissions
# +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
```

### Current Permissions
- ConfigMapSync CRD: Full CRUD operations
- Leader election: configmaps, leases, events
- Metrics: tokenreviews, subjectaccessreviews

## Testing Strategy

### Unit Tests
- Test sync logic with mock clients
- Test error scenarios and edge cases
- Test finalizer handling

### Integration Tests
- Test with real Kubernetes cluster
- Test cross-namespace synchronization
- Test deletion scenarios

### E2E Tests
- Deploy operator and create ConfigMapSync resources
- Verify ConfigMaps are synchronized correctly
- Test operator restart scenarios

## Future Enhancements

### Phase 2 Features
- Multi-target synchronization (one source → multiple destinations)
- Selective key synchronization
- ConfigMap transformation rules
- Webhooks for validation

### Phase 3 Features
- Bidirectional synchronization
- Conflict resolution policies
- Metrics and monitoring
- Advanced RBAC integration

## Getting Started

1. Follow the implementation steps in order
2. Test each step thoroughly before proceeding
3. Update this document as implementation progresses
4. Add any discovered requirements or issues

## Notes
- Prefer incremental implementation with testing at each step
- Focus on error handling and edge cases
- Keep security and RBAC principles in mind
- Document any assumptions or limitations discovered during implementation