package services_test

import (
	"context"
	"log/slog"
	"net/http"
	"testing"

	gocml "github.com/rschmied/gocmlclient"
	"github.com/rschmied/gocmlclient/pkg/client"

	"github.com/jarcoal/httpmock"
)

func setupMocks() *http.Client {
	// Activate httpmock to start intercepting requests.
	testClient := &http.Client{}
	httpmock.ActivateNonDefault(testClient)

	// 1. Mock the user endpoint response.
	// The handler function receives the request and returns a mock response.
	httpmock.RegisterResponder("GET", "https://mock/api/v0/system_information",
		httpmock.NewJsonResponderOrPanic(200, map[string]string{
			"id":   "123",
			"name": "John Doe",
		}),
	)

	// 2. Mock the posts endpoint response.
	// The handler for the second API call.
	// httpmock.RegisterResponder("GET", "https://api.service.com/posts?userID=123",
	// 	httpmock.NewJsonResponderOrPanic(200, []map[string]string{
	// 		{"id": "post-1", "title": "First Post"},
	// 		{"id": "post-2", "title": "Second Post"},
	// 	}),
	// )
	return testClient
}

func teardownMocks(client *http.Client) {
	// Deactivate httpmock to stop intercepting requests.
	// httpmock.DeactivateAndReset()
	httpmock.DeactivateNonDefault(client)
	httpmock.Reset()
}

func TestGetUserAndPosts(t *testing.T) {
	testClient := setupMocks()
	defer teardownMocks(testClient)

	// Assuming your client constructor takes an http.Client.
	c, err := gocml.New(
		"https://mock",
		client.WithHTTPClient(testClient),
		// client.WithInsecureTLS(),
		// client.WithUsernamePassword(username, password),
		client.WithToken("sometoken"),
		// client.WithLogger(logger),
	)
	if err != nil {
		slog.Error("new", "err", err)
		return
	}

	ctx := context.Background()
	// Call the function you want to test.
	err = c.System.Ready(ctx)
	// Assert that there were no errors.
	t.Errorf("Expected 1 HTTP call, but got %d", httpmock.GetTotalCallCount())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	t.Errorf("Expected 1 HTTP call, but got %d", httpmock.GetTotalCallCount())
	// // Assert the returned user data is what we mocked.
	// if user.ID != "123" {
	// 	t.Errorf("Expected user ID '123', got '%s'", user.ID)
	// }
	//
	// // Assert the returned posts data is what we mocked.
	// if len(posts) != 2 {
	// 	t.Fatalf("Expected 2 posts, got %d", len(posts))
	// }
}
