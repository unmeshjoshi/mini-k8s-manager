package v1alpha1

// ClusterPhase represents the current phase of cluster actuation
type ClusterPhase string

const (
	// ClusterPhasePending indicates the cluster is pending creation
	ClusterPhasePending ClusterPhase = "Pending"

	// ClusterPhaseProvisioning indicates the cluster is being provisioned
	ClusterPhaseProvisioning ClusterPhase = "Provisioning"

	// ClusterPhaseRunning indicates the cluster is running
	ClusterPhaseRunning ClusterPhase = "Running"

	// ClusterPhaseUpdating indicates the cluster is being updated
	ClusterPhaseUpdating ClusterPhase = "Updating"

	// ClusterPhaseFailed indicates the cluster operation has failed
	ClusterPhaseFailed ClusterPhase = "Failed"

	// ClusterPhaseDeleting indicates the cluster is being deleted
	ClusterPhaseDeleting ClusterPhase = "Deleting"
)
