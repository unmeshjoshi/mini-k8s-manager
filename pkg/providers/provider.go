package providers

import (
	"context"
	"errors"

	"github.com/unmeshjoshi/mini-k8s-manager/pkg/api/v1alpha1"
)

// Common errors that providers may return
var (
	ErrClusterExists    = errors.New("cluster already exists")
	ErrClusterNotFound  = errors.New("cluster not found")
	ErrInvalidConfig    = errors.New("invalid cluster configuration")
	ErrProviderNotReady = errors.New("provider not ready")
)

// Provider defines the interface that all infrastructure providers must implement
type Provider interface {
	// CreateCluster creates a new Kubernetes cluster
	// Returns ErrClusterExists if a cluster with the same name already exists
	// Returns ErrInvalidConfig if the cluster configuration is invalid
	CreateCluster(ctx context.Context, cluster *v1alpha1.Cluster) error

	// DeleteCluster deletes an existing Kubernetes cluster
	// Returns ErrClusterNotFound if the cluster does not exist
	DeleteCluster(ctx context.Context, cluster *v1alpha1.Cluster) error

	// GetClusterStatus retrieves the current status of a cluster
	// Returns ErrClusterNotFound if the cluster does not exist
	GetClusterStatus(ctx context.Context, cluster *v1alpha1.Cluster) (*v1alpha1.ClusterStatus, error)

	// UpdateCluster updates an existing cluster's configuration
	// Returns ErrClusterNotFound if the cluster does not exist
	// Returns ErrInvalidConfig if the update configuration is invalid
	UpdateCluster(ctx context.Context, cluster *v1alpha1.Cluster) error
}

// BaseProvider provides common functionality for providers
type BaseProvider struct {
	Name string
}

// ValidateClusterSpec validates common cluster configuration
func (b *BaseProvider) ValidateClusterSpec(spec *v1alpha1.ClusterSpec) error {
	if spec == nil {
		return ErrInvalidConfig
	}

	if spec.KubernetesVersion == "" {
		return errors.New("kubernetes version is required")
	}

	if spec.ControlPlane.Count < 1 {
		return errors.New("at least one control plane node is required")
	}

	if spec.Workers.Count < 0 {
		return errors.New("worker count cannot be negative")
	}

	return nil
}
