package services

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/rschmied/gocmlclient/internal/testutil"
	"github.com/rschmied/gocmlclient/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestLabCreate(t *testing.T) {
	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	if !testutil.IsLiveTesting() {
		httpmock.RegisterResponder("POST", "https://mock/api/v0/labs",
			httpmock.NewStringResponder(200, `{"id":"lab_uuid","state":"DEFINED_ON_CORE","created":"2025-08-26T09:41:36+00:00","modified":"2025-08-26T09:41:36+00:00","lab_title":"this","owner":"00000000-0000-4000-a000-000000000000","owner_username":"admin","effective_permissions":["lab_admin","lab_exec","lab_edit","lab_view"]}`))
		httpmock.RegisterResponder("DELETE", "https://mock/api/v0/labs/lab_uuid",
			httpmock.NewJsonResponderOrPanic(204, nil))
	}

	service := NewLabService(client)
	ctx := context.Background()

	lab := models.Lab{Title: "this"}
	newLab, err := service.Create(ctx, lab)
	if err != nil {
		testutil.PrettyPrintError(err)
	}
	assert.NoError(t, err)
	json.NewEncoder(os.Stdout).Encode(newLab)

	err = service.Delete(ctx, newLab.ID)
	assert.NoError(t, err)
}
