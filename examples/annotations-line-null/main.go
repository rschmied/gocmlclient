package main

import (
	"context"
	"fmt"
	"os"
	"time"

	gocml "github.com/rschmied/gocmlclient"
	"github.com/rschmied/gocmlclient/pkg/models"
)

func main() {
	baseURL := os.Getenv("CML_BASE_URL")
	token := os.Getenv("CML_TOKEN")
	if baseURL == "" || token == "" {
		fmt.Fprintln(os.Stderr, "set CML_BASE_URL and CML_TOKEN")
		os.Exit(2)
	}

	c, err := gocml.New(baseURL, gocml.WithToken(token))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Mutates server state: creates a lab + a line annotation and then deletes the lab.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	lab, err := c.Lab.Create(ctx, models.LabCreateRequest{
		Title: fmt.Sprintf("example-ann-%d", time.Now().UnixNano()),
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1) //nolint:gocritic
	}
	defer func() {
		_ = c.Lab.Delete(context.Background(), lab.ID)
	}()

	arrow := models.LineStyleArrow
	created, err := c.Annotation.Create(ctx, lab.ID, models.AnnotationCreate{
		Type: models.AnnotationTypeLine,
		Line: &models.LineAnnotation{
			Type:        models.AnnotationTypeLine,
			BorderColor: "#000000",
			BorderStyle: "",
			Color:       "#ffffff",
			Thickness:   1,
			X1:          10,
			Y1:          10,
			X2:          100,
			Y2:          10,
			ZIndex:      0,
			LineStart:   &arrow,
			LineEnd:     &arrow,
		},
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if created.Line == nil {
		fmt.Fprintln(os.Stderr, "create did not return a line annotation")
		os.Exit(1)
	}

	// Clear both ends: PATCH includes line_start/line_end keys with null.
	_, err = c.Annotation.Update(ctx, lab.ID, created.Line.ID, models.AnnotationUpdate{
		Type: models.AnnotationTypeLine,
		Line: &models.LineAnnotationPartial{
			Type:      models.AnnotationTypeLine,
			LineStart: nil,
			LineEnd:   nil,
		},
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("ok")
}
