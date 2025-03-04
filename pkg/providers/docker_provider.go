package providers

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/image"
	"net"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/unmeshjoshi/mini-k8s-manager/pkg/api/v1alpha1"
)

// DockerProvider implements the Provider interface for Docker
type DockerProvider struct {
	*BaseProvider
	config *v1alpha1.DockerProviderConfig
	client *client.Client
}

// containerInfo represents container information for testing
type containerInfo struct {
	ID     string
	Names  []string
	Labels map[string]string
	State  string
}

// NewDockerProvider creates a new Docker provider instance
func NewDockerProvider(config *v1alpha1.DockerProviderConfig) (*DockerProvider, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithVersion("1.45"))
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	return &DockerProvider{
		BaseProvider: &BaseProvider{},
		config:       config,
		client:       cli,
	}, nil
}

// getClusterNetworkName returns the Docker network name for the cluster
func (p *DockerProvider) getClusterNetworkName(cluster *v1alpha1.Cluster) string {
	return fmt.Sprintf("cluster-%s-net", cluster.Name)
}

// getNodeName returns the Docker container name for a node
func (p *DockerProvider) getNodeName(cluster *v1alpha1.Cluster, role string, index int) string {
	return fmt.Sprintf("cluster-%s-%s-%d", cluster.Name, role, index)
}

// getClusterFilters returns Docker filters for the cluster resources
func (p *DockerProvider) getClusterFilters(cluster *v1alpha1.Cluster) filters.Args {
	filters := filters.NewArgs()
	filters.Add("label", fmt.Sprintf("cluster=%s", cluster.Name))
	return filters
}

// parseMemory converts memory string (e.g., "2Gi") to bytes
func (p *DockerProvider) parseMemory(memory string) int64 {
	memory = strings.ToUpper(memory)
	var multiplier int64

	switch {
	case strings.HasSuffix(memory, "KI"):
		multiplier = 1024
	case strings.HasSuffix(memory, "MI"):
		multiplier = 1024 * 1024
	case strings.HasSuffix(memory, "GI"):
		multiplier = 1024 * 1024 * 1024
	default:
		return 0
	}

	value := strings.TrimSuffix(strings.TrimSuffix(memory, "I"), "KMG")
	var bytes int64
	fmt.Sscanf(value, "%d", &bytes)
	return bytes * multiplier
}

// createNetwork creates a Docker network for the cluster
func (p *DockerProvider) createNetwork(ctx context.Context, cluster *v1alpha1.Cluster) (string, error) {
	networkName := p.getClusterNetworkName(cluster)

	// Check if network already exists.
	filterArgs := filters.NewArgs()
	filterArgs.Add("name", networkName)
	networks, err := p.client.NetworkList(ctx, network.ListOptions{Filters: filterArgs})
	if err != nil {
		return "", fmt.Errorf("failed to list networks: %w", err)
	}

	if len(networks) > 0 {
		return networks[0].ID, nil
	}

	// Prepare IPAM configuration.
	ipamConfig := []network.IPAMConfig{
		{
			Subnet:  p.config.Spec.Network.CIDR,
			Gateway: p.getNetworkGateway(p.config.Spec.Network.CIDR),
		},
	}
	ipam := &network.IPAM{
		Driver: "default",
		Config: ipamConfig,
	}

	// Use the new network.CreateOptions.
	networkCreateOpts := network.CreateOptions{
		Driver:     "bridge",
		IPAM:       ipam,
		EnableIPv6: &p.config.Spec.Network.EnableIPv6,
		Internal:   false,
		Attachable: false,
		Ingress:    false,
		// Note: "Scope" is no longer a field in CreateOptions.
	}

	// Create the network.
	netResp, err := p.client.NetworkCreate(ctx, networkName, networkCreateOpts)
	if err != nil {
		return "", fmt.Errorf("failed to create network: %w", err)
	}

	return netResp.ID, nil
}

// getNetworkGateway returns the gateway IP for a given CIDR
func (p *DockerProvider) getNetworkGateway(cidr string) string {
	_, ipNet, _ := net.ParseCIDR(cidr)
	if ipNet == nil {
		return ""
	}

	// Get the first usable IP in the subnet as gateway
	ip := ipNet.IP.To4()
	ip[3]++
	return ip.String()
}

// createNode creates a Docker container for a Kubernetes node
func (p *DockerProvider) createNode(ctx context.Context, cluster *v1alpha1.Cluster, nodeName, role, networkID string, machineConfig v1alpha1.MachineConfig) error {
	fmt.Printf("Creating node %s with role %s\n", nodeName, role)

	// Pull the node image first
	imageRef := fmt.Sprintf("kindest/node:%s", cluster.Spec.KubernetesVersion)
	fmt.Printf("Pulling image %s...\n", imageRef)
	reader, err := p.client.ImagePull(ctx, imageRef, image.PullOptions{})
	if err != nil {
		fmt.Printf("Failed to pull image: %v\n", err)
		return fmt.Errorf("failed to pull node image: %w", err)
	}
	fmt.Printf("Successfully pulled image %s\n", imageRef)
	defer reader.Close()

	// Create container configuration
	config := &container.Config{
		Image:    fmt.Sprintf("kindest/node:%s", cluster.Spec.KubernetesVersion),
		Hostname: nodeName,
		Labels: map[string]string{
			"cluster": cluster.Name,
			"role":    role,
		},
	}

	// Create host configuration
	hostConfig := &container.HostConfig{
		Privileged: true,
		Resources: container.Resources{
			Memory:   p.parseMemory(machineConfig.Memory),
			NanoCPUs: int64(machineConfig.CPUCount * 1e9), // Convert CPU count to nanoCPUs
		},
	}

	// Create network configuration
	networkingConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			p.getClusterNetworkName(cluster): {
				NetworkID: networkID,
			},
		},
	}

	// Create container
	fmt.Printf("Creating container with name %s...\n", nodeName)
	contResp, err := p.client.ContainerCreate(ctx, config, hostConfig, networkingConfig, nil, nodeName)
	if err != nil {
		fmt.Printf("Failed to create container: %v\n", err)
		return fmt.Errorf("failed to create container: %w", err)
	}
	fmt.Printf("Successfully created container %s with ID %s\n", nodeName, contResp.ID)

	// Start container
	fmt.Printf("Starting container %s...\n", contResp.ID)
	if err := p.client.ContainerStart(ctx, contResp.ID, container.StartOptions{}); err != nil {
		fmt.Printf("Failed to start container: %v\n", err)
		return fmt.Errorf("failed to start container: %w", err)
	}
	fmt.Printf("Successfully started container %s\n", contResp.ID)

	return nil
}

// CreateCluster creates a new Kubernetes cluster using Docker containers
func (p *DockerProvider) CreateCluster(ctx context.Context, cluster *v1alpha1.Cluster) error {
	fmt.Printf("Creating cluster %s...\n", cluster.Name)

	// Check if cluster already exists
	fmt.Printf("Checking if cluster already exists...\n")
	exists, err := p.clusterExists(ctx, cluster)
	if err != nil {
		fmt.Printf("Failed to check if cluster exists: %v\n", err)
		return fmt.Errorf("failed to check if cluster exists: %w", err)
	}
	if exists {
		return ErrClusterExists
	}

	// Create network
	fmt.Printf("Creating network for cluster %s...\n", cluster.Name)
	networkID, err := p.createNetwork(ctx, cluster)
	if err != nil {
		fmt.Printf("Failed to create network: %v\n", err)
		return fmt.Errorf("failed to create network: %w", err)
	}
	fmt.Printf("Successfully created network with ID %s\n", networkID)

	// Create control plane nodes
	fmt.Printf("Creating %d control plane nodes...\n", cluster.Spec.ControlPlane.Count)
	for i := 0; i < int(cluster.Spec.ControlPlane.Count); i++ {
		nodeName := p.getNodeName(cluster, "control-plane", i)
		fmt.Printf("Creating control plane node %s (%d/%d)...\n", nodeName, i+1, cluster.Spec.ControlPlane.Count)
		if err := p.createNode(ctx, cluster, nodeName, "control-plane", networkID, cluster.Spec.ControlPlane.MachineConfig); err != nil {
			fmt.Printf("Failed to create control plane node %s: %v\n", nodeName, err)
			// Cleanup on failure
			p.DeleteCluster(ctx, cluster)
			return fmt.Errorf("failed to create control plane node %s: %w", nodeName, err)
		}
		fmt.Printf("Successfully created control plane node %s\n", nodeName)
	}

	// Create worker nodes
	fmt.Printf("Creating %d worker nodes...\n", cluster.Spec.Workers.Count)
	for i := 0; i < int(cluster.Spec.Workers.Count); i++ {
		nodeName := p.getNodeName(cluster, "worker", i)
		fmt.Printf("Creating worker node %s (%d/%d)...\n", nodeName, i+1, cluster.Spec.Workers.Count)
		if err := p.createNode(ctx, cluster, nodeName, "worker", networkID, cluster.Spec.Workers.MachineConfig); err != nil {
			fmt.Printf("Failed to create worker node %s: %v\n", nodeName, err)
			// Cleanup on failure
			p.DeleteCluster(ctx, cluster)
			return fmt.Errorf("failed to create worker node %s: %w", nodeName, err)
		}
		fmt.Printf("Successfully created worker node %s\n", nodeName)
	}

	return nil
}

// DeleteCluster deletes an existing Kubernetes cluster
func (p *DockerProvider) DeleteCluster(ctx context.Context, cluster *v1alpha1.Cluster) error {
	fmt.Printf("Deleting cluster %s...\n", cluster.Name)

	// List all containers for this cluster
	fmt.Printf("Listing containers in cluster %s...\n", cluster.Name)
	containers, err := p.client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: p.getClusterFilters(cluster),
	})
	if err != nil {
		fmt.Printf("Failed to list containers: %v\n", err)
		return fmt.Errorf("failed to list containers: %w", err)
	}
	fmt.Printf("Found %d containers to delete\n", len(containers))

	// Stop and remove all containers
	for _, cont := range containers {
		// Stop container with a timeout
		fmt.Printf("Stopping container %s...\n", cont.ID[:12])
		timeoutSeconds := int(30)
		if err := p.client.ContainerStop(ctx, cont.ID, container.StopOptions{Timeout: &timeoutSeconds}); err != nil {
			fmt.Printf("Failed to stop container %s: %v\n", cont.ID[:12], err)
			return fmt.Errorf("failed to stop container %s: %w", cont.ID, err)
		}
		fmt.Printf("Successfully stopped container %s\n", cont.ID[:12])

		// Remove container
		fmt.Printf("Removing container %s...\n", cont.ID[:12])
		if err := p.client.ContainerRemove(ctx, cont.ID, container.RemoveOptions{
			RemoveVolumes: true,
			Force:         true,
		}); err != nil {
			fmt.Printf("Failed to remove container %s: %v\n", cont.ID[:12], err)
			return fmt.Errorf("failed to remove container %s: %w", cont.ID, err)
		}
		fmt.Printf("Successfully removed container %s\n", cont.ID[:12])
	}

	// Remove network
	fmt.Printf("Listing networks for cluster %s...\n", cluster.Name)
	networks, err := p.client.NetworkList(ctx, network.ListOptions{
		Filters: p.getClusterFilters(cluster),
	})
	if err != nil {
		fmt.Printf("Failed to list networks: %v\n", err)
		return fmt.Errorf("failed to list networks: %w", err)
	}
	fmt.Printf("Found %d networks to delete\n", len(networks))

	for _, net := range networks {
		fmt.Printf("Removing network %s...\n", net.Name)
		if err := p.client.NetworkRemove(ctx, net.ID); err != nil {
			fmt.Printf("Failed to remove network %s: %v\n", net.Name, err)
			return fmt.Errorf("failed to remove network %s: %w", net.ID, err)
		}
		fmt.Printf("Successfully removed network %s\n", net.Name)
	}

	return nil
}

// GetClusterStatus returns the current status of the cluster
func (p *DockerProvider) GetClusterStatus(ctx context.Context, cluster *v1alpha1.Cluster) (*v1alpha1.ClusterStatus, error) {
	status := &v1alpha1.ClusterStatus{}

	// List all containers for this cluster
	containers, err := p.client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: p.getClusterFilters(cluster),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	// Count control plane and worker nodes
	var controlPlaneCount, workerCount int32
	allRunning := true

	for _, cont := range containers {
		if cont.State != "running" {
			allRunning = false
			continue
		}

		switch cont.Labels["role"] {
		case "control-plane":
			controlPlaneCount++
		case "worker":
			workerCount++
		}
	}

	// Set status fields
	if len(containers) == 0 {
		status.Phase = "NotFound"
	} else if !allRunning {
		status.Phase = "Starting"
	} else if controlPlaneCount == cluster.Spec.ControlPlane.Count &&
		workerCount == cluster.Spec.Workers.Count {
		status.Phase = "Running"
	} else {
		status.Phase = "Pending"
	}

	status.ControlPlaneReady = (controlPlaneCount == cluster.Spec.ControlPlane.Count)
	status.WorkersReady = workerCount

	return status, nil
}

// clusterExists checks if a cluster with the given name already exists
func (p *DockerProvider) clusterExists(ctx context.Context, cluster *v1alpha1.Cluster) (bool, error) {
	containers, err := p.client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: p.getClusterFilters(cluster),
	})
	if err != nil {
		return false, fmt.Errorf("failed to list containers: %w", err)
	}
	return len(containers) > 0, nil
}

// networkExists checks if a network with the given name exists
func (p *DockerProvider) networkExists(ctx context.Context, networkName string) (bool, error) {
	networks, err := p.client.NetworkList(ctx, network.ListOptions{
		Filters: filters.NewArgs(filters.Arg("name", networkName)),
	})
	if err != nil {
		return false, fmt.Errorf("failed to list networks: %w", err)
	}
	return len(networks) > 0, nil
}

// getNetworkInfo returns network configuration for a node
func (p *DockerProvider) getNetworkInfo(networkID string) *network.NetworkingConfig {
	return &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			networkID: {
				NetworkID: networkID,
			},
		},
	}
}

// getResourceAllocation returns container resource allocation configuration
func (p *DockerProvider) getResourceAllocation(config v1alpha1.MachineConfig) *container.Resources {
	memoryBytes := p.parseMemory(config.Memory)
	return &container.Resources{
		Memory:    memoryBytes,
		CPUQuota:  int64(config.CPUCount * 100000),
		CPUPeriod: 100000,
	}
}

// UpdateCluster updates an existing cluster's configuration
func (p *DockerProvider) UpdateCluster(ctx context.Context, cluster *v1alpha1.Cluster) error {
	// Check if cluster exists
	exists, err := p.clusterExists(ctx, cluster)
	if err != nil {
		return fmt.Errorf("failed to check cluster existence: %w", err)
	}
	if !exists {
		return ErrClusterNotFound
	}

	// For now, we'll implement a simple update that only supports scaling workers
	// Get current cluster status
	status, err := p.GetClusterStatus(ctx, cluster)
	if err != nil {
		return fmt.Errorf("failed to get cluster status: %w", err)
	}

	// Calculate the difference in worker count
	currentWorkers := status.WorkersReady
	desiredWorkers := cluster.Spec.Workers.Count

	if currentWorkers == desiredWorkers {
		// No changes needed
		return nil
	}

	// Get network ID
	networkName := p.getClusterNetworkName(cluster)
	networks, err := p.client.NetworkList(ctx, network.ListOptions{Filters: filters.NewArgs(filters.Arg("name", networkName))})
	if err != nil || len(networks) == 0 {
		return fmt.Errorf("failed to get cluster network: %w", err)
	}
	networkID := networks[0].ID

	if desiredWorkers > currentWorkers {
		// Scale up: Create new worker nodes
		for i := currentWorkers; i < desiredWorkers; i++ {
			nodeName := p.getNodeName(cluster, "worker", int(i))
			if err := p.createNode(ctx, cluster, nodeName, "worker", networkID, cluster.Spec.Workers.MachineConfig); err != nil {
				return fmt.Errorf("failed to create worker node %s: %w", nodeName, err)
			}
		}
	} else {
		// Scale down: Remove excess worker nodes
		for i := currentWorkers - 1; i >= desiredWorkers; i-- {
			nodeName := p.getNodeName(cluster, "worker", int(i))

			// Find container by name
			containers, err := p.client.ContainerList(ctx, container.ListOptions{All: true, Filters: filters.NewArgs(filters.Arg("name", nodeName))})
			if err != nil {
				return fmt.Errorf("failed to list containers: %w", err)
			}

			if len(containers) == 0 {
				continue // Container already gone
			}

			// Stop and remove the container
			timeout := 60 // seconds
			if err := p.client.ContainerStop(ctx, containers[0].ID, container.StopOptions{Timeout: &timeout}); err != nil {
				return fmt.Errorf("failed to stop container %s: %w", nodeName, err)
			}

			if err := p.client.ContainerRemove(ctx, containers[0].ID, container.RemoveOptions{Force: true}); err != nil {
				return fmt.Errorf("failed to remove container %s: %w", nodeName, err)
			}
		}
	}

	return nil
}
