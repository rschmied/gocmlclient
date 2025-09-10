package services

import (
	"context"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/rschmied/gocmlclient/internal/api"
	"github.com/rschmied/gocmlclient/internal/testutil"
	"github.com/rschmied/gocmlclient/pkg/models"
	"github.com/stretchr/testify/assert"
)

func addLabResponders() {
	httpmock.RegisterResponder("POST", "https://mock/api/v0/labs",
		httpmock.NewStringResponder(200, `{"id":"lab_uuid","state":"DEFINED_ON_CORE","created":"2025-08-26T09:41:36+00:00","modified":"2025-08-26T09:41:36+00:00","lab_title":"this","owner":"00000000-0000-4000-a000-000000000000","owner_username":"admin","effective_permissions":["lab_admin","lab_exec","lab_edit","lab_view"]}`))
	httpmock.RegisterResponder("PATCH", "https://mock/api/v0/labs/lab_uuid",
		httpmock.NewStringResponder(200, `{"id":"lab_uuid","state":"DEFINED_ON_CORE","created":"2025-08-26T09:41:36+00:00","modified":"2025-08-26T09:41:36+00:00","lab_title":"this","owner":"00000000-0000-4000-a000-000000000000","owner_username":"admin","effective_permissions":["lab_admin","lab_exec","lab_edit","lab_view"]}`))
	httpmock.RegisterResponder("DELETE", "https://mock/api/v0/labs/lab_uuid",
		httpmock.NewJsonResponderOrPanic(204, nil))
}

func initLabTest(t *testing.T) (*api.Client, func()) {
	client, cleanup := testutil.NewAPIClient(t)
	if !testutil.IsLiveTesting() {
		addLabResponders()
	}
	return client, cleanup
}

func TestLabCreate(t *testing.T) {
	client, cleanup := initLabTest(t)
	defer cleanup()

	service := NewLabService(client, nil, nil, nil, nil)
	ctx := context.Background()

	lab := models.LabCreateRequest{Title: "this"}
	newLab, err := service.Create(ctx, lab)
	if err != nil {
		testutil.PrettyPrintError(err)
	}
	assert.NoError(t, err)

	err = service.Delete(ctx, newLab.ID)
	assert.NoError(t, err)
}
