/*
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
*/

package controller

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	appsv1 "operators/src/ConfigMapSync/api/v1"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	ConfigMapSyncFinalizer = "configmapsync.apps.kapendra.com/finalizer"
	TypeSynced             = "Synced"
	TypeSourceAvailable    = "SourceAvailable"
	TypeReady              = "Ready"
)

// ConfigMapSyncReconciler reconciles a ConfigMapSync object
type ConfigMapSyncReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=apps.kapendra.com,resources=configmapsyncs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.kapendra.com,resources=configmapsyncs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps.kapendra.com,resources=configmapsyncs/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ConfigMapSync object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// The controller follows a one-way sync pattern: source -> destination
// If the source ConfigMap changes, it will be reflected in the destination
func (r *ConfigMapSyncReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	// Initialize logger for this reconciliation run
	logger := log.FromContext(ctx)

	// Step 1: Fetch the ConfigMapSync resource that triggered this reconciliation
	configMapSync := &appsv1.ConfigMapSync{}
	if err := r.Get(ctx, req.NamespacedName, configMapSync); err != nil {
		// Resource might have been deleted, ignore error
		logger.Error(err, "Failed to fetch ConfigMapSync resource")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Check if someone wants to delete this ConfigMapSync
	if configMapSync.DeletionTimestamp != nil {
		logger.Info("ConfigMapSync is being deleted, starting cleanup")
		destinationKey := types.NamespacedName{
			Name:      configMapSync.Spec.ConfigMapName,
			Namespace: configMapSync.Spec.DestinationNamespace,
		}
		destinationConfigMap := &corev1.ConfigMap{}
		err := r.Get(ctx, destinationKey, destinationConfigMap)
		if err != nil {
			if apierrors.IsNotFound(err) {
				logger.Info("Destination ConfigMap not found, skipping cleanup")
			}
		}
		if err == nil {
			logger.Info("Destination ConfigMap found and deleting it")
			err := r.Delete(ctx, destinationConfigMap)
			if err != nil {
				logger.Error(err, "Failed to delete destination ConfigMap")
				return ctrl.Result{}, err
			}
			logger.Info("Destination ConfigMap deleted successfully")
		}

		// Remove finalizer from ConfigMapSync
		logger.Info("Removing configmapsync finalizer")
		controllerutil.RemoveFinalizer(configMapSync, ConfigMapSyncFinalizer)
		err = r.Update(ctx, configMapSync)
		if err != nil {
			logger.Error(err, "Failed to remove finalizer from ConfigMapSync")
			return ctrl.Result{}, err
		}
		logger.Info("ConfigMapSync finalizer removed successfully")
		return ctrl.Result{}, nil
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(configMapSync, ConfigMapSyncFinalizer) {
		logger.Info("Adding Finalizer to ConfigMapSync")
		controllerutil.AddFinalizer(configMapSync, ConfigMapSyncFinalizer)
		err := r.Update(ctx, configMapSync)
		if err != nil {
			logger.Error(err, "Failed to add finalizer")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// Log the sync operation details for observability
	logger.Info("Processing ConfigMapSync",
		"sourceNameSpace", configMapSync.Spec.SourceNamespace,
		"destinationNameSpace", configMapSync.Spec.DestinationNamespace,
		"configMapName", configMapSync.Spec.ConfigMapName,
	)

	// Step 2: Fetch the source ConfigMap from the source namespace
	sourceConfigMap := &corev1.ConfigMap{}
	sourceKey := types.NamespacedName{
		Name:      configMapSync.Spec.ConfigMapName,
		Namespace: configMapSync.Spec.SourceNamespace,
	}

	err := r.Get(ctx, sourceKey, sourceConfigMap)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Source ConfigMap not found, skipping sync", "sourceKey", sourceKey)
			// Update status to show source not found
			configMapSync.Status.SyncStatus = "Failed"
			configMapSync.Status.Message = "Source ConfigMap not found"
			configMapSync.Status.SourceExists = false
			configMapSync.Status.DestinationExists = false
			configMapSync.Status.LastSyncTime = time.Now().Format(time.RFC3339)
			err = r.Status().Update(ctx, configMapSync)
			if err != nil {
				logger.Error(err, "Failed to update ConfigMapSync status")
			}
			return ctrl.Result{RequeueAfter: time.Minute * 5}, nil
		}

		configMapSync.Status.RetryCount++
		backoffDelay := r.calculateBackoffDuration(configMapSync.Status.RetryCount, time.Second*30)
		logger.Error(err, "Failed to fetch source ConfigMap, retrying with backoff",
			"sourceKey", sourceKey,
			"retryCount", configMapSync.Status.RetryCount,
			"retryAfter", backoffDelay,
		)

		r.setCondition(configMapSync, TypeSynced, metav1.ConditionFalse, "SyncFailed", "Failed to fetch source ConfigMap")
		r.setCondition(configMapSync, TypeSourceAvailable, metav1.ConditionFalse, "FetchError", "Error accessing source ConfigMap")
		r.setCondition(configMapSync, TypeReady, metav1.ConditionFalse, "NotReady", "Source ConfigMap fetch failed")
		configMapSync.Status.SyncStatus = "Failed"
		configMapSync.Status.Message = "Failed to fetch source ConfigMap"
		configMapSync.Status.SourceExists = false
		configMapSync.Status.DestinationExists = false
		configMapSync.Status.LastSyncTime = time.Now().Format(time.RFC3339)
		err = r.Status().Update(ctx, configMapSync)
		if err != nil {
			logger.Error(err, "Failed to update ConfigMapSync status")
		}
		return ctrl.Result{RequeueAfter: backoffDelay}, nil
	}

	logger.Info("Source ConfigMap fetched successfully", "sourceKey", sourceKey, "dataKeys", len(sourceConfigMap.Data))

	// Step 3: Prepare the destination ConfigMap structure with source data
	destinationConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapSync.Spec.ConfigMapName,
			Namespace: configMapSync.Spec.DestinationNamespace,
			// TODO: Add labels/annotations to track the sync relationship
		},
		Data: sourceConfigMap.Data, // Copy all data from source
	}

	if destinationConfigMap.Labels == nil {
		destinationConfigMap.Labels = make(map[string]string)
	}
	destinationConfigMap.Labels["configmapsync.apps.kapendra.com/sync-name"] = configMapSync.Name
	destinationConfigMap.Labels["configmapsync.apps.kapendra.com/sync-namespace"] = configMapSync.Namespace
	destinationConfigMap.Labels["configmapsync.apps.kapendra.com/managed-by"] = "configmapsync-controller"
	sourceHash := r.calculateSourceHash(sourceConfigMap.Data)
	destinationConfigMap.Annotations["configmapsync.apps.kapendra.com/source-hash"] = sourceHash
	destinationConfigMap.Annotations["configmapsync.apps.kapendra.com/last-sync"] = time.Now().Format(time.RFC3339)

	// Step 4: Check if destination ConfigMap already exists
	destinationKey := types.NamespacedName{
		Name:      configMapSync.Spec.ConfigMapName,
		Namespace: configMapSync.Spec.DestinationNamespace,
	}

	existingConfigMap := &corev1.ConfigMap{}
	err = r.Get(ctx, destinationKey, existingConfigMap)

	if err != nil {
		if apierrors.IsNotFound(err) {
			// Case 1: Destination ConfigMap doesn't exist - create it
			logger.Info("Destination ConfigMap not found, creating new one", "destinationKey", destinationKey)
			err = r.Create(ctx, destinationConfigMap)
			if err != nil {
				configMapSync.Status.RetryCount++
				backoffDelay := r.calculateBackoffDuration(configMapSync.Status.RetryCount, time.Minute*1)
				logger.Error(err, "Failed to create destination ConfigMap, retrying with backoff",
					"destinationKey", destinationKey,
					"retryCount", configMapSync.Status.RetryCount,
					"retryAfter", backoffDelay)
				return ctrl.Result{RequeueAfter: backoffDelay}, nil
			}
			logger.Info("Destination ConfigMap created successfully", "destinationKey", destinationKey)
		} else {
			// Unexpected error occurred while fetching destination ConfigMap
			logger.Error(err, "Failed to fetch destination ConfigMap", "destinationKey", destinationKey)
			return ctrl.Result{}, err
		}
	} else {
		// Case 2: Destination ConfigMap exists - update it with source data
		logger.Info("Destination ConfigMap found, updating with source data", "destinationKey", destinationKey)

		// Preserve existing ObjectMeta but update Data section
		existingConfigMap.Data = sourceConfigMap.Data

		err = r.Update(ctx, existingConfigMap)
		if err != nil {
			configMapSync.Status.RetryCount++
			backoffDelay := r.calculateBackoffDuration(configMapSync.Status.RetryCount, time.Minute*1)
			logger.Error(err, "Failed to update destination ConfigMap, retrying with backoff",
				"destinationKey", destinationKey,
				"retryCount", configMapSync.Status.RetryCount,
				"retryAfter", backoffDelay)
			return ctrl.Result{RequeueAfter: backoffDelay}, nil
		}
		logger.Info("Destination ConfigMap updated successfully", "destinationKey", destinationKey)
	}

	configMapSync.Status.RetryCount = 0 // Update status after successful sync
	configMapSync.Status.LastSyncTime = time.Now().Format(time.RFC3339)
	r.setCondition(configMapSync, TypeSynced, metav1.ConditionTrue, "SyncSucceeded", "ConfigMap synced successfully")
	r.setCondition(configMapSync, TypeSourceAvailable, metav1.ConditionTrue, "SourceFound", "Source ConfigMap exists and accessible")
	r.setCondition(configMapSync, TypeReady, metav1.ConditionTrue, "AllComponentsReady", "All sync components are functioning properly")
	configMapSync.Status.SyncStatus = "Success"
	configMapSync.Status.Message = "ConfigMap synced successfully"
	configMapSync.Status.SourceExists = true
	configMapSync.Status.DestinationExists = true

	err = r.Status().Update(ctx, configMapSync)
	if err != nil {
		logger.Error(err, "Failed to update ConfigMapSync status")
		// Don't return error - sync succeeded even if status update failed
	}
	// Sync operation completed successfully
	logger.Info("ConfigMap sync completed successfully",
		"sourceKey", sourceKey,
		"destinationKey", destinationKey,
	)

	return ctrl.Result{}, nil
}

func (r *ConfigMapSyncReconciler) setCondition(configMapSync *appsv1.ConfigMapSync, conditionType string, status metav1.ConditionStatus, reason string, message string) {
	condition := metav1.Condition{
		Type:               conditionType,
		Status:             status,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            message,
	}
	meta.SetStatusCondition(&configMapSync.Status.Conditions, condition)
}

func (r *ConfigMapSyncReconciler) calculateBackoffDuration(retryCount int, baseDelay time.Duration) time.Duration {
	if retryCount == 0 {
		return baseDelay
	}
	backoff := baseDelay * time.Duration(1<<uint(retryCount))
	maxDelay := time.Minute * 10

	if backoff > maxDelay {
		return maxDelay
	}
	return backoff
}

func (r *ConfigMapSyncReconciler) calculateSourceHash(sourceData map[string]string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%v", sourceData))))
}

// SetupWithManager sets up the controller with the Manager.
// This configures the controller to watch ConfigMapSync resources
// and triggers reconciliation when they change.
func (r *ConfigMapSyncReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.ConfigMapSync{}). // Watch ConfigMapSync resources
		Named("configmapsync").       // Give the controller a name
		Complete(r)
}
