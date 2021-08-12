package controller

import (
	"context"
	"fmt"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"
	"strings"
	"testing"
	"time"
)

func TestFCTController(t *testing.T) {
	client := fake.NewSimpleClientset()
	osNames := []string{"ubuntu", "centos", FlatcarOSName}
	expectedFlatcarCount := 0
	ctx := context.TODO()
	stopCh := make(chan struct{})

	generateNodeName := func(os string) string {
		return fmt.Sprintf("test-node-%s", os)
	}
	generateNodeSpec := func(os string) *apiv1.Node {
		if strings.Contains(os, FlatcarOSName) {
			expectedFlatcarCount++
		}

		return &apiv1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: generateNodeName(os),
				Labels: map[string]string{
					"test_label": fmt.Sprintf("label-%s", os),
				},
			},
			Status: apiv1.NodeStatus{
				NodeInfo: apiv1.NodeSystemInfo{
					OperatingSystem: os,
				},
			},
		}
	}

	t.Run("spin up new nodes", func(t *testing.T) {

		for _, os := range osNames {
			nodeSpec := generateNodeSpec(os)
			node, err := client.CoreV1().Nodes().Create(ctx, nodeSpec, metav1.CreateOptions{})
			if err != nil {
				t.Errorf("Failed to create node: %s", err.Error())
			} else if node.Status.NodeInfo.OperatingSystem != os {
				t.Errorf("mismatched operating system name for node %s. Expected %s, got %s",
					node.Name, os, node.Status.NodeInfo.OperatingSystem)
			}
		}
	})

	// start up the controller
	informer := CreateNodeInformer(client)
	ctl := NewController(client, informer)
	go ctl.Run(stopCh)

	// we could use a retry mechanism instead of sleeping, that way
	// we stop execution once we get a success result (of if we timeout).
	time.Sleep(3 * time.Second)

	getFlatcarNodes := func() (*apiv1.NodeList, error) {
		flatcarFilter, err := labels.NewRequirement("k8c.io/uses-container-linux", selection.Equals, []string{"true"})
		if err != nil {
			return nil, err
		}

		selector := labels.NewSelector().Add(*flatcarFilter)
		nodes, err := client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{LabelSelector: selector.String()})
		if err != nil {
			return nil, err
		}
		return nodes, nil
	}

	t.Run("test for properly labelled nodes", func(t *testing.T) {
		nodes, err := getFlatcarNodes()
		if err != nil {
			t.Errorf("failed to get flatcar nodes: %s", err.Error())
			return
		}
		matchCount := len(nodes.Items)
		if matchCount != expectedFlatcarCount {
			t.Errorf("expected %d node(s) with flatcar label, got %d",
				expectedFlatcarCount, matchCount)
			return
		}
		for _, n := range nodes.Items {
			t.Logf("found node: %s", n.Name)
		}
	})

	t.Run("controller relabels node if label is deleted", func(t *testing.T) {
		deleteLabelPatch := fmt.Sprintf(`[{"op": "remove", "path": "/metadata/labels/%s"}]`, usesFlatcarLabelKey)
		_, err := client.CoreV1().Nodes().Patch(ctx, generateNodeName(FlatcarOSName),
			types.JSONPatchType, []byte(deleteLabelPatch), metav1.PatchOptions{})
		if err != nil {
			t.Errorf("failed to remove label from node: %s", err.Error())
		}
		// wait for a moment before checking for labels
		time.Sleep(1 * time.Second)

		nodes, err := getFlatcarNodes()
		if err != nil {
			t.Errorf("failed to get flatcar nodes: %s", err.Error())
			return
		}
		matchCount := len(nodes.Items)
		if matchCount != expectedFlatcarCount {
			t.Errorf("expected %d node(s) with flatcar label, got %d",
				expectedFlatcarCount, matchCount)
			return
		}
		for _, n := range nodes.Items {
			t.Logf("found node: %s", n.Name)
		}
	})

	stopCh <- struct{}{}
}
