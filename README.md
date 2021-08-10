## flatcar-tag-controller (fct/fct-controller)

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
Clone the repo and in the `node-label-controller` folder, build the application binary by running:
```bash
go build -o build/fct ./cmd/flatcartag
```
The command will generate the executable in `$PWD/build/fct`.
Next, run `./build/fct` to watch the nodes in your cluster. 

In out-of-cluster mode, `fct` will prioritize the value of your `$KUBECONFIG`
environment variable over `$HOME/.kube/config`. As a result, you can run it on your choice
cluster by setting `$KUBECONFIG` appropriately. 
### in-cluster
To run `fct` in a cluster, apply the RBAC manifest with:
```bash
kubectl apply -f fct-rbac.yml
```
The above command will:
- create a service account named `fct-sa` for the controller to use.
- a cluster role with permissions to {get, list, watch, and patch} pods and nodes in the cluster.
- create a cluster role-binding that binds the service account to the cluster role.

Next, apply the controller deployment by running:
```bash
kubectl apply -f fct-deployment.yml
```
This will start a pod for the controller using the `idoko/fct` docker image.

## Testing
Tests live in `test/e2e`. They are insufficient though, as they use the mock kubeclient provided by
`"k8s.io/client-go/kubernetes/fake"`. This means labelling doesn't work rightly.
## Resources
- https://github.com/kubernetes-sigs/controller-runtime
- https://github.com/kubernetes/sample-controller/
- https://github.com/bitnami-labs/kubewatch
- 