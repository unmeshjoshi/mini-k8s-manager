package cluster

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/unmeshjoshi/mini-k8s-manager/pkg/api/v1alpha1"
	"testing"
)

// TestCreateCluster tests the CreateCluster function.
func TestCreateCluster(t *testing.T) {
	cluster := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			KubernetesVersion: v1alpha1.TestKubernetesVersion,
			ControlPlane:      v1alpha1.ControlPlaneConfig{Count: 1},
			Workers:           v1alpha1.WorkerConfig{Count: 1},
		},
	}

	// Call CreateCluster
	err := CreateCluster(context.Background(), cluster)

	// Assert no error
	assert.NoError(t, err)
	// Assert cluster status is updated
	assert.Equal(t, v1alpha1.ClusterPhaseRunning, cluster.Status.Phase)
	assert.True(t, cluster.Status.ControlPlaneReady)
	assert.Equal(t, int32(1), cluster.Status.WorkersReady)
}
