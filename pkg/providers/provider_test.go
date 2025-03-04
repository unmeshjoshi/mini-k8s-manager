package providers

import (
	"context"
	"testing"

	"github.com/unmeshjoshi/mini-k8s-manager/pkg/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)



func TestProviderInterface(t *testing.T) {
	ctx := context.Background()

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

	// Test 1: Test interface method signatures
	var provider Provider = &MockProvider{} // Verify MockProvider implements Provider interface
	// Test basic provider operations
	if err := provider.CreateCluster(ctx, cluster); err != nil {
		t.Errorf("Basic provider operation failed: %v", err)
	}

	// Test 2: Test error conditions
	errProvider := &MockProvider{
		CreateClusterFunc: func(ctx context.Context, c *v1alpha1.Cluster) error {
			return ErrClusterExists
		},
		DeleteClusterFunc: func(ctx context.Context, c *v1alpha1.Cluster) error {
			return ErrClusterNotFound
		},
		GetClusterStatusFunc: func(ctx context.Context, c *v1alpha1.Cluster) (*v1alpha1.ClusterStatus, error) {
			return nil, ErrClusterNotFound
		},
		UpdateClusterFunc: func(ctx context.Context, c *v1alpha1.Cluster) error {
			return ErrClusterNotFound
		},
	}

	// Test create cluster with error
	err := errProvider.CreateCluster(ctx, cluster)
	if err != ErrClusterExists {
		t.Errorf("Expected ErrClusterExists, got %v", err)
	}

	// Test delete cluster with error
	err = errProvider.DeleteCluster(ctx, cluster)
	if err != ErrClusterNotFound {
		t.Errorf("Expected ErrClusterNotFound, got %v", err)
	}

	// Test get cluster status with error
	_, err = errProvider.GetClusterStatus(ctx, cluster)
	if err != ErrClusterNotFound {
		t.Errorf("Expected ErrClusterNotFound, got %v", err)
	}

	// Test update cluster with error
	err = errProvider.UpdateCluster(ctx, cluster)
	if err != ErrClusterNotFound {
		t.Errorf("Expected ErrClusterNotFound, got %v", err)
	}

	// Test 3: Test successful operations
	successProvider := &MockProvider{
		CreateClusterFunc: func(ctx context.Context, c *v1alpha1.Cluster) error {
			return nil
		},
		DeleteClusterFunc: func(ctx context.Context, c *v1alpha1.Cluster) error {
			return nil
		},
		GetClusterStatusFunc: func(ctx context.Context, c *v1alpha1.Cluster) (*v1alpha1.ClusterStatus, error) {
			return &v1alpha1.ClusterStatus{
				Phase:            "Running",
				ControlPlaneReady: true,
				WorkersReady:     2,
			}, nil
		},
		UpdateClusterFunc: func(ctx context.Context, c *v1alpha1.Cluster) error {
			return nil
		},
	}

	// Test successful create
	if err := successProvider.CreateCluster(ctx, cluster); err != nil {
		t.Errorf("Expected successful create, got error: %v", err)
	}

	// Test successful get status
	status, err := successProvider.GetClusterStatus(ctx, cluster)
	if err != nil {
		t.Errorf("Expected successful status get, got error: %v", err)
	}
	if status.Phase != "Running" {
		t.Errorf("Expected Running phase, got %s", status.Phase)
	}
}
