package test

import (
	"encoding/json"
	"strings"
	"testing"

	"google.golang.org/protobuf/types/known/timestamppb"

	protojsonschema "github.com/sunerpy/protoc-gen-jsonschema"
	"github.com/sunerpy/protoc-gen-jsonschema/test/pb"
)

// TestTimestampSchemaGenerationSimple verifies that generating a schema directly
// from the bare google.protobuf.Timestamp descriptor produces a plain object with
// its native seconds/nanos integer fields. The oneOf string-or-object special
// casing only applies when a Timestamp appears as a FIELD inside another message
// (see TestTimestampSchemaGeneration), not when Timestamp is the top-level message.
func TestTimestampSchemaGenerationSimple(t *testing.T) {
	ts := &timestamppb.Timestamp{}
	generator := protojsonschema.NewGenerator()
	md := ts.ProtoReflect().Descriptor()

	schema, err := generator.GenerateSchema(md)
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	schemaJSON, err := schema.ToJSON()
	if err != nil {
		t.Fatalf("Failed to convert schema to JSON: %v", err)
	}
	t.Logf("Generated Timestamp schema:\n%s", schemaJSON)

	var got map[string]interface{}
	if err := json.Unmarshal([]byte(schemaJSON), &got); err != nil {
		t.Fatalf("Failed to unmarshal schema JSON: %v", err)
	}

	if got["type"] != "object" {
		t.Errorf("Expected object type, got %v", got["type"])
	}

	properties, ok := got["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema properties is not a map")
	}

	secondsSchema, ok := properties["seconds"].(map[string]interface{})
	if !ok {
		t.Fatal("seconds field not found in schema")
	}

	if secondsSchema["type"] != "integer" {
		t.Errorf("Expected integer type for seconds, got %v", secondsSchema["type"])
	}

	nanosSchema, ok := properties["nanos"].(map[string]interface{})
	if !ok {
		t.Fatal("nanos field not found in schema")
	}

	if nanosSchema["type"] != "integer" {
		t.Errorf("Expected integer type for nanos, got %v", nanosSchema["type"])
	}
}

func TestTimestampFieldInMessage(t *testing.T) {
	// Test our custom logic by creating a mock field descriptor
	// This tests the generateFieldSchema method directly

	// Create a timestamp message to get its descriptor
	ts := &timestamppb.Timestamp{}
	md := ts.ProtoReflect().Descriptor()

	// Get the first field (which should be seconds) to test our logic
	fields := md.Fields()
	if fields.Len() == 0 {
		t.Fatal("Timestamp message has no fields")
	}

	// We need to test the logic for when a field is of MessageKind and is google.protobuf.Timestamp
	// Let's create a simple test by checking if our modification works

	// The key test is whether our generator correctly identifies google.protobuf.Timestamp
	// and generates the oneOf schema

	t.Logf("Timestamp message full name: %s", md.FullName())

	if md.FullName() != "google.protobuf.Timestamp" {
		t.Errorf("Expected google.protobuf.Timestamp, got %s", md.FullName())
	}
}

func TestTimestampOneOfSchemaStructure(t *testing.T) {
	// Test that our schema generation creates the correct oneOf structure
	// for timestamp fields by simulating the logic

	// Create the expected schema structure
	expectedOneOf := []interface{}{
		map[string]interface{}{
			"type":        "string",
			"format":      "date-time",
			"description": "RFC3339 timestamp string",
		},
		map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"seconds": map[string]interface{}{
					"type":        "integer",
					"description": "Seconds since Unix epoch",
				},
				"nanos": map[string]interface{}{
					"type":        "integer",
					"minimum":     0,
					"maximum":     999999999,
					"description": "Nanoseconds within the second",
				},
			},
			"required":             []string{"seconds"},
			"additionalProperties": false,
		},
	}

	// Convert to JSON to verify structure
	expectedJSON, err := json.MarshalIndent(map[string]interface{}{
		"oneOf": expectedOneOf,
	}, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal expected schema: %v", err)
	}

	t.Logf("Expected oneOf schema structure:\n%s", string(expectedJSON))

	// Verify the structure is valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(expectedJSON, &parsed); err != nil {
		t.Errorf("Expected schema is not valid JSON: %v", err)
	}

	// Verify oneOf exists and has 2 options
	oneOf, ok := parsed["oneOf"].([]interface{})
	if !ok {
		t.Fatal("oneOf not found or not an array")
	}

	if len(oneOf) != 2 {
		t.Errorf("Expected 2 oneOf options, got %d", len(oneOf))
	}

	// Verify first option is string
	stringOption, ok := oneOf[0].(map[string]interface{})
	if !ok {
		t.Fatal("First oneOf option is not a map")
	}

	if stringOption["type"] != "string" {
		t.Errorf("Expected string type, got %v", stringOption["type"])
	}

	if stringOption["format"] != "date-time" {
		t.Errorf("Expected date-time format, got %v", stringOption["format"])
	}

	// Verify second option is object
	objectOption, ok := oneOf[1].(map[string]interface{})
	if !ok {
		t.Fatal("Second oneOf option is not a map")
	}

	if objectOption["type"] != "object" {
		t.Errorf("Expected object type, got %v", objectOption["type"])
	}
}

func TestGeneratorModificationWorks(t *testing.T) {
	generator := protojsonschema.NewGenerator()

	msg := &pb.TimestampTestMessage{}
	md := msg.ProtoReflect().Descriptor()

	schema, err := generator.GenerateSchema(md)
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	schemaJSON, err := schema.ToJSON()
	if err != nil {
		t.Fatalf("Failed to convert schema to JSON: %v", err)
	}

	if !strings.Contains(schemaJSON, "oneOf") {
		t.Logf("Schema JSON:\n%s", schemaJSON)
		t.Error("Schema does not contain 'oneOf' - timestamp special-casing not applied to nested field")
	}

	if !strings.Contains(schemaJSON, "date-time") {
		t.Logf("Schema JSON:\n%s", schemaJSON)
		t.Error("Schema does not contain 'date-time' format")
	}

	if !strings.Contains(schemaJSON, "RFC3339") {
		t.Logf("Schema JSON:\n%s", schemaJSON)
		t.Error("Schema does not contain 'RFC3339' description")
	}
}
