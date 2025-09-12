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

func addLabsGetResponders() {
	httpmock.RegisterResponder("GET", "https://mock/api/v0/labs",
		httpmock.NewStringResponder(200, `["uuid-1", "uuid-2", "uuid-3"]`))
}

func addLabCreateResponders() {
	httpmock.RegisterResponder("POST", "https://mock/api/v0/labs",
		httpmock.NewStringResponder(200, `{"id":"lab_uuid","state":"DEFINED_ON_CORE","created":"2025-08-26T09:41:36+00:00","modified":"2025-08-26T09:41:36+00:00","lab_title":"this","owner":"00000000-0000-4000-a000-000000000000","owner_username":"admin","effective_permissions":["lab_admin","lab_exec","lab_edit","lab_view"]}`))
	httpmock.RegisterResponder("PATCH", "https://mock/api/v0/labs/lab_uuid",
		httpmock.NewStringResponder(200, `{"id":"lab_uuid","state":"DEFINED_ON_CORE","created":"2025-08-26T09:41:36+00:00","modified":"2025-08-26T09:41:36+00:00","lab_title":"this","owner":"00000000-0000-4000-a000-000000000000","owner_username":"admin","effective_permissions":["lab_admin","lab_exec","lab_edit","lab_view"]}`))
	httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/lab_uuid",
		httpmock.NewStringResponder(200, `{"id":"lab_uuid","state":"DEFINED_ON_CORE","created":"2025-08-26T09:41:36+00:00","modified":"2025-08-26T09:41:36+00:00","lab_title":"this","owner":"00000000-0000-4000-a000-000000000000","owner_username":"admin","effective_permissions":["lab_admin","lab_exec","lab_edit","lab_view"]}`))
	httpmock.RegisterResponder("PUT", "https://mock/api/v0/labs/lab_uuid/state/start",
		httpmock.NewStringResponder(200, `{"id":"lab_uuid","state":"QUEUED","created":"2025-08-26T09:41:36+00:00","modified":"2025-08-26T09:41:36+00:00","lab_title":"this","owner":"00000000-0000-4000-a000-000000000000","owner_username":"admin","effective_permissions":["lab_admin","lab_exec","lab_edit","lab_view"]}`))
	httpmock.RegisterResponder("PUT", "https://mock/api/v0/labs/lab_uuid/state/stop",
		httpmock.NewStringResponder(200, `{"id":"lab_uuid","state":"STOPPED","created":"2025-08-26T09:41:36+00:00","modified":"2025-08-26T09:41:36+00:00","lab_title":"this","owner":"00000000-0000-4000-a000-000000000000","owner_username":"admin","effective_permissions":["lab_admin","lab_exec","lab_edit","lab_view"]}`))
	httpmock.RegisterResponder("PUT", "https://mock/api/v0/labs/lab_uuid/state/wipe",
		httpmock.NewStringResponder(200, `{"id":"lab_uuid","state":"DEFINED_ON_CORE","created":"2025-08-26T09:41:36+00:00","modified":"2025-08-26T09:41:36+00:00","lab_title":"this","owner":"00000000-0000-4000-a000-000000000000","owner_username":"admin","effective_permissions":["lab_admin","lab_exec","lab_edit","lab_view"]}`))
	httpmock.RegisterResponder("DELETE", "https://mock/api/v0/labs/lab_uuid",
		httpmock.NewJsonResponderOrPanic(204, nil))
}

func initLabTest(t *testing.T, responders func()) (*api.Client, func()) {
	client, cleanup := testutil.NewAPIClient(t)
	if !testutil.IsLiveTesting() {
		responders()
	}
	return client, cleanup
}

func TestLabs(t *testing.T) {
	client, cleanup := initLabTest(t, addLabsGetResponders)
	defer cleanup()

	service := NewLabService(client, nil, nil, nil)
	ctx := context.Background()

	labs, err := service.Labs(ctx, true)
	if err != nil {
		testutil.PrettyPrintError(err)
	}
	assert.NoError(t, err)
	assert.Len(t, labs, 3)
}

func TestLabCreate(t *testing.T) {
	client, cleanup := initLabTest(t, addLabCreateResponders)
	defer cleanup()

	service := NewLabService(client, nil, nil, nil)
	ctx := context.Background()

	lab := models.LabCreateRequest{Title: "this"}
	newLab, err := service.Create(ctx, lab)
	if err != nil {
		testutil.PrettyPrintError(err)
	}
	assert.NoError(t, err)

	newLab, err = service.GetByID(ctx, newLab.ID, false)
	assert.NoError(t, err)

	err = service.Start(ctx, newLab.ID)
	assert.NoError(t, err)

	err = service.Stop(ctx, newLab.ID)
	assert.NoError(t, err)

	err = service.Wipe(ctx, newLab.ID)
	assert.NoError(t, err)

	err = service.Delete(ctx, newLab.ID)
	assert.NoError(t, err)
}
