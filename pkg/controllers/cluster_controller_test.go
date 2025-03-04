package controllers

import (
	"context"
	"testing"
	"time"

	"github.com/unmeshjoshi/mini-k8s-manager/pkg/api/v1alpha1"
	"github.com/unmeshjoshi/mini-k8s-manager/pkg/providers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestClusterReconciler(t *testing.T) {
	// Register cluster types
	s := runtime.NewScheme()
	scheme.AddToScheme(s)
	v1alpha1.AddToScheme(s)

	// Create a fake client
	client := fake.NewClientBuilder().WithScheme(s).Build()

	// Create a mock provider
	mockProvider := &providers.MockProvider{}

	// Create the reconciler
	reconciler := &ClusterReconciler{
		Client:   client,
		Scheme:   s,
		Provider: mockProvider,
	}

	// Create a test cluster
	cluster := &v1alpha1.Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cluster.mini-k8s.io/v1alpha1",
			Kind:       "Cluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster",
			Namespace: "default",
		},
		Spec: v1alpha1.ClusterSpec{
			KubernetesVersion: v1alpha1.TestKubernetesVersion,
			ControlPlane: v1alpha1.ControlPlaneConfig{
				Count: 1,
				MachineConfig: v1alpha1.MachineConfig{
					Memory:   "2Gi",
					CPUCount: 2,
				},
			},
			Workers: v1alpha1.WorkerConfig{
				Count: 2,
				MachineConfig: v1alpha1.MachineConfig{
					Memory:   "4Gi",
					CPUCount: 2,
				},
			},
		},
	}

	// Create the cluster in the fake client
	err := client.Create(context.Background(), cluster)
	if err != nil {
		t.Fatalf("Failed to create test cluster: %v", err)
	}

	// Test initial reconciliation
	_, err = reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
		},
	})
	if err != nil {
		t.Fatalf("Failed to reconcile cluster: %v", err)
	}

	// Get the updated cluster
	updatedCluster := &v1alpha1.Cluster{}
	err = client.Get(context.Background(), types.NamespacedName{
		Name:      cluster.Name,
		Namespace: cluster.Namespace,
	}, updatedCluster)
	if err != nil {
		t.Fatalf("Failed to get updated cluster: %v", err)
	}

	// Verify initial phase is set to Pending
	if updatedCluster.Status.Phase != v1alpha1.ClusterPhasePending {
		t.Errorf("Expected cluster phase to be Pending, got %s", updatedCluster.Status.Phase)
	}

	// Verify finalizer is added
	if !containsString(updatedCluster.ObjectMeta.Finalizers, clusterFinalizer) {
		t.Error("Expected finalizer to be added")
	}

	// Test deletion
	now := metav1.Now()
	updatedCluster.ObjectMeta.DeletionTimestamp = &now
	err = client.Update(context.Background(), updatedCluster)
	if err != nil {
		t.Fatalf("Failed to update cluster: %v", err)
	}

	// Reconcile deletion
	_, err = reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
		},
	})
	if err != nil {
		t.Fatalf("Failed to reconcile cluster deletion: %v", err)
	}

	// Verify cluster is deleted
	err = client.Get(context.Background(), types.NamespacedName{
		Name:      cluster.Name,
		Namespace: cluster.Namespace,
	}, updatedCluster)
	if err == nil {
		t.Error("Expected cluster to be deleted")
	}
}

func TestClusterPhaseTransitions(t *testing.T) {
	// Register cluster types
	s := runtime.NewScheme()
	scheme.AddToScheme(s)
	v1alpha1.AddToScheme(s)

	// Create a fake client
	client := fake.NewClientBuilder().WithScheme(s).Build()

	// Create a mock provider
	mockProvider := &providers.MockProvider{}

	// Create the reconciler
	reconciler := &ClusterReconciler{
		Client:   client,
		Scheme:   s,
		Provider: mockProvider,
	}

	// Create a test cluster
	cluster := &v1alpha1.Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cluster.mini-k8s.io/v1alpha1",
			Kind:       "Cluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster",
			Namespace: "default",
		},
		Spec: v1alpha1.ClusterSpec{
			KubernetesVersion: v1alpha1.TestKubernetesVersion,
			ControlPlane: v1alpha1.ControlPlaneConfig{
				Count: 1,
			},
		},
	}

	// Create the cluster
	err := client.Create(context.Background(), cluster)
	if err != nil {
		t.Fatalf("Failed to create test cluster: %v", err)
	}

	// Test phase transitions
	phases := []v1alpha1.ClusterPhase{
		v1alpha1.ClusterPhasePending,
		v1alpha1.ClusterPhaseProvisioning,
		v1alpha1.ClusterPhaseRunning,
	}

	for i, phase := range phases {
		// Reconcile
		_, err = reconciler.Reconcile(context.Background(), ctrl.Request{
			NamespacedName: types.NamespacedName{
				Name:      cluster.Name,
				Namespace: cluster.Namespace,
			},
		})
		if err != nil {
			t.Fatalf("Failed to reconcile cluster at phase %s: %v", phase, err)
		}

		// Get the updated cluster
		updatedCluster := &v1alpha1.Cluster{}
		err = client.Get(context.Background(), types.NamespacedName{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
		}, updatedCluster)
		if err != nil {
			t.Fatalf("Failed to get updated cluster: %v", err)
		}

		// For the first reconciliation, phase should be Pending
		if i == 0 {
			if updatedCluster.Status.Phase != v1alpha1.ClusterPhasePending {
				t.Errorf("Expected phase to be Pending, got %s", updatedCluster.Status.Phase)
			}
			continue
		}

		// Verify phase transition
		if updatedCluster.Status.Phase != phase {
			t.Errorf("Expected phase to be %s, got %s", phase, updatedCluster.Status.Phase)
		}

		// Wait a bit before next reconciliation
		time.Sleep(100 * time.Millisecond)
	}
}
