// Package models provides the models for Cisco Modeling Labs
// here: annotation related types
package models

import (
	"encoding/json"
	"fmt"
)

// AnnotationType is the discriminator for classic annotations.
type AnnotationType string

const (
	AnnotationTypeText      AnnotationType = "text"
	AnnotationTypeRectangle AnnotationType = "rectangle"
	AnnotationTypeEllipse   AnnotationType = "ellipse"
	AnnotationTypeLine      AnnotationType = "line"
)

// BorderStyle matches OpenAPI `BorderStyle`.
// Values observed in schema: "", "2,2", "4,2".
type BorderStyle string

// LineStyle matches OpenAPI `LineStyle`.
// Values observed in schema: "arrow", "square", "circle".
type LineStyle string

// Annotation is a discriminated union wrapper.
// Exactly one of Text/Rectangle/Ellipse/Line is set after unmarshaling.
type Annotation struct {
	Type      AnnotationType
	Text      *TextAnnotationResponse
	Rectangle *RectangleAnnotationResponse
	Ellipse   *EllipseAnnotationResponse
	Line      *LineAnnotationResponse
}

func (a *Annotation) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	var probe struct {
		Type AnnotationType `json:"type"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return err
	}
	a.Type = probe.Type

	switch probe.Type {
	case AnnotationTypeText:
		var v TextAnnotationResponse
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		a.Text, a.Rectangle, a.Ellipse, a.Line = &v, nil, nil, nil
	case AnnotationTypeRectangle:
		var v RectangleAnnotationResponse
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		a.Text, a.Rectangle, a.Ellipse, a.Line = nil, &v, nil, nil
	case AnnotationTypeEllipse:
		var v EllipseAnnotationResponse
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		a.Text, a.Rectangle, a.Ellipse, a.Line = nil, nil, &v, nil
	case AnnotationTypeLine:
		var v LineAnnotationResponse
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		a.Text, a.Rectangle, a.Ellipse, a.Line = nil, nil, nil, &v
	default:
		return fmt.Errorf("unknown annotation type %q", probe.Type)
	}
	return nil
}

func (a Annotation) MarshalJSON() ([]byte, error) {
	switch a.Type {
	case AnnotationTypeText:
		if a.Text == nil {
			return nil, fmt.Errorf("annotation type %q but Text is nil", a.Type)
		}
		return json.Marshal(a.Text)
	case AnnotationTypeRectangle:
		if a.Rectangle == nil {
			return nil, fmt.Errorf("annotation type %q but Rectangle is nil", a.Type)
		}
		return json.Marshal(a.Rectangle)
	case AnnotationTypeEllipse:
		if a.Ellipse == nil {
			return nil, fmt.Errorf("annotation type %q but Ellipse is nil", a.Type)
		}
		return json.Marshal(a.Ellipse)
	case AnnotationTypeLine:
		if a.Line == nil {
			return nil, fmt.Errorf("annotation type %q but Line is nil", a.Type)
		}
		return json.Marshal(a.Line)
	default:
		return nil, fmt.Errorf("unknown annotation type %q", a.Type)
	}
}

// AnnotationCreate is the request payload for creating an annotation.
// Exactly one of Text/Rectangle/Ellipse/Line should be set.
type AnnotationCreate struct {
	Type      AnnotationType
	Text      *TextAnnotation
	Rectangle *RectangleAnnotation
	Ellipse   *EllipseAnnotation
	Line      *LineAnnotation
}

func (a AnnotationCreate) MarshalJSON() ([]byte, error) {
	switch a.Type {
	case AnnotationTypeText:
		if a.Text == nil {
			return nil, fmt.Errorf("annotation type %q but Text is nil", a.Type)
		}
		return json.Marshal(a.Text)
	case AnnotationTypeRectangle:
		if a.Rectangle == nil {
			return nil, fmt.Errorf("annotation type %q but Rectangle is nil", a.Type)
		}
		return json.Marshal(a.Rectangle)
	case AnnotationTypeEllipse:
		if a.Ellipse == nil {
			return nil, fmt.Errorf("annotation type %q but Ellipse is nil", a.Type)
		}
		return json.Marshal(a.Ellipse)
	case AnnotationTypeLine:
		if a.Line == nil {
			return nil, fmt.Errorf("annotation type %q but Line is nil", a.Type)
		}
		return json.Marshal(a.Line)
	default:
		return nil, fmt.Errorf("unknown annotation type %q", a.Type)
	}
}

// AnnotationUpdate is the request payload for patching an annotation.
// OpenAPI requires `type` and accepts partial fields.
type AnnotationUpdate struct {
	Type      AnnotationType
	Text      *TextAnnotationPartial
	Rectangle *RectangleAnnotationPartial
	Ellipse   *EllipseAnnotationPartial
	Line      *LineAnnotationPartial
}

func (a AnnotationUpdate) MarshalJSON() ([]byte, error) {
	switch a.Type {
	case AnnotationTypeText:
		if a.Text == nil {
			return nil, fmt.Errorf("annotation type %q but Text is nil", a.Type)
		}
		return json.Marshal(a.Text)
	case AnnotationTypeRectangle:
		if a.Rectangle == nil {
			return nil, fmt.Errorf("annotation type %q but Rectangle is nil", a.Type)
		}
		return json.Marshal(a.Rectangle)
	case AnnotationTypeEllipse:
		if a.Ellipse == nil {
			return nil, fmt.Errorf("annotation type %q but Ellipse is nil", a.Type)
		}
		return json.Marshal(a.Ellipse)
	case AnnotationTypeLine:
		if a.Line == nil {
			return nil, fmt.Errorf("annotation type %q but Line is nil", a.Type)
		}
		return json.Marshal(a.Line)
	default:
		return nil, fmt.Errorf("unknown annotation type %q", a.Type)
	}
}

type TextAnnotation struct {
	Rotation    float64        `json:"rotation"`
	Type        AnnotationType `json:"type"`
	BorderColor string         `json:"border_color"`
	BorderStyle BorderStyle    `json:"border_style"`
	Color       string         `json:"color"`
	Thickness   float64        `json:"thickness"`
	X1          float64        `json:"x1"`
	Y1          float64        `json:"y1"`
	ZIndex      float64        `json:"z_index"`
	TextBold    bool           `json:"text_bold"`
	TextContent string         `json:"text_content"`
	TextFont    string         `json:"text_font"`
	TextItalic  bool           `json:"text_italic"`
	TextSize    float64        `json:"text_size"`
	TextUnit    string         `json:"text_unit"`
}

type TextAnnotationPartial struct {
	Rotation    *float64       `json:"rotation,omitempty"`
	Type        AnnotationType `json:"type"`
	BorderColor *string        `json:"border_color,omitempty"`
	BorderStyle *BorderStyle   `json:"border_style,omitempty"`
	Color       *string        `json:"color,omitempty"`
	Thickness   *float64       `json:"thickness,omitempty"`
	X1          *float64       `json:"x1,omitempty"`
	Y1          *float64       `json:"y1,omitempty"`
	ZIndex      *float64       `json:"z_index,omitempty"`
	TextBold    *bool          `json:"text_bold,omitempty"`
	TextContent *string        `json:"text_content,omitempty"`
	TextFont    *string        `json:"text_font,omitempty"`
	TextItalic  *bool          `json:"text_italic,omitempty"`
	TextSize    *float64       `json:"text_size,omitempty"`
	TextUnit    *string        `json:"text_unit,omitempty"`
}

type TextAnnotationResponse struct {
	ID UUID `json:"id"`
	TextAnnotation
}

type RectangleAnnotation struct {
	X2           float64        `json:"x2"`
	Y2           float64        `json:"y2"`
	Rotation     float64        `json:"rotation"`
	Type         AnnotationType `json:"type"`
	BorderColor  string         `json:"border_color"`
	BorderStyle  BorderStyle    `json:"border_style"`
	Color        string         `json:"color"`
	Thickness    float64        `json:"thickness"`
	X1           float64        `json:"x1"`
	Y1           float64        `json:"y1"`
	ZIndex       float64        `json:"z_index"`
	BorderRadius float64        `json:"border_radius"`
}

type RectangleAnnotationPartial struct {
	X2           *float64       `json:"x2,omitempty"`
	Y2           *float64       `json:"y2,omitempty"`
	Rotation     *float64       `json:"rotation,omitempty"`
	Type         AnnotationType `json:"type"`
	BorderColor  *string        `json:"border_color,omitempty"`
	BorderStyle  *BorderStyle   `json:"border_style,omitempty"`
	Color        *string        `json:"color,omitempty"`
	Thickness    *float64       `json:"thickness,omitempty"`
	X1           *float64       `json:"x1,omitempty"`
	Y1           *float64       `json:"y1,omitempty"`
	ZIndex       *float64       `json:"z_index,omitempty"`
	BorderRadius *float64       `json:"border_radius,omitempty"`
}

type RectangleAnnotationResponse struct {
	ID UUID `json:"id"`
	RectangleAnnotation
}

type EllipseAnnotation struct {
	X2          float64        `json:"x2"`
	Y2          float64        `json:"y2"`
	Rotation    float64        `json:"rotation"`
	Type        AnnotationType `json:"type"`
	BorderColor string         `json:"border_color"`
	BorderStyle BorderStyle    `json:"border_style"`
	Color       string         `json:"color"`
	Thickness   float64        `json:"thickness"`
	X1          float64        `json:"x1"`
	Y1          float64        `json:"y1"`
	ZIndex      float64        `json:"z_index"`
}

type EllipseAnnotationPartial struct {
	X2          *float64       `json:"x2,omitempty"`
	Y2          *float64       `json:"y2,omitempty"`
	Rotation    *float64       `json:"rotation,omitempty"`
	Type        AnnotationType `json:"type"`
	BorderColor *string        `json:"border_color,omitempty"`
	BorderStyle *BorderStyle   `json:"border_style,omitempty"`
	Color       *string        `json:"color,omitempty"`
	Thickness   *float64       `json:"thickness,omitempty"`
	X1          *float64       `json:"x1,omitempty"`
	Y1          *float64       `json:"y1,omitempty"`
	ZIndex      *float64       `json:"z_index,omitempty"`
}

type EllipseAnnotationResponse struct {
	ID UUID `json:"id"`
	EllipseAnnotation
}

type LineAnnotation struct {
	X2          float64        `json:"x2"`
	Y2          float64        `json:"y2"`
	Type        AnnotationType `json:"type"`
	BorderColor string         `json:"border_color"`
	BorderStyle BorderStyle    `json:"border_style"`
	Color       string         `json:"color"`
	Thickness   float64        `json:"thickness"`
	X1          float64        `json:"x1"`
	Y1          float64        `json:"y1"`
	ZIndex      float64        `json:"z_index"`
	// line_start/line_end are required by the schema but may be null.
	LineStart *LineStyle `json:"line_start"`
	LineEnd   *LineStyle `json:"line_end"`
}

type LineAnnotationPartial struct {
	X2          *float64       `json:"x2,omitempty"`
	Y2          *float64       `json:"y2,omitempty"`
	Type        AnnotationType `json:"type"`
	BorderColor *string        `json:"border_color,omitempty"`
	BorderStyle *BorderStyle   `json:"border_style,omitempty"`
	Color       *string        `json:"color,omitempty"`
	Thickness   *float64       `json:"thickness,omitempty"`
	X1          *float64       `json:"x1,omitempty"`
	Y1          *float64       `json:"y1,omitempty"`
	ZIndex      *float64       `json:"z_index,omitempty"`
	LineStart   *LineStyle     `json:"line_start,omitempty"`
	LineEnd     *LineStyle     `json:"line_end,omitempty"`
}

type LineAnnotationResponse struct {
	ID UUID `json:"id"`
	LineAnnotation
}

type SmartAnnotation struct {
	ID            UUID         `json:"id"`
	BorderColor   *string      `json:"border_color,omitempty"`
	BorderStyle   *BorderStyle `json:"border_style,omitempty"`
	FillColor     *string      `json:"fill_color,omitempty"`
	GroupDistance *int         `json:"group_distance,omitempty"`
	IsOn          *bool        `json:"is_on,omitempty"`
	Label         *string      `json:"label,omitempty"`
	Padding       *int         `json:"padding,omitempty"`
	Tag           *string      `json:"tag,omitempty"`
	TagOffsetX    *int         `json:"tag_offset_x,omitempty"`
	TagOffsetY    *int         `json:"tag_offset_y,omitempty"`
	TagSize       *int         `json:"tag_size,omitempty"`
	Thickness     *int         `json:"thickness,omitempty"`
	ZIndex        *int         `json:"z_index,omitempty"`
}

type SmartAnnotationUpdate struct {
	BorderColor   *string      `json:"border_color,omitempty"`
	BorderStyle   *BorderStyle `json:"border_style,omitempty"`
	FillColor     *string      `json:"fill_color,omitempty"`
	GroupDistance *int         `json:"group_distance,omitempty"`
	IsOn          *bool        `json:"is_on,omitempty"`
	Label         *string      `json:"label,omitempty"`
	Padding       *int         `json:"padding,omitempty"`
	Tag           *string      `json:"tag,omitempty"`
	TagOffsetX    *int         `json:"tag_offset_x,omitempty"`
	TagOffsetY    *int         `json:"tag_offset_y,omitempty"`
	TagSize       *int         `json:"tag_size,omitempty"`
	Thickness     *int         `json:"thickness,omitempty"`
	ZIndex        *int         `json:"z_index,omitempty"`
}
