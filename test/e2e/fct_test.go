package e2e

import (
	"context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

// create 5 nodes - some with flatcar, some with other image
// check that fetching by os actually returns the correct number of nodes
// run controller
// fetch nodes by label and ensure that the number matches up.

func TestNodeLabelController(t *testing.T) {
	client := fake.NewSimpleClientset()
	nodeSpec := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "test-node-",
		},
		Status: v1.NodeStatus{
			NodeInfo: v1.NodeSystemInfo{
				OperatingSystem: "l",
			},
		},
	}

	t.Run("create new node", func(t *testing.T) {
		node, err := client.CoreV1().Nodes().Create(context.TODO(), nodeSpec, metav1.CreateOptions{})
		if err != nil {
			t.Errorf("Failed to create node: %s", err.Error())
		}
		if node == nil {
			t.Errorf("Created nil node")
		} else if node.Status.NodeInfo.OperatingSystem != "l" {
			t.Errorf("mismatched operating system name for node %s. Expected %s, got %s",
				node.Name, "l", node.Status.NodeInfo.OperatingSystem)
		}
	})

	t.Run("error out on invalid config", func(t *testing.T) {
	})

	t.Run("correctly figure out flatcar nodes", func(t *testing.T) {

	})

	t.Run("correctly label flatcar nodes", func(t *testing.T) {

	})
}