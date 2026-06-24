package test

import (
	"encoding/json"
	"strings"
	"testing"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/sunerpy/protoc-gen-jsonschema/test/pb"
)

// schemaAsMap generates a schema and round-trips it through JSON so every nested
// value is a plain map[string]interface{}. This avoids brittle type assertions
// against the named jsonschema.Schema type, and exercises the exact serialized
// shape that downstream JSON consumers receive.
func schemaAsMap(t *testing.T, msg interface{ GetJSONSchema() string }) map[string]interface{} {
	t.Helper()
	var out map[string]interface{}
	if err := json.Unmarshal([]byte(msg.GetJSONSchema()), &out); err != nil {
		t.Fatalf("failed to unmarshal generated schema: %v", err)
	}
	return out
}

func TestTimestampSchemaGeneration(t *testing.T) {
	msg := &pb.TimestampTestMessage{}
	schema := schemaAsMap(t, msg)

	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema properties is not a map")
	}

	createdAtSchema, ok := properties["createdAt"].(map[string]interface{})
	if !ok {
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

	if nanosProperty["minimum"] != float64(0) {
		t.Errorf("Expected minimum 0 for nanos, got %v", nanosProperty["minimum"])
	}

	if nanosProperty["maximum"] != float64(999999999) {
		t.Errorf("Expected maximum 999999999 for nanos, got %v", nanosProperty["maximum"])
	}
}

func TestTimestampArraySchemaGeneration(t *testing.T) {
	msg := &pb.TimestampTestMessage{}
	schema := schemaAsMap(t, msg)

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
	// protojson only accepts RFC3339 string form for google.protobuf.Timestamp.
	// The schema's oneOf object branch targets generic JSON consumers, not the
	// proto runtime, so only the string form is exercised against protojson here.
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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := &pb.TimestampTestMessage{}
			err := protojson.Unmarshal([]byte(tc.jsonData), msg)
			if err != nil {
				t.Errorf("protojson.Unmarshal failed: %v", err)
			}

			if msg.RequiredTimestamp == nil {
				t.Error("RequiredTimestamp should not be nil")
			}

			if msg.CreatedAt == nil {
				t.Error("CreatedAt should not be nil")
			}

			jsonBytes, err := protojson.Marshal(msg)
			if err != nil {
				t.Errorf("protojson.Marshal failed: %v", err)
			}

			t.Logf("Marshaled JSON: %s", string(jsonBytes))
		})
	}
}
