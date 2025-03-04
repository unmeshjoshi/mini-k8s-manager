package providers

import (
	"context"
	"testing"
	"time"

	"github.com/unmeshjoshi/mini-k8s-manager/pkg/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDockerProvider(t *testing.T) {
	ctx := context.Background()
	provider, err := NewDockerProvider(&v1alpha1.DockerProviderConfig{
		Spec: v1alpha1.DockerProviderConfigSpec{
			Network: v1alpha1.NetworkConfig{
				CIDR:          "172.20.0.0/16",
				SubnetMask:    24,
				ExposedPorts:  []int32{6443},
				EnableIPv6:    false,
				DNSNameserver: "8.8.8.8",
			},
			ResourceLimits: v1alpha1.ResourceLimitsConfig{
				CPU: v1alpha1.ResourceLimit{
					Default: "2",
					Min:     "1",
					Max:     "4",
				},
				Memory: v1alpha1.ResourceLimit{
					Default: "2Gi",
					Min:     "1Gi",
					Max:     "8Gi",
				},
				Storage: v1alpha1.ResourceLimit{
					Default: "20Gi",
					Min:     "10Gi",
					Max:     "100Gi",
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Failed to create Docker provider: %v", err)
	}

	cluster := &v1alpha1.Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cluster.mini-k8s.io/v1alpha1",
			Kind:       "Cluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-docker-cluster",
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

	// Test 1: Container Lifecycle Management
	t.Run("Container Lifecycle", func(t *testing.T) {
		// Create cluster
		err := provider.CreateCluster(ctx, cluster)
		if err != nil {
			t.Fatalf("Failed to create cluster: %v", err)
		}

		// Verify cluster status
		status, err := provider.GetClusterStatus(ctx, cluster)
		if err != nil {
			t.Fatalf("Failed to get cluster status: %v", err)
		}

		// Wait for cluster to be ready (max 2 minutes)
		deadline := time.Now().Add(2 * time.Minute)
		for status.Phase != "Running" && time.Now().Before(deadline) {
			time.Sleep(5 * time.Second)
			status, err = provider.GetClusterStatus(ctx, cluster)
			if err != nil {
				t.Fatalf("Failed to get cluster status: %v", err)
			}
		}

		if status.Phase != "Running" {
			t.Errorf("Expected cluster phase to be Running, got %s", status.Phase)
		}

		if !status.ControlPlaneReady {
			t.Error("Control plane should be ready")
		}

		if status.WorkersReady != cluster.Spec.Workers.Count {
			t.Errorf("Expected %d workers ready, got %d", cluster.Spec.Workers.Count, status.WorkersReady)
		}

		// Clean up
		err = provider.DeleteCluster(ctx, cluster)
		if err != nil {
			t.Fatalf("Failed to delete cluster: %v", err)
		}
	})

	// Test 2: Network Setup
	t.Run("Network Setup", func(t *testing.T) {
		err := provider.CreateCluster(ctx, cluster)
		if err != nil {
			t.Fatalf("Failed to create cluster: %v", err)
		}

		// Verify network configuration
		networkID := provider.getClusterNetworkName(cluster)
		networkInfo := provider.getNetworkInfo(networkID)
		if networkInfo == nil {
			t.Fatal("Failed to get network info")
		}

		if networkInfo.EndpointsConfig[networkID] == nil {
			t.Error("Network endpoint configuration should exist")
		}

		err = provider.DeleteCluster(ctx, cluster)
		if err != nil {
			t.Fatalf("Failed to delete cluster: %v", err)
		}
	})

	// Test 3: Resource Allocation
	t.Run("Resource Allocation", func(t *testing.T) {
		err := provider.CreateCluster(ctx, cluster)
		if err != nil {
			t.Fatalf("Failed to create cluster: %v", err)
		}

		// Verify resource allocation
		resources := provider.getResourceAllocation(cluster.Spec.ControlPlane.MachineConfig)
		if resources == nil {
			t.Fatal("Failed to get resource allocation")
		}

		// Check control plane resources
		expectedCPUQuota := int64(cluster.Spec.ControlPlane.MachineConfig.CPUCount * 100000)
		if resources.CPUQuota != expectedCPUQuota {
			t.Errorf("Expected CPU quota %d, got %d",
				expectedCPUQuota,
				resources.CPUQuota)
		}

		expectedMemory := provider.parseMemory(cluster.Spec.ControlPlane.MachineConfig.Memory)
		if resources.Memory != expectedMemory {
			t.Errorf("Expected memory %d, got %d",
				expectedMemory,
				resources.Memory)
		}

		err = provider.DeleteCluster(ctx, cluster)
		if err != nil {
			t.Fatalf("Failed to delete cluster: %v", err)
		}
	})

	// Test 4: Cleanup
	t.Run("Cleanup", func(t *testing.T) {
		// Create and immediately delete cluster
		err := provider.CreateCluster(ctx, cluster)
		if err != nil {
			t.Fatalf("Failed to create cluster: %v", err)
		}

		err = provider.DeleteCluster(ctx, cluster)
		if err != nil {
			t.Fatalf("Failed to delete cluster: %v", err)
		}

		// Verify all resources are cleaned up
		exists, err := provider.clusterExists(ctx, cluster)
		if err != nil {
			t.Fatalf("Failed to check cluster existence: %v", err)
		}

		if exists {
			t.Error("Cluster should not exist after deletion")
		}

		// Verify network cleanup
		networkName := provider.getClusterNetworkName(cluster)
		networkExists, err := provider.networkExists(ctx, networkName)
		if err != nil {
			t.Fatalf("Failed to check network existence: %v", err)
		}

		if networkExists {
			t.Error("Network should not exist after deletion")
		}
	})
}

func TestCreateSingleNode(t *testing.T) {
	ctx := context.Background()
	provider, err := NewDockerProvider(&v1alpha1.DockerProviderConfig{
		Spec: v1alpha1.DockerProviderConfigSpec{
			Network: v1alpha1.NetworkConfig{
				CIDR:          "10.10.0.0/16",
				SubnetMask:    24,
				ExposedPorts:  []int32{6443},
				EnableIPv6:    false,
				DNSNameserver: "8.8.8.8",
			},
		},
	})
	if err != nil {
		t.Fatalf("Failed to create Docker provider: %v", err)
	}

	cluster := &v1alpha1.Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cluster.mini-k8s.io/v1alpha1",
			Kind:       "Cluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-single-node",
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
		},
	}

	// Create network first
	networkID, err := provider.createNetwork(ctx, cluster)
	if err != nil {
		t.Fatalf("Failed to create network: %v", err)
	}
	
	// Create a single control plane node
	nodeName := provider.getNodeName(cluster, "control-plane", 0)
	err = provider.createNode(ctx, cluster, nodeName, "control-plane", networkID, cluster.Spec.ControlPlane.MachineConfig)
	if err != nil {
		t.Fatalf("Failed to create node: %v", err)
	}

	// Clean up
	err = provider.DeleteCluster(ctx, cluster)
	if err != nil {
		t.Fatalf("Failed to clean up: %v", err)
	}
}

// Helper types for testing
type NetworkInfo struct {
	CIDR         string
	PortsExposed map[int32]bool
}

type NodeResources struct {
	CPU    int32
	Memory string
}

type ClusterResources struct {
	ControlPlane NodeResources
	Workers      []NodeResources
}
