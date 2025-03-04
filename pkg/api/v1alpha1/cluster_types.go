//go:generate controller-gen object paths="./..."

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Cluster is the Schema for the clusters API
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

// DeepCopyObject implements runtime.Object interface
func (c *Cluster) DeepCopyObject() runtime.Object {
	copy := &Cluster{}
	c.DeepCopyInto(copy)
	return copy
}

// DeepCopyInto copies all properties of this object into another object of the same type
func (c *Cluster) DeepCopyInto(out *Cluster) {
	*out = *c
	out.TypeMeta = c.TypeMeta
	c.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	c.Spec.DeepCopyInto(&out.Spec)
	c.Status.DeepCopyInto(&out.Status)
}

// DeepCopy creates a deep copy of Cluster
func (c *Cluster) DeepCopy() *Cluster {
	if c == nil {
		return nil
	}
	out := new(Cluster)
	c.DeepCopyInto(out)
	return out
}

// ClusterSpec defines the desired state of Cluster
type ClusterSpec struct {
	// KubernetesVersion is the version of Kubernetes to deploy
	// +kubebuilder:validation:Required
	KubernetesVersion string `json:"kubernetesVersion"`

	// ControlPlane defines the desired state of the control plane nodes
	// +kubebuilder:validation:Required
	ControlPlane ControlPlaneConfig `json:"controlPlane"`

	// Workers defines the desired state of the worker nodes
	// +kubebuilder:validation:Required
	Workers WorkerConfig `json:"workers"`
}

// DeepCopyInto copies all properties of this object into another object of the same type
func (in *ClusterSpec) DeepCopyInto(out *ClusterSpec) {
	*out = *in
	in.ControlPlane.DeepCopyInto(&out.ControlPlane)
	in.Workers.DeepCopyInto(&out.Workers)
}

// ControlPlaneConfig defines the configuration for control plane nodes
type ControlPlaneConfig struct {
	// Count is the number of control plane nodes
	// +kubebuilder:validation:Minimum=1
	Count int32 `json:"count"`

	// MachineConfig defines the hardware configuration for the nodes
	MachineConfig MachineConfig `json:"machineConfig"`
}

// DeepCopyInto copies all properties of this object into another object of the same type
func (in *ControlPlaneConfig) DeepCopyInto(out *ControlPlaneConfig) {
	*out = *in
	in.MachineConfig.DeepCopyInto(&out.MachineConfig)
}

// WorkerConfig defines the configuration for worker nodes
type WorkerConfig struct {
	// Count is the number of worker nodes
	// +kubebuilder:validation:Minimum=0
	Count int32 `json:"count"`

	// MachineConfig defines the hardware configuration for the nodes
	MachineConfig MachineConfig `json:"machineConfig"`
}

// DeepCopyInto copies all properties of this object into another object of the same type
func (in *WorkerConfig) DeepCopyInto(out *WorkerConfig) {
	*out = *in
	in.MachineConfig.DeepCopyInto(&out.MachineConfig)
}

// MachineConfig defines the hardware configuration for a node
type MachineConfig struct {
	// Memory is the amount of memory to allocate to the node (e.g., "2Gi")
	Memory string `json:"memory"`

	// CPUCount is the number of CPUs to allocate to the node
	// +kubebuilder:validation:Minimum=1
	CPUCount int32 `json:"cpuCount"`
}

// DeepCopyInto copies all properties of this object into another object of the same type
func (in *MachineConfig) DeepCopyInto(out *MachineConfig) {
	*out = *in
}

// ClusterStatus defines the observed state of Cluster
type ClusterStatus struct {
	// Phase represents the current phase of cluster actuation
	// +kubebuilder:validation:Enum=Pending;Provisioning;Running;Updating;Failed;Deleting
	Phase ClusterPhase `json:"phase,omitempty"`

	Message string `json:"message,omitempty"`

	// Conditions represent the latest available observations of an object's state
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ControlPlaneReady indicates if the control plane is ready
	ControlPlaneReady bool `json:"controlPlaneReady"`

	// WorkersReady indicates the number of workers that are ready
	WorkersReady int32 `json:"workersReady"`
}

// DeepCopyInto copies all properties of this object into another object of the same type
func (in *ClusterStatus) DeepCopyInto(out *ClusterStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterList contains a list of Cluster
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}

// DeepCopyObject implements runtime.Object interface
func (c *ClusterList) DeepCopyObject() runtime.Object {
	copy := &ClusterList{}
	c.DeepCopyInto(copy)
	return copy
}

// DeepCopyInto copies all properties of this object into another object of the same type
func (c *ClusterList) DeepCopyInto(out *ClusterList) {
	*out = *c
	out.TypeMeta = c.TypeMeta
	c.ListMeta.DeepCopyInto(&out.ListMeta)
	if c.Items != nil {
		in, out := &c.Items, &out.Items
		*out = make([]Cluster, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy creates a deep copy of ClusterList
func (c *ClusterList) DeepCopy() *ClusterList {
	if c == nil {
		return nil
	}
	out := new(ClusterList)
	c.DeepCopyInto(out)
	return out
}

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}
