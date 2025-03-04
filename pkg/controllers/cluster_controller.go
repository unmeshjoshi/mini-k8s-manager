package controllers

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	clusterv1alpha1 "github.com/unmeshjoshi/mini-k8s-manager/pkg/api/v1alpha1"
	"github.com/unmeshjoshi/mini-k8s-manager/pkg/providers"
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Provider providers.Provider
}

// +kubebuilder:rbac:groups=cluster.mini-k8s.io,resources=clusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster.mini-k8s.io,resources=clusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cluster.mini-k8s.io,resources=clusters/finalizers,verbs=update

func (r *ClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Reconciling Cluster", "name", req.Name, "namespace", req.Namespace)

	// Get the Cluster resource
	var cluster clusterv1alpha1.Cluster
	if err := r.Get(ctx, req.NamespacedName, &cluster); err != nil {
		// Handle deletion
		if client.IgnoreNotFound(err) != nil {
			log.Error(err, "Failed to get Cluster")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// Initialize status if not set
	if cluster.Status.Phase == "" {
		cluster.Status.Phase = clusterv1alpha1.ClusterPhasePending
		if err := r.Status().Update(ctx, &cluster); err != nil {
			log.Error(err, "Failed to update Cluster status")
			return ctrl.Result{}, err
		}
	}

	// Add finalizer if not present
	if !containsString(cluster.ObjectMeta.Finalizers, clusterFinalizer) {
		cluster.ObjectMeta.Finalizers = append(cluster.ObjectMeta.Finalizers, clusterFinalizer)
		if err := r.Update(ctx, &cluster); err != nil {
			log.Error(err, "Failed to add finalizer")
			return ctrl.Result{}, err
		}
	}

	// Handle deletion
	if !cluster.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, &cluster)
	}

	// Handle cluster lifecycle based on current phase
	switch cluster.Status.Phase {
	case clusterv1alpha1.ClusterPhasePending:
		return r.handlePendingPhase(ctx, &cluster)
	case clusterv1alpha1.ClusterPhaseProvisioning:
		return r.handleProvisioningPhase(ctx, &cluster)
	case clusterv1alpha1.ClusterPhaseRunning:
		return r.handleRunningPhase(ctx, &cluster)
	case clusterv1alpha1.ClusterPhaseFailed:
		return r.handleFailedPhase(ctx, &cluster)
	default:
		log.Info("Unknown cluster phase", "phase", cluster.Status.Phase)
		return ctrl.Result{}, nil
	}
}

const clusterFinalizer = "cluster.mini-k8s.io/finalizer"

func (r *ClusterReconciler) handleDeletion(ctx context.Context, cluster *clusterv1alpha1.Cluster) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Handling cluster deletion", "name", cluster.Name)

	if containsString(cluster.ObjectMeta.Finalizers, clusterFinalizer) {
		// Delete the cluster using provider
		if err := r.Provider.DeleteCluster(ctx, cluster); err != nil {
			log.Error(err, "Failed to delete cluster")
			return ctrl.Result{}, err
		}

		// Remove finalizer
		cluster.ObjectMeta.Finalizers = removeString(cluster.ObjectMeta.Finalizers, clusterFinalizer)
		if err := r.Update(ctx, cluster); err != nil {
			log.Error(err, "Failed to remove finalizer")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *ClusterReconciler) handlePendingPhase(ctx context.Context, cluster *clusterv1alpha1.Cluster) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Handling pending phase", "name", cluster.Name)

	// Update status to Provisioning
	cluster.Status.Phase = clusterv1alpha1.ClusterPhaseProvisioning
	if err := r.Status().Update(ctx, cluster); err != nil {
		log.Error(err, "Failed to update status to Provisioning")
		return ctrl.Result{}, err
	}

	return ctrl.Result{Requeue: true}, nil
}

func (r *ClusterReconciler) handleProvisioningPhase(ctx context.Context, cluster *clusterv1alpha1.Cluster) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Handling provisioning phase", "name", cluster.Name)

	// Create the cluster using provider
	if err := r.Provider.CreateCluster(ctx, cluster); err != nil {
		log.Error(err, "Failed to create cluster")
		cluster.Status.Phase = clusterv1alpha1.ClusterPhaseFailed
		cluster.Status.Message = fmt.Sprintf("Failed to create cluster: %v", err)
		if updateErr := r.Status().Update(ctx, cluster); updateErr != nil {
			log.Error(updateErr, "Failed to update status")
			return ctrl.Result{}, updateErr
		}
		return ctrl.Result{}, err
	}

	// Update status to Running
	cluster.Status.Phase = clusterv1alpha1.ClusterPhaseRunning
	if err := r.Status().Update(ctx, cluster); err != nil {
		log.Error(err, "Failed to update status to Running")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ClusterReconciler) handleRunningPhase(ctx context.Context, cluster *clusterv1alpha1.Cluster) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Handling running phase", "name", cluster.Name)

	// Get cluster status from provider
	status, err := r.Provider.GetClusterStatus(ctx, cluster)
	if err != nil {
		log.Error(err, "Failed to get cluster status")
		return ctrl.Result{}, err
	}

	// Update cluster status
	cluster.Status = *status
	if err := r.Status().Update(ctx, cluster); err != nil {
		log.Error(err, "Failed to update cluster status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: 30}, nil
}

func (r *ClusterReconciler) handleFailedPhase(ctx context.Context, cluster *clusterv1alpha1.Cluster) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Handling failed phase", "name", cluster.Name)

	// For now, just requeue after some time
	return ctrl.Result{RequeueAfter: 300}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clusterv1alpha1.Cluster{}).
		Complete(r)
}

// Helper functions
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) []string {
	var result []string
	for _, item := range slice {
		if item != s {
			result = append(result, item)
		}
	}
	return result
}
