package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// DockerProviderConfig is the Schema for Docker provider configuration
type DockerProviderConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec DockerProviderConfigSpec `json:"spec,omitempty"`
}

// DockerProviderConfigSpec defines the desired state of DockerProviderConfig
type DockerProviderConfigSpec struct {
	// Network configuration for Docker provider
	Network NetworkConfig `json:"network"`

	// ResourceLimits defines resource constraints for Docker containers
	ResourceLimits ResourceLimitsConfig `json:"resourceLimits"`
}

// NetworkConfig defines the network configuration for Docker provider
type NetworkConfig struct {
	// CIDR range for the Docker network
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^([0-9]{1,3}\.){3}[0-9]{1,3}/[0-9]{1,2}$`
	CIDR string `json:"cidr"`

	// SubnetMask is the subnet mask for the network (e.g., 24 for /24)
	// +kubebuilder:validation:Minimum=16
	// +kubebuilder:validation:Maximum=28
	SubnetMask int32 `json:"subnetMask"`

	// ExposedPorts are the ports to expose from the Docker containers
	// +optional
	ExposedPorts []int32 `json:"exposedPorts,omitempty"`

	// EnableIPv6 enables IPv6 support
	// +optional
	EnableIPv6 bool `json:"enableIPv6,omitempty"`

	// DNSNameserver is the DNS nameserver to use
	// +optional
	DNSNameserver string `json:"dnsNameserver,omitempty"`
}

// ValidateDefaults checks if the network configuration has valid defaults
func (n *NetworkConfig) ValidateDefaults() bool {
	if n.CIDR == "" {
		return false
	}
	if n.SubnetMask < 16 || n.SubnetMask > 28 {
		return false
	}
	if n.DNSNameserver == "" {
		n.DNSNameserver = "8.8.8.8"
	}
	return true
}

// ResourceLimitsConfig defines resource limits for Docker containers
type ResourceLimitsConfig struct {
	// CPU resource limits
	// +kubebuilder:validation:Required
	CPU ResourceLimit `json:"cpu"`

	// Memory resource limits
	// +kubebuilder:validation:Required
	Memory ResourceLimit `json:"memory"`

	// Storage resource limits
	// +kubebuilder:validation:Required
	Storage ResourceLimit `json:"storage"`
}

// ResourceLimit defines min, max, and default values for a resource
type ResourceLimit struct {
	// Default value for the resource
	// +kubebuilder:validation:Required
	Default string `json:"default"`

	// Minimum allowed value for the resource
	// +kubebuilder:validation:Required
	Min string `json:"min"`

	// Maximum allowed value for the resource
	// +kubebuilder:validation:Required
	Max string `json:"max"`
}

// MergeWith merges the current DockerProviderConfig with another one
func (d *DockerProviderConfig) MergeWith(other *DockerProviderConfig) *DockerProviderConfig {
	result := d.DeepCopy()

	if other.Spec.Network.CIDR != "" {
		result.Spec.Network.CIDR = other.Spec.Network.CIDR
	}
	if other.Spec.Network.SubnetMask != 0 {
		result.Spec.Network.SubnetMask = other.Spec.Network.SubnetMask
	}
	if len(other.Spec.Network.ExposedPorts) > 0 {
		result.Spec.Network.ExposedPorts = make([]int32, len(other.Spec.Network.ExposedPorts))
		copy(result.Spec.Network.ExposedPorts, other.Spec.Network.ExposedPorts)
	}
	if other.Spec.Network.DNSNameserver != "" {
		result.Spec.Network.DNSNameserver = other.Spec.Network.DNSNameserver
	}

	// Merge resource limits
	if other.Spec.ResourceLimits.CPU.Default != "" {
		result.Spec.ResourceLimits.CPU = other.Spec.ResourceLimits.CPU
	}
	if other.Spec.ResourceLimits.Memory.Default != "" {
		result.Spec.ResourceLimits.Memory = other.Spec.ResourceLimits.Memory
	}
	if other.Spec.ResourceLimits.Storage.Default != "" {
		result.Spec.ResourceLimits.Storage = other.Spec.ResourceLimits.Storage
	}

	return result
}

// DeepCopyObject implements runtime.Object interface
func (d *DockerProviderConfig) DeepCopyObject() runtime.Object {
	copy := &DockerProviderConfig{}
	d.DeepCopyInto(copy)
	return copy
}

// DeepCopyInto copies all properties of this object into another object of the same type
func (d *DockerProviderConfig) DeepCopyInto(out *DockerProviderConfig) {
	*out = *d
	out.TypeMeta = d.TypeMeta
	d.ObjectMeta.DeepCopyInto(&out.ObjectMeta)

	// Deep copy network config
	out.Spec.Network.CIDR = d.Spec.Network.CIDR
	out.Spec.Network.SubnetMask = d.Spec.Network.SubnetMask
	out.Spec.Network.EnableIPv6 = d.Spec.Network.EnableIPv6
	out.Spec.Network.DNSNameserver = d.Spec.Network.DNSNameserver
	if d.Spec.Network.ExposedPorts != nil {
		out.Spec.Network.ExposedPorts = make([]int32, len(d.Spec.Network.ExposedPorts))
		copy(out.Spec.Network.ExposedPorts, d.Spec.Network.ExposedPorts)
	}

	// Deep copy resource limits
	out.Spec.ResourceLimits = d.Spec.ResourceLimits
}

// DeepCopy creates a deep copy of DockerProviderConfig
func (d *DockerProviderConfig) DeepCopy() *DockerProviderConfig {
	if d == nil {
		return nil
	}
	out := new(DockerProviderConfig)
	d.DeepCopyInto(out)
	return out
}

// +kubebuilder:object:root=true

// DockerProviderConfigList contains a list of DockerProviderConfig
type DockerProviderConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DockerProviderConfig `json:"items"`
}

// DeepCopyInto copies all properties of this object into another object of the same type
func (d *DockerProviderConfigList) DeepCopyInto(out *DockerProviderConfigList) {
	*out = *d
	out.TypeMeta = d.TypeMeta
	d.ListMeta.DeepCopyInto(&out.ListMeta)
	if d.Items != nil {
		out.Items = make([]DockerProviderConfig, len(d.Items))
		for i := range d.Items {
			d.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

func init() {
	SchemeBuilder.Register(&DockerProviderConfig{}, &DockerProviderConfigList{})
}
