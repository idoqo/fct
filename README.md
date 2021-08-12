## flatcar-tag-controller (fct/fct-controller)

This is a Kubernetes controller that automatically adds a label (`k8c.io/uses-container-linux`)
to Nodes running Flatcar Container Linux as their base operating system. Nodes are detected by checking
for the string "flatcar" in the operating system name.
## Milestones
- [x] Watch k8s node objects
- [x] Check for nodes using Flatcar Container Linux.
- [x] Attach a label (`k8c.io/uses-container-linux:‌‌'true'`) to the Node if it uses FC Linux. 
- [x] Write a Dockerfile for the controller 
- [x] Write a Kubernetes Deployment for the controller
- [x] Write the RBAC manifests required for the controller

## Usage
### out-of-cluster
Clone the repo and in the `node-label-controller` folder, build the application binary by running:
```bash
make build
```
The command will generate the executable in `$PWD/build/fct`.
Next, run `./build/fct` to watch the nodes in your cluster. 

In out-of-cluster mode, `fct` will prioritize the value of your `$KUBECONFIG`
environment variable over `$HOME/.kube/config`. As a result, you can run it on your choice
cluster by setting `$KUBECONFIG` appropriately. 
### in-cluster
Install `fct` in a cluster by running:
```bash
make install
```

The above command will:
- create a service account named `fct-sa` for the controller to use.
- a cluster role with permissions to {get, list, watch, and patch} nodes in the cluster.
- create a cluster role-binding that binds the service account to the cluster role.
- apply the fct-deployment that is based on the `idoko/fct` docker image.

You can remove the controller from the cluster by running:
```bash
make uninstall
```

## Limitations
- `kubectl` keeps timing out for me when using the provided cluster, so I haven't been able to test the
controller against it yet.
- I'm not sure about the correct OS name for nodes running Flatcar Container Linux hence, controller checks
  if the operating system name contains the string "flatcar". 
Updating it to the correct name should be trivial (by updating the value of `container.FlatcarOSName`) and
  checking for equality. This would be done once I'm able to access a node running Flatcar.
  
## Checks
The controller works against:
- a fake cluster (using the unit tests in `pkg/controller/controller_test.go`)
## Resources
- https://github.com/kubernetes-sigs/controller-runtime
- https://github.com/kubernetes/sample-controller/
- https://github.com/bitnami-labs/kubewatch
