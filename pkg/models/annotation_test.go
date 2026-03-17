package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnnotation_Unmarshal_Discriminator(t *testing.T) {
	js := `{"id":"a1","type":"text","border_color":"#000","border_style":"","color":"#fff","thickness":1,"x1":1,"y1":2,"z_index":0,"rotation":0,"text_bold":false,"text_content":"hi","text_font":"sans","text_italic":false,"text_size":12,"text_unit":"px"}`
	var a Annotation
	err := json.Unmarshal([]byte(js), &a)
	assert.NoError(t, err)
	assert.Equal(t, AnnotationTypeText, a.Type)
	if assert.NotNil(t, a.Text) {
		assert.Equal(t, UUID("a1"), a.Text.ID)
		assert.Equal(t, "hi", a.Text.TextContent)
	}
}

func TestAnnotation_Unmarshal_Discriminator_AllTypes(t *testing.T) {
	tests := []struct {
		name string
		js   string
		want AnnotationType
	}{
		{
			name: "rectangle",
			want: AnnotationTypeRectangle,
			js:   `{"id":"r1","type":"rectangle","border_color":"#000","border_style":"","color":"#fff","thickness":1,"x1":1,"y1":2,"x2":3,"y2":4,"z_index":0,"rotation":0,"border_radius":1}`,
		},
		{
			name: "ellipse",
			want: AnnotationTypeEllipse,
			js:   `{"id":"e1","type":"ellipse","border_color":"#000","border_style":"","color":"#fff","thickness":1,"x1":1,"y1":2,"x2":3,"y2":4,"z_index":0,"rotation":0}`,
		},
		{
			name: "line",
			want: AnnotationTypeLine,
			js:   `{"id":"l1","type":"line","border_color":"#000","border_style":"","color":"#fff","thickness":1,"x1":1,"y1":2,"x2":3,"y2":4,"z_index":0,"line_start":null,"line_end":null}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var a Annotation
			err := json.Unmarshal([]byte(tt.js), &a)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, a.Type)
			switch tt.want {
			case AnnotationTypeRectangle:
				assert.NotNil(t, a.Rectangle)
				assert.Equal(t, UUID("r1"), a.Rectangle.ID)
			case AnnotationTypeEllipse:
				assert.NotNil(t, a.Ellipse)
				assert.Equal(t, UUID("e1"), a.Ellipse.ID)
			case AnnotationTypeLine:
				assert.NotNil(t, a.Line)
				assert.Equal(t, UUID("l1"), a.Line.ID)
			}
		})
	}
}

func TestAnnotationCreate_Marshal(t *testing.T) {
	in := AnnotationCreate{Type: AnnotationTypeText, Text: &TextAnnotation{Type: AnnotationTypeText, BorderColor: "#000", BorderStyle: "", Color: "#fff", Thickness: 1, X1: 1, Y1: 2, ZIndex: 0, Rotation: 0, TextBold: false, TextContent: "hi", TextFont: "sans", TextItalic: false, TextSize: 12, TextUnit: "px"}}
	b, err := json.Marshal(in)
	assert.NoError(t, err)
	var m map[string]any
	err = json.Unmarshal(b, &m)
	assert.NoError(t, err)
	assert.Equal(t, "text", m["type"])
}

func TestAnnotationUpdate_Marshal_RequiresTypeInStruct(t *testing.T) {
	content := "updated"
	in := AnnotationUpdate{Type: AnnotationTypeText, Text: &TextAnnotationPartial{Type: AnnotationTypeText, TextContent: &content}}
	b, err := json.Marshal(in)
	assert.NoError(t, err)
	var m map[string]any
	err = json.Unmarshal(b, &m)
	assert.NoError(t, err)
	assert.Equal(t, "text", m["type"])
	assert.Equal(t, "updated", m["text_content"])
}

func TestAnnotation_MarshalJSON_AllTypes(t *testing.T) {
	// text
	a := Annotation{Type: AnnotationTypeText, Text: &TextAnnotationResponse{ID: "a1", TextAnnotation: TextAnnotation{Type: AnnotationTypeText, BorderColor: "#000", BorderStyle: "", Color: "#fff", Thickness: 1, X1: 1, Y1: 2, ZIndex: 0, Rotation: 0, TextBold: false, TextContent: "hi", TextFont: "sans", TextItalic: false, TextSize: 12, TextUnit: "px"}}}
	b, err := json.Marshal(a)
	assert.NoError(t, err)
	var m map[string]any
	assert.NoError(t, json.Unmarshal(b, &m))
	assert.Equal(t, "text", m["type"])

	// rectangle
	a = Annotation{Type: AnnotationTypeRectangle, Rectangle: &RectangleAnnotationResponse{ID: "r1", RectangleAnnotation: RectangleAnnotation{Type: AnnotationTypeRectangle, BorderColor: "#000", BorderStyle: "", Color: "#fff", Thickness: 1, X1: 1, Y1: 2, X2: 3, Y2: 4, ZIndex: 0, Rotation: 0, BorderRadius: 1}}}
	b, err = json.Marshal(a)
	assert.NoError(t, err)
	assert.NoError(t, json.Unmarshal(b, &m))
	assert.Equal(t, "rectangle", m["type"])

	// ellipse
	a = Annotation{Type: AnnotationTypeEllipse, Ellipse: &EllipseAnnotationResponse{ID: "e1", EllipseAnnotation: EllipseAnnotation{Type: AnnotationTypeEllipse, BorderColor: "#000", BorderStyle: "", Color: "#fff", Thickness: 1, X1: 1, Y1: 2, X2: 3, Y2: 4, ZIndex: 0, Rotation: 0}}}
	b, err = json.Marshal(a)
	assert.NoError(t, err)
	assert.NoError(t, json.Unmarshal(b, &m))
	assert.Equal(t, "ellipse", m["type"])

	// line
	a = Annotation{Type: AnnotationTypeLine, Line: &LineAnnotationResponse{ID: "l1", LineAnnotation: LineAnnotation{Type: AnnotationTypeLine, BorderColor: "#000", BorderStyle: "", Color: "#fff", Thickness: 1, X1: 1, Y1: 2, X2: 3, Y2: 4, ZIndex: 0, LineStart: nil, LineEnd: nil}}}
	b, err = json.Marshal(a)
	assert.NoError(t, err)
	assert.NoError(t, json.Unmarshal(b, &m))
	assert.Equal(t, "line", m["type"])
}

func TestAnnotation_MarshalJSON_Errors(t *testing.T) {
	_, err := json.Marshal(Annotation{Type: AnnotationTypeText})
	assert.Error(t, err)
	_, err = json.Marshal(Annotation{Type: "nope"})
	assert.Error(t, err)
}

func TestAnnotationCreate_Marshal_AllTypes(t *testing.T) {
	tests := []struct {
		name string
		in   AnnotationCreate
		want string
	}{
		{
			name: "line",
			want: "line",
			in:   AnnotationCreate{Type: AnnotationTypeLine, Line: &LineAnnotation{Type: AnnotationTypeLine, BorderColor: "#000", BorderStyle: "", Color: "#fff", Thickness: 1, X1: 1, Y1: 2, X2: 3, Y2: 4, ZIndex: 0, LineStart: nil, LineEnd: nil}},
		},
		{
			name: "rectangle",
			want: "rectangle",
			in:   AnnotationCreate{Type: AnnotationTypeRectangle, Rectangle: &RectangleAnnotation{Type: AnnotationTypeRectangle, BorderColor: "#000", BorderStyle: "", Color: "#fff", Thickness: 1, X1: 1, Y1: 2, X2: 3, Y2: 4, ZIndex: 0, Rotation: 0, BorderRadius: 1}},
		},
		{
			name: "ellipse",
			want: "ellipse",
			in:   AnnotationCreate{Type: AnnotationTypeEllipse, Ellipse: &EllipseAnnotation{Type: AnnotationTypeEllipse, BorderColor: "#000", BorderStyle: "", Color: "#fff", Thickness: 1, X1: 1, Y1: 2, X2: 3, Y2: 4, ZIndex: 0, Rotation: 0}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := json.Marshal(tt.in)
			assert.NoError(t, err)
			var m map[string]any
			assert.NoError(t, json.Unmarshal(b, &m))
			assert.Equal(t, tt.want, m["type"])
		})
	}
}

func TestAnnotationUpdate_Marshal_RectangleAndLine(t *testing.T) {
	x2 := 9.0
	upd := AnnotationUpdate{Type: AnnotationTypeRectangle, Rectangle: &RectangleAnnotationPartial{Type: AnnotationTypeRectangle, X2: &x2}}
	b, err := json.Marshal(upd)
	assert.NoError(t, err)
	var m map[string]any
	assert.NoError(t, json.Unmarshal(b, &m))
	assert.Equal(t, "rectangle", m["type"])
	assert.Equal(t, float64(9), m["x2"])

	color := "#abc"
	upd = AnnotationUpdate{Type: AnnotationTypeLine, Line: &LineAnnotationPartial{Type: AnnotationTypeLine, Color: &color}}
	b, err = json.Marshal(upd)
	assert.NoError(t, err)
	assert.NoError(t, json.Unmarshal(b, &m))
	assert.Equal(t, "line", m["type"])
	assert.Equal(t, "#abc", m["color"])
}

func TestAnnotationUpdate_Marshal_Errors(t *testing.T) {
	_, err := json.Marshal(AnnotationUpdate{Type: AnnotationTypeEllipse})
	assert.Error(t, err)
	_, err = json.Marshal(AnnotationUpdate{Type: "nope"})
	assert.Error(t, err)
}
