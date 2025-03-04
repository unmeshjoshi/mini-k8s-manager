package v1alpha1

import (
	"testing"
	"encoding/json"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDockerProviderConfig(t *testing.T) {
	config := &DockerProviderConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cluster.mini-k8s.io/v1alpha1",
			Kind:       "DockerProviderConfig",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-docker-config",
			Namespace: "default",
		},
		Spec: DockerProviderConfigSpec{
			Network: NetworkConfig{
				CIDR:          "172.20.0.0/16",
				SubnetMask:    24,
				ExposedPorts:  []int32{80, 443, 6443},
				EnableIPv6:    false,
				DNSNameserver: "8.8.8.8",
			},
			ResourceLimits: ResourceLimitsConfig{
				CPU: ResourceLimit{
					Default: "2",
					Min:     "1",
					Max:     "4",
				},
				Memory: ResourceLimit{
					Default: "2Gi",
					Min:     "1Gi",
					Max:     "8Gi",
				},
				Storage: ResourceLimit{
					Default: "20Gi",
					Min:     "10Gi",
					Max:     "100Gi",
				},
			},
		},
	}

	// Test 1: Validate required fields
	if config.Spec.Network.CIDR == "" {
		t.Error("Network CIDR should not be empty")
	}
	if config.Spec.Network.SubnetMask == 0 {
		t.Error("Network SubnetMask should not be 0")
	}

	// Test 2: Test default values
	if !config.Spec.Network.ValidateDefaults() {
		t.Error("Network configuration should have valid defaults")
	}

	// Test 3: Test YAML serialization
	yamlData, err := yaml.Marshal(config)
	if err != nil {
		t.Errorf("Failed to marshal config to YAML: %v", err)
	}

	var unmarshaledConfig DockerProviderConfig
	err = yaml.Unmarshal(yamlData, &unmarshaledConfig)
	if err != nil {
		t.Errorf("Failed to unmarshal config from YAML: %v", err)
	}

	if unmarshaledConfig.Spec.Network.CIDR != config.Spec.Network.CIDR {
		t.Error("YAML serialization/deserialization failed to preserve Network CIDR")
	}

	// Test 4: Test JSON serialization
	jsonData, err := json.Marshal(config)
	if err != nil {
		t.Errorf("Failed to marshal config to JSON: %v", err)
	}

	err = json.Unmarshal(jsonData, &unmarshaledConfig)
	if err != nil {
		t.Errorf("Failed to unmarshal config from JSON: %v", err)
	}

	if unmarshaledConfig.Spec.ResourceLimits.CPU.Default != config.Spec.ResourceLimits.CPU.Default {
		t.Error("JSON serialization/deserialization failed to preserve CPU Default")
	}

	// Test 5: Test config merge behavior
	defaultConfig := DockerProviderConfig{
		Spec: DockerProviderConfigSpec{
			Network: NetworkConfig{
				CIDR:          "10.0.0.0/16",
				SubnetMask:    24,
				EnableIPv6:    false,
				DNSNameserver: "8.8.8.8",
			},
			ResourceLimits: ResourceLimitsConfig{
				CPU: ResourceLimit{
					Default: "1",
					Min:     "0.5",
					Max:     "2",
				},
			},
		},
	}

	merged := defaultConfig.MergeWith(config)
	if merged.Spec.Network.CIDR != config.Spec.Network.CIDR {
		t.Error("Merge should prefer non-default CIDR")
	}
	if len(merged.Spec.Network.ExposedPorts) != len(config.Spec.Network.ExposedPorts) {
		t.Error("Merge should preserve exposed ports")
	}
}
