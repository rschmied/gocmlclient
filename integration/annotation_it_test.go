//go:build integration

package integration

import (
	"testing"

	"github.com/rschmied/gocmlclient/pkg/models"
)

func TestIntegration_Annotations(t *testing.T) {
	cfg := LoadConfigFromEnv()
	c := newClient(t, cfg)
	wireClientServices(c)
	requireReady(t, c, cfg)

	ctx, cancel := testContext(t, cfg)
	defer cancel()

	if !cfg.AllowMutations {
		t.Skip("set CML_IT_ALLOW_MUTATIONS=1 to run")
	}

	lab := createTempLab(t, c, cfg, "it-ann")

	// Create a simple text annotation.
	create := models.AnnotationCreate{Type: models.AnnotationTypeText, Text: &models.TextAnnotation{Type: models.AnnotationTypeText, BorderColor: "#000000", BorderStyle: "", Color: "#ffffff", Thickness: 1, X1: 10, Y1: 10, ZIndex: 0, Rotation: 0, TextBold: false, TextContent: "hello", TextFont: "sans", TextItalic: false, TextSize: 12, TextUnit: "px"}}
	ann, err := c.Annotation.Create(ctx, lab.ID, create)
	if err != nil {
		requireNoErrorOrSkipStatus(t, err, 404, 400, 403)
	}

	// List
	list, err := c.Annotation.List(ctx, lab.ID)
	if err != nil {
		requireNoErrorOrSkipStatus(t, err, 404)
	}
	_ = list

	// Get
	if ann.Text != nil {
		_, err = c.Annotation.Get(ctx, lab.ID, ann.Text.ID)
		if err != nil {
			requireNoErrorOrSkipStatus(t, err, 404)
		}

		// Patch
		updated := "hello-updated"
		upd := models.AnnotationUpdate{Type: models.AnnotationTypeText, Text: &models.TextAnnotationPartial{Type: models.AnnotationTypeText, TextContent: &updated}}
		_, err = c.Annotation.Update(ctx, lab.ID, ann.Text.ID, upd)
		if err != nil {
			requireNoErrorOrSkipStatus(t, err, 404, 400)
		}

		// Delete
		err = c.Annotation.Delete(ctx, lab.ID, ann.Text.ID)
		if err != nil {
			requireNoErrorOrSkipStatus(t, err, 404, 400)
		}
	}
}
