package cmlclient

import (
	"context"
	"fmt"
	"os"
	"testing"
)

func TestLabLiveServer(t *testing.T) {
	testHost := os.Getenv("TEST_HOST")
	testToken := os.Getenv("TEST_TOKEN")

	if testHost == "" || testToken == "" {
		t.Skip("skipping live server test: TEST_HOST and TEST_TOKEN environment variables must be set")
	}

	ctx := context.Background()
	client := New(testHost, true) // true for insecure TLS
	if client == nil {
		t.Fatal("Failed to create CML client")
	}
	client.apiToken = testToken

	t.Run("CreateLab", func(t *testing.T) {
		lab, err := client.LabCreate(ctx, Lab{Title: "cmlclient test lab"})
		if err != nil {
			t.Fatalf("failed to create lab definitions: %v", err)
		}

		node := &Node{
			LabID:          lab.ID,
			Label:          "node0",
			X:              0,
			Y:              0,
			HideLinks:      false,
			NodeDefinition: "iosv",
			Interfaces: InterfaceList{
				&Interface{
					Slot:       0,
					Label:      "GigabitEthernet0/0",
					MACaddress: "00:55:aa:de:ad:be",
				},
			},
			Tags: []string{"router"},
		}

		node, err = client.NodeCreate(ctx, node)
		if err != nil {
			t.Fatalf("failed to create node: %v", err)
		}
		fmt.Printf("node, %+v %s\n", node, *node.Configuration)

		node, err = client.NodeGet(ctx, node)
		if err != nil {
			t.Fatalf("failed to get node: %v", err)
		}
		fmt.Printf("node, %+v %s\n", node, *node.Configuration)

		lab, err = client.LabGet(ctx, lab.ID, true)
		if err != nil {
			t.Fatalf("failed to get lab: %v", err)
		}
		fmt.Printf("lab, %+vn", lab)

		err = client.LabDestroy(ctx, lab.ID)
		if err != nil {
			t.Fatalf("failed to destroy lab definitions: %v", err)
		}
	})
}
