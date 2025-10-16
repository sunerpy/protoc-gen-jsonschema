package test

import (
	"strings"
	"testing"

	protojsonschema "github.com/sunerpy/protoc-gen-jsonschema"
	"github.com/sunerpy/protoc-gen-jsonschema/test/pb"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestTimestampSchemaGeneration(t *testing.T) {
	// Create a test message instance
	msg := &pb.TimestampTestMessage{}

	// Generate schema using the protobuf generator
	generator := protojsonschema.NewGenerator()
	schema, err := generator.GenerateSchema(msg.ProtoReflect().Descriptor())
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	// Convert schema to JSON for inspection
	schemaJSON, err := schema.ToJSON()
	if err != nil {
		t.Fatalf("Failed to convert schema to JSON: %v", err)
	}

	t.Logf("Generated schema:\n%s", schemaJSON)

	// Verify that timestamp fields have oneOf structure
	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema properties is not a map")
	}

	// Check createdAt field (should be in the schema)
	createdAtSchema, ok := properties["createdAt"].(map[string]interface{})
	if !ok {
		// Print available properties for debugging
		t.Logf("Available properties: %v", properties)
		t.Fatal("createdAt field not found in schema")
	}

	oneOf, ok := createdAtSchema["oneOf"].([]interface{})
	if !ok {
		t.Fatal("createdAt field does not have oneOf structure")
	}

	if len(oneOf) != 2 {
		t.Fatalf("Expected 2 oneOf options, got %d", len(oneOf))
	}

	// Verify first option is string with date-time format
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

	// Verify second option is object with seconds and nanos
	objectOption, ok := oneOf[1].(map[string]interface{})
	if !ok {
		t.Fatal("Second oneOf option is not a map")
	}

	if objectOption["type"] != "object" {
		t.Errorf("Expected object type, got %v", objectOption["type"])
	}

	objProperties, ok := objectOption["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Object option does not have properties")
	}

	// Check seconds property
	secondsProperty, ok := objProperties["seconds"].(map[string]interface{})
	if !ok {
		t.Fatal("seconds property not found")
	}

	if secondsProperty["type"] != "integer" {
		t.Errorf("Expected integer type for seconds, got %v", secondsProperty["type"])
	}

	// Check nanos property
	nanosProperty, ok := objProperties["nanos"].(map[string]interface{})
	if !ok {
		t.Fatal("nanos property not found")
	}

	if nanosProperty["type"] != "integer" {
		t.Errorf("Expected integer type for nanos, got %v", nanosProperty["type"])
	}

	if nanosProperty["minimum"] != 0 {
		t.Errorf("Expected minimum 0 for nanos, got %v", nanosProperty["minimum"])
	}

	if nanosProperty["maximum"] != 999999999 {
		t.Errorf("Expected maximum 999999999 for nanos, got %v", nanosProperty["maximum"])
	}
}

func TestTimestampArraySchemaGeneration(t *testing.T) {
	// Create a test message instance
	msg := &pb.TimestampTestMessage{}

	// Generate schema
	generator := protojsonschema.NewGenerator()
	schema, err := generator.GenerateSchema(msg.ProtoReflect().Descriptor())
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	// Check array field
	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema properties is not a map")
	}

	eventTimestampsSchema, ok := properties["eventTimestamps"].(map[string]interface{})
	if !ok {
		t.Fatal("eventTimestamps field not found in schema")
	}

	if eventTimestampsSchema["type"] != "array" {
		t.Errorf("Expected array type, got %v", eventTimestampsSchema["type"])
	}

	items, ok := eventTimestampsSchema["items"].(map[string]interface{})
	if !ok {
		t.Fatal("Array items not found")
	}

	oneOf, ok := items["oneOf"].([]interface{})
	if !ok {
		t.Fatal("Array items do not have oneOf structure")
	}

	if len(oneOf) != 2 {
		t.Fatalf("Expected 2 oneOf options for array items, got %d", len(oneOf))
	}
}

func TestTimestampJSONSchemaContent(t *testing.T) {
	// Get the generated schema JSON string
	msg := &pb.TimestampTestMessage{}
	schemaJSON := msg.GetJSONSchema()

	t.Logf("Generated schema JSON:\n%s", schemaJSON)

	// Verify that the schema contains expected oneOf structures
	if !strings.Contains(schemaJSON, "oneOf") {
		t.Error("Generated schema should contain 'oneOf' structure")
	}

	if !strings.Contains(schemaJSON, "date-time") {
		t.Error("Generated schema should contain 'date-time' format")
	}

	if !strings.Contains(schemaJSON, "RFC3339") {
		t.Error("Generated schema should contain 'RFC3339' description")
	}

	// Verify that both string and object formats are supported
	if !strings.Contains(schemaJSON, `"type":"string"`) {
		t.Error("Generated schema should support string type")
	}

	if !strings.Contains(schemaJSON, `"type":"object"`) {
		t.Error("Generated schema should support object type")
	}

	// Verify nanos constraints
	if !strings.Contains(schemaJSON, `"maximum":999999999`) {
		t.Error("Generated schema should have maximum constraint for nanos")
	}

	if !strings.Contains(schemaJSON, `"minimum":0`) {
		t.Error("Generated schema should have minimum constraint for nanos")
	}

	t.Log("✓ Generated schema contains all expected timestamp validation structures")
}

func TestProtojsonCompatibilityWithGeneratedSchema(t *testing.T) {
	// Test that protojson can unmarshal RFC3339 strings successfully
	testCases := []struct {
		name     string
		jsonData string
	}{
		{
			name: "RFC3339 string",
			jsonData: `{
				"requiredTimestamp": "2023-01-01T00:00:00Z",
				"createdAt": "2023-01-01T12:30:45.123Z"
			}`,
		},
		{
			name: "Object format",
			jsonData: `{
				"requiredTimestamp": {"seconds": 1672531200, "nanos": 0},
				"createdAt": {"seconds": 1672576245, "nanos": 123000000}
			}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := &pb.TimestampTestMessage{}
			err := protojson.Unmarshal([]byte(tc.jsonData), msg)
			if err != nil {
				t.Errorf("protojson.Unmarshal failed: %v", err)
			}

			// Verify that timestamps were parsed correctly
			if msg.RequiredTimestamp == nil {
				t.Error("RequiredTimestamp should not be nil")
			}

			if msg.CreatedAt == nil {
				t.Error("CreatedAt should not be nil")
			}

			// Test marshaling back to JSON
			jsonBytes, err := protojson.Marshal(msg)
			if err != nil {
				t.Errorf("protojson.Marshal failed: %v", err)
			}

			t.Logf("Marshaled JSON: %s", string(jsonBytes))
		})
	}
}
