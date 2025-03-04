package v1alpha1

import (
	"testing"
	"encoding/json"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestClusterCreation(t *testing.T) {
	cluster := &Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cluster.mini-k8s.io/v1alpha1",
			Kind:       "Cluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster",
			Namespace: "default",
		},
		Spec: ClusterSpec{
			KubernetesVersion: TestKubernetesVersion,
			ControlPlane: ControlPlaneConfig{
				Count: 1,
				MachineConfig: MachineConfig{
					Memory:   "2Gi",
					CPUCount: 2,
				},
			},
			Workers: WorkerConfig{
				Count: 2,
				MachineConfig: MachineConfig{
					Memory:   "4Gi",
					CPUCount: 2,
				},
			},
		},
	}

	// Test 1: Validate required fields
	if cluster.Spec.KubernetesVersion == "" {
		t.Error("KubernetesVersion should not be empty")
	}
	if cluster.Spec.ControlPlane.Count < 1 {
		t.Error("ControlPlane count should be at least 1")
	}

	// Test 2: Test YAML serialization
	yamlData, err := yaml.Marshal(cluster)
	if err != nil {
		t.Errorf("Failed to marshal cluster to YAML: %v", err)
	}

	var unmarshaledCluster Cluster
	err = yaml.Unmarshal(yamlData, &unmarshaledCluster)
	if err != nil {
		t.Errorf("Failed to unmarshal cluster from YAML: %v", err)
	}

	if unmarshaledCluster.Spec.KubernetesVersion != cluster.Spec.KubernetesVersion {
		t.Error("YAML serialization/deserialization failed to preserve KubernetesVersion")
	}

	// Test 3: Test JSON serialization
	jsonData, err := json.Marshal(cluster)
	if err != nil {
		t.Errorf("Failed to marshal cluster to JSON: %v", err)
	}

	err = json.Unmarshal(jsonData, &unmarshaledCluster)
	if err != nil {
		t.Errorf("Failed to unmarshal cluster from JSON: %v", err)
	}

	if unmarshaledCluster.Spec.Workers.Count != cluster.Spec.Workers.Count {
		t.Error("JSON serialization/deserialization failed to preserve Workers count")
	}
}
