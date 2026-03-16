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
