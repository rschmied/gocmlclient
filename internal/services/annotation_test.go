package services

import (
	"context"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/rschmied/gocmlclient/internal/testutil"
	"github.com/rschmied/gocmlclient/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestAnnotationService_CRUD(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - test expects specific mock data")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	// list
	httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/lab-1/annotations",
		httpmock.NewStringResponder(200, `[
			{"id":"a1","type":"text","border_color":"#000","border_style":"","color":"#fff","thickness":1,"x1":1,"y1":2,"z_index":0,"rotation":0,"text_bold":false,"text_content":"hi","text_font":"sans","text_italic":false,"text_size":12,"text_unit":"px"}
		]`))

	// create
	httpmock.RegisterResponder("POST", "https://mock/api/v0/labs/lab-1/annotations",
		httpmock.NewStringResponder(200, `{"id":"a2","type":"text","border_color":"#000","border_style":"","color":"#fff","thickness":1,"x1":1,"y1":2,"z_index":0,"rotation":0,"text_bold":false,"text_content":"hi","text_font":"sans","text_italic":false,"text_size":12,"text_unit":"px"}`))

	// get
	httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/lab-1/annotations/a2",
		httpmock.NewStringResponder(200, `{"id":"a2","type":"text","border_color":"#000","border_style":"","color":"#fff","thickness":1,"x1":1,"y1":2,"z_index":0,"rotation":0,"text_bold":false,"text_content":"hi","text_font":"sans","text_italic":false,"text_size":12,"text_unit":"px"}`))

	// patch
	httpmock.RegisterResponder("PATCH", "https://mock/api/v0/labs/lab-1/annotations/a2",
		httpmock.NewStringResponder(200, `{"id":"a2","type":"text","border_color":"#000","border_style":"","color":"#fff","thickness":1,"x1":1,"y1":2,"z_index":0,"rotation":0,"text_bold":false,"text_content":"updated","text_font":"sans","text_italic":false,"text_size":12,"text_unit":"px"}`))

	// delete
	httpmock.RegisterResponder("DELETE", "https://mock/api/v0/labs/lab-1/annotations/a2",
		httpmock.NewStringResponder(204, ``))

	svc := NewAnnotationService(client)
	ctx := context.Background()

	list, err := svc.List(ctx, "lab-1")
	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, models.AnnotationTypeText, list[0].Type)

	create := models.AnnotationCreate{Type: models.AnnotationTypeText, Text: &models.TextAnnotation{Type: models.AnnotationTypeText, BorderColor: "#000", BorderStyle: "", Color: "#fff", Thickness: 1, X1: 1, Y1: 2, ZIndex: 0, Rotation: 0, TextBold: false, TextContent: "hi", TextFont: "sans", TextItalic: false, TextSize: 12, TextUnit: "px"}}
	created, err := svc.Create(ctx, "lab-1", create)
	assert.NoError(t, err)
	assert.Equal(t, models.AnnotationTypeText, created.Type)

	fetched, err := svc.Get(ctx, "lab-1", "a2")
	assert.NoError(t, err)
	assert.Equal(t, models.UUID("a2"), fetched.Text.ID)

	updatedText := "updated"
	upd := models.AnnotationUpdate{Type: models.AnnotationTypeText, Text: &models.TextAnnotationPartial{Type: models.AnnotationTypeText, TextContent: &updatedText}}
	updated, err := svc.Update(ctx, "lab-1", "a2", upd)
	assert.NoError(t, err)
	assert.Equal(t, "updated", updated.Text.TextContent)

	err = svc.Delete(ctx, "lab-1", "a2")
	assert.NoError(t, err)
}
