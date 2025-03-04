package providers

import (
	"context"

	"github.com/unmeshjoshi/mini-k8s-manager/pkg/api/v1alpha1"
)

// MockProvider implements Provider interface for testing
type MockProvider struct {
	CreateClusterFunc    func(ctx context.Context, cluster *v1alpha1.Cluster) error
	DeleteClusterFunc    func(ctx context.Context, cluster *v1alpha1.Cluster) error
	GetClusterStatusFunc func(ctx context.Context, cluster *v1alpha1.Cluster) (*v1alpha1.ClusterStatus, error)
	UpdateClusterFunc    func(ctx context.Context, cluster *v1alpha1.Cluster) error
}

func (m *MockProvider) CreateCluster(ctx context.Context, cluster *v1alpha1.Cluster) error {
	if m.CreateClusterFunc != nil {
		return m.CreateClusterFunc(ctx, cluster)
	}
	return nil
}

func (m *MockProvider) DeleteCluster(ctx context.Context, cluster *v1alpha1.Cluster) error {
	if m.DeleteClusterFunc != nil {
		return m.DeleteClusterFunc(ctx, cluster)
	}
	return nil
}

func (m *MockProvider) GetClusterStatus(ctx context.Context, cluster *v1alpha1.Cluster) (*v1alpha1.ClusterStatus, error) {
	if m.GetClusterStatusFunc != nil {
		return m.GetClusterStatusFunc(ctx, cluster)
	}
	return &v1alpha1.ClusterStatus{
		Phase:             v1alpha1.ClusterPhaseRunning,
		ControlPlaneReady: true,
		WorkersReady:      cluster.Spec.Workers.Count,
	}, nil
}

func (m *MockProvider) UpdateCluster(ctx context.Context, cluster *v1alpha1.Cluster) error {
	if m.UpdateClusterFunc != nil {
		return m.UpdateClusterFunc(ctx, cluster)
	}
	return nil
}
