## About
This is a Kubernetes controller that automatically adds a label (`k8c.io/uses-container-linux`)
to Nodes running Flatcar Container Linux as their base operating system.
## Milestones
- [x] Watch k8s node objects
- [x] Check for nodes using Flatcar Container Linux
- [x] Attach a label (`k8c.io/uses-container-linux:‌‌'true'`) to the Node if it uses FC Linux. 
- [x] Write a Dockerfile for the controller 
- [x] Write a Kubernetes Deployment for the controller
- [x] Write the RBAC manifests required for the controller

## Usage
### out-of-cluster
### in-cluster

## Testing

## Resources