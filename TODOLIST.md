# Mini-K8s-Manager Development Tasks

## Prerequisites

### Required Tools
```
1. devbox (for isolated development environment)
2. Docker (latest version)
3. Go (1.21 or later)
4. kubectl
5. kind (for local development)
6. kubebuilder (for CRD generation)
```

### Development Environment Setup
```
Task: Set up isolated development environment with devbox
Prompt: 
1. Install devbox:
   curl -fsSL https://get.jetpack.io/devbox | bash

2. Create devbox.json with configuration:
   - Required packages: Go 1.21.6, kubectl, kind, Docker
   - Environment variables: GOPATH, KUBEBUILDER_ASSETS
   - Init hooks for Kubebuilder installation
   - Development scripts for common tasks

3. Initialize and enter devbox environment:
   devbox init
   devbox shell

Test: Verify environment setup:
- All tools accessible within devbox shell
- Correct versions installed
- Development scripts working
```

### Initial Setup
```
Task: Set up development environment
Prompt: Install and verify the following tools:
1. Install kubebuilder:
   curl -L -o kubebuilder https://go.kubebuilder.io/dl/latest/$(go env GOOS)/$(go env GOARCH)
   chmod +x kubebuilder && mv kubebuilder /usr/local/bin/

2. Initialize kubebuilder project:
   kubebuilder init --domain mini-k8s.io --repo github.com/yourusername/mini-k8s-manager

3. Create initial CRD:
   kubebuilder create api --group cluster --version v1alpha1 --kind Cluster

Test: Verify all tools are installed and working:
- docker --version
- go version
- kubectl version
- kind --version
- kubebuilder version
```

## Phase 1: Project Setup and Core API Types

### 1. Project Initialization 
```
Task: Initialize the project structure and basic Go module
Prompt: Create a new Go project with module name github.com/yourusername/mini-k8s-manager with the following directory structure:
- cmd/manager/
- pkg/api/v1alpha1/
- pkg/cluster/
- pkg/providers/
- pkg/controllers/
- deploy/
- examples/
Include basic go.mod and necessary initial dependencies.
Test: Verify project builds with go build ./...

Status: Completed
- Created all required directories
- Initialized Go module
- Added basic main.go
- Verified build succeeds
```

### 2. Core API Types - Cluster 
```
Task: Implement core Cluster API type
Prompt: Create a Cluster CRD type that includes:
- Basic TypeMeta and ObjectMeta
- ClusterSpec with KubernetesVersion, ControlPlane config, and WorkerNode config
- ClusterStatus for current state
Test Cases:
1. Test Cluster struct creation with all fields
2. Test validation of required fields
3. Test deep copy implementation
4. Test serialization to/from YAML

Status: Completed
- Created Cluster CRD with all required fields and validation
- Implemented proper deep copy methods
- Added comprehensive test coverage
- All tests passing
```

### 3. Core API Types - Provider 
```
Task: Implement Provider configuration types
Prompt: Create Provider-specific configuration types starting with Docker:
- Create DockerProviderConfig CRD
- Include network configuration
- Include resource limits
Test Cases:
1. Test provider config validation
2. Test default values
3. Test config merge behavior

Status: Completed
- Created DockerProviderConfig CRD with network and resource configurations
- Added validation using kubebuilder markers
- Implemented config merge functionality
- Added comprehensive test coverage
- All tests passing
```

## Phase 2: Provider Implementation

### 4. Provider Interface 
```
Task: Define the Provider interface
Prompt: Create a Provider interface that defines:
- CreateCluster method
- DeleteCluster method
- GetClusterStatus method
- UpdateCluster method
Test Cases:
1. Create mock provider implementing interface
2. Test interface method signatures
3. Test error conditions

Status: Completed
- Defined Provider interface with all required methods
- Created common error types for provider operations
- Implemented BaseProvider with common validation logic
- Created MockProvider for testing
- Added comprehensive test coverage
- All tests passing
```

### 5. Docker Provider
```
Task: Implement Docker provider
Prompt: Create Docker provider implementation that:
- Manages container lifecycle
- Sets up container networking
- Handles resource allocation
Test Cases:
1. Test container creation
2. Test network setup
3. Test resource allocation
4. Test cleanup
```

## Phase 3: Cluster Management

### 6. Cluster Controller
```
Task: Implement basic cluster controller
Prompt: Create a controller that:
- Watches Cluster resources
- Implements basic reconciliation loop
- Handles cluster lifecycle
Test Cases:
1. Test controller initialization
2. Test reconciliation loop
3. Test error handling
4. Test status updates
```

### 7. Cluster Operations
```
Task: Implement cluster operations
Prompt: Create operations package for:
- Cluster creation workflow
- Cluster deletion workflow
- Node scaling operations
Test Cases:
1. Test cluster creation flow
2. Test deletion cleanup
3. Test scaling operations
4. Test concurrent operations
```

## Phase 4: Version Management

### 8. Version Controller
```
Task: Implement version management
Prompt: Create version controller that:
- Validates version compatibility
- Manages upgrade process
- Handles rollbacks
Test Cases:
1. Test version validation
2. Test upgrade workflow
3. Test rollback scenarios
```

### 9. Node Upgrader
```
Task: Implement node upgrade mechanism
Prompt: Create node upgrader that:
- Handles control plane upgrades
- Manages worker node upgrades
- Ensures cluster stability
Test Cases:
1. Test control plane upgrade
2. Test worker node upgrade
3. Test upgrade ordering
4. Test failure scenarios
```

## Phase 5: CLI Implementation

### 10. Basic CLI Structure
```
Task: Implement CLI framework
Prompt: Create CLI application with:
- Root command structure
- Common flags and options
- Configuration loading
Test Cases:
1. Test command registration
2. Test flag parsing
3. Test config loading
```

### 11. CLI Commands
```
Task: Implement core CLI commands
Prompt: Create commands for:
- Cluster creation
- Cluster deletion
- Status checking
- Version upgrade
Test Cases:
1. Test each command execution
2. Test input validation
3. Test output formatting
```

## Phase 6: Integration

### 12. End-to-End Testing
```
Task: Implement E2E tests
Prompt: Create end-to-end tests that:
- Test full cluster lifecycle
- Verify all operations
- Test upgrade scenarios
Test Cases:
1. Full cluster creation and deletion
2. Version upgrade workflow
3. Error recovery scenarios
```

### 13. Documentation
```
Task: Create comprehensive documentation
Prompt: Write documentation including:
- Architecture overview
- API reference
- User guide
- Developer guide
Test: Verify documentation accuracy and completeness
```

## Development Guidelines

1. **TDD Approach**
   - Write tests first
   - Implement minimum code to pass tests
   - Refactor while maintaining test coverage

2. **Code Quality**
   - Maintain >80% test coverage
   - Follow Go best practices
   - Use consistent error handling

3. **Git Workflow**
   - Create feature branches
   - Write descriptive commit messages
   - Include tests with each feature

4. **Documentation**
   - Document as you code
   - Include examples
   - Keep README updated

## Progress Tracking

- [x] Phase 1 Complete
- [ ] Phase 2 Complete
- [ ] Phase 3 Complete
- [ ] Phase 4 Complete
- [ ] Phase 5 Complete
- [ ] Phase 6 Complete

Each task should be completed with:
1. Tests written first
2. Implementation to pass tests
3. Documentation updated
4. Code reviewed
5. Integration tests passing
