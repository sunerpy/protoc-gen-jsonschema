package test

import (
	"encoding/json"
	"testing"

	protojsonschema "github.com/sunerpy/protoc-gen-jsonschema"
	"github.com/sunerpy/protoc-gen-jsonschema/test/pb"
)

func fieldSchema(t *testing.T, fieldName string) map[string]interface{} {
	t.Helper()
	g := protojsonschema.NewGenerator()
	schema, err := g.GenerateSchema((&pb.TimestampTestMessage{}).ProtoReflect().Descriptor())
	if err != nil {
		t.Fatalf("GenerateSchema failed: %v", err)
	}
	data, err := json.Marshal(schema)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	props, ok := m["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("properties not a map")
	}
	field, ok := props[fieldName].(map[string]interface{})
	if !ok {
		t.Fatalf("field %q not found", fieldName)
	}
	return field
}

func TestFieldOptions_DescriptionAndExample(t *testing.T) {
	created := fieldSchema(t, "createdAt")
	if created["description"] != "Creation timestamp" {
		t.Errorf("expected description from option, got %v", created["description"])
	}
	if created["example"] != "2023-01-01T00:00:00Z" {
		t.Errorf("expected example from option, got %v", created["example"])
	}
}

func TestFieldOptions_PlainStringField(t *testing.T) {
	name := fieldSchema(t, "name")
	if name["type"] != "string" {
		t.Errorf("expected string type for name, got %v", name["type"])
	}
	if name["description"] != "Name field for comparison" {
		t.Errorf("expected name description from option, got %v", name["description"])
	}
}

func TestRequiredOption_PropagatesToSchema(t *testing.T) {
	g := protojsonschema.NewGenerator()
	schema, err := g.GenerateSchema((&pb.TimestampTestMessage{}).ProtoReflect().Descriptor())
	if err != nil {
		t.Fatalf("GenerateSchema failed: %v", err)
	}
	data, _ := json.Marshal(schema)
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	required, ok := m["required"].([]interface{})
	if !ok {
		t.Fatal("expected required array in schema")
	}
	found := false
	for _, r := range required {
		if r == "requiredTimestamp" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected requiredTimestamp in required list, got %v", required)
	}
}

func TestMessageTitleAndDescriptionOptions(t *testing.T) {
	g := protojsonschema.NewGenerator()
	schema, err := g.GenerateSchema((&pb.TimestampTestMessage{}).ProtoReflect().Descriptor())
	if err != nil {
		t.Fatalf("GenerateSchema failed: %v", err)
	}
	if schema["title"] != "Timestamp Test Message" {
		t.Errorf("expected title from message option, got %v", schema["title"])
	}
	if schema["description"] != "Message for testing Timestamp field JSON Schema generation" {
		t.Errorf("expected description from message option, got %v", schema["description"])
	}
}

func TestRepeatedField_IsArray(t *testing.T) {
	events := fieldSchema(t, "eventTimestamps")
	if events["type"] != "array" {
		t.Errorf("expected array type for repeated field, got %v", events["type"])
	}
	if _, ok := events["items"]; !ok {
		t.Error("expected items in array schema")
	}
}
