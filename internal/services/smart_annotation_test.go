package services

import (
	"context"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"

	"github.com/rschmied/gocmlclient/internal/testutil"
	"github.com/rschmied/gocmlclient/pkg/models"
)

func TestSmartAnnotationService_ListGetUpdate(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - test expects specific mock data")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/lab-1/smart_annotations",
		httpmock.NewStringResponder(200, `[{"id":"s1","label":"x","is_on":true}]`))

	httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/lab-1/smart_annotations/s1",
		httpmock.NewStringResponder(200, `{"id":"s1","label":"x","is_on":true}`))

	httpmock.RegisterResponder("PATCH", "https://mock/api/v0/labs/lab-1/smart_annotations/s1",
		httpmock.NewStringResponder(200, `{"id":"s1","label":"y","is_on":false}`))

	svc := NewSmartAnnotationService(client)
	ctx := context.Background()

	list, err := svc.List(ctx, "lab-1")
	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, models.UUID("s1"), list[0].ID)

	got, err := svc.Get(ctx, "lab-1", "s1")
	assert.NoError(t, err)
	assert.Equal(t, models.UUID("s1"), got.ID)

	label := "y"
	isOn := false
	upd := models.SmartAnnotationUpdate{Label: &label, IsOn: &isOn}
	updated, err := svc.Update(ctx, "lab-1", "s1", upd)
	assert.NoError(t, err)
	if assert.NotNil(t, updated.Label) {
		assert.Equal(t, "y", *updated.Label)
	}
}
