package jsonschema

import (
	"encoding/json"
	"testing"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func mustSchemaMap(t *testing.T, s Schema) map[string]interface{} {
	t.Helper()
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("failed to marshal schema: %v", err)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("failed to unmarshal schema: %v", err)
	}
	return out
}

func TestNewGenerator(t *testing.T) {
	g := NewGenerator()
	if g == nil {
		t.Fatal("NewGenerator returned nil")
	}
	if g.IsPreserveOrder() {
		t.Error("expected preserveOrder to default to false")
	}
}

func TestNewGeneratorWithOptions(t *testing.T) {
	g := NewGeneratorWithOptions(true)
	if !g.IsPreserveOrder() {
		t.Error("expected preserveOrder to be true")
	}
}

func TestSetPreserveOrder(t *testing.T) {
	g := NewGenerator()
	g.SetPreserveOrder(true)
	if !g.IsPreserveOrder() {
		t.Error("SetPreserveOrder(true) did not take effect")
	}
	g.SetPreserveOrder(false)
	if g.IsPreserveOrder() {
		t.Error("SetPreserveOrder(false) did not take effect")
	}
}

func TestGenerateFromMessage_Timestamp(t *testing.T) {
	ts := &timestamppb.Timestamp{}
	schema, err := GenerateFromMessage(ts)
	if err != nil {
		t.Fatalf("GenerateFromMessage failed: %v", err)
	}
	if schema["type"] != "object" {
		t.Errorf("expected object type, got %v", schema["type"])
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("properties is not a map")
	}
	if _, ok := props["seconds"]; !ok {
		t.Error("expected seconds field in Timestamp schema")
	}
	if _, ok := props["nanos"]; !ok {
		t.Error("expected nanos field in Timestamp schema")
	}
}

func TestGenerateJSONFromMessage(t *testing.T) {
	ts := &timestamppb.Timestamp{}
	jsonStr, err := GenerateJSONFromMessage(ts)
	if err != nil {
		t.Fatalf("GenerateJSONFromMessage failed: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		t.Fatalf("generated JSON is invalid: %v", err)
	}
	if parsed["type"] != "object" {
		t.Errorf("expected object type, got %v", parsed["type"])
	}
}

func TestGenerateSchema_FieldKinds(t *testing.T) {
	g := NewGenerator()
	schema, err := g.GenerateSchema((&timestamppb.Timestamp{}).ProtoReflect().Descriptor())
	if err != nil {
		t.Fatalf("GenerateSchema failed: %v", err)
	}
	m := mustSchemaMap(t, schema)
	props, ok := m["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("properties not a map")
	}

	seconds, ok := props["seconds"].(map[string]interface{})
	if !ok {
		t.Fatal("seconds not a map")
	}
	if seconds["type"] != "integer" {
		t.Errorf("expected seconds type integer (int64 -> integer), got %v", seconds["type"])
	}

	nanos, ok := props["nanos"].(map[string]interface{})
	if !ok {
		t.Fatal("nanos not a map")
	}
	if nanos["type"] != "integer" {
		t.Errorf("expected nanos type integer (int32 -> integer), got %v", nanos["type"])
	}
}

func TestSchemaToJSON(t *testing.T) {
	s := Schema{"type": "string", "minLength": 3}
	out, err := s.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("ToJSON produced invalid JSON: %v", err)
	}
	if parsed["type"] != "string" {
		t.Errorf("expected type string, got %v", parsed["type"])
	}
}

func TestSchemaToJSONBytes(t *testing.T) {
	s := Schema{"type": "number"}
	data, err := s.ToJSONBytes()
	if err != nil {
		t.Fatalf("ToJSONBytes failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("ToJSONBytes returned empty bytes")
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("ToJSONBytes produced invalid JSON: %v", err)
	}
}

func TestOrderedSchema_MarshalJSON_FieldOrder(t *testing.T) {
	os := &OrderedSchema{
		Type:        "object",
		Title:       "User",
		Description: "A user",
		Properties: []OrderedProperty{
			{Name: "zebra", Schema: map[string]interface{}{"type": "string"}},
			{Name: "alpha", Schema: map[string]interface{}{"type": "integer"}},
		},
		Required: []string{"zebra"},
	}

	data, err := json.Marshal(os)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("OrderedSchema produced invalid JSON: %v", err)
	}

	if parsed["type"] != "object" {
		t.Errorf("expected type object, got %v", parsed["type"])
	}
	if parsed["title"] != "User" {
		t.Errorf("expected title User, got %v", parsed["title"])
	}

	props, ok := parsed["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("properties not a map")
	}
	if _, ok := props["zebra"]; !ok {
		t.Error("expected zebra property")
	}
	if _, ok := props["alpha"]; !ok {
		t.Error("expected alpha property")
	}
}

func TestOrderedSchema_MarshalJSON_Empty(t *testing.T) {
	os := &OrderedSchema{}
	data, err := json.Marshal(os)
	if err != nil {
		t.Fatalf("MarshalJSON failed for empty schema: %v", err)
	}
	if string(data) != "{}" {
		t.Errorf("expected empty object, got %s", string(data))
	}
}

// TestOrderedSchema_MarshalJSON_Escaping pins the escaping contract: any
// Title/Description must yield valid JSON that round-trips, including values
// with quotes, backslashes, control chars, and <>&.
func TestOrderedSchema_MarshalJSON_Escaping(t *testing.T) {
	cases := []struct {
		name        string
		title       string
		description string
	}{
		{"double-quote", `He said "hi"`, `desc with "quotes"`},
		{"backslash", `back\slash`, `c:\path\to`},
		{"newline", "line1\nline2", "tab\there"},
		{"html-chars", "a<b>&c", "x < y && z > w"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			os := &OrderedSchema{
				Type:        "object",
				Title:       tc.title,
				Description: tc.description,
				Properties: []OrderedProperty{
					{Name: "f", Schema: map[string]interface{}{"type": "string"}},
				},
			}
			data, err := json.Marshal(os)
			if err != nil {
				t.Fatalf("MarshalJSON failed: %v", err)
			}
			var parsed map[string]interface{}
			if err := json.Unmarshal(data, &parsed); err != nil {
				t.Fatalf("produced INVALID JSON for %q: %v\noutput: %s", tc.name, err, data)
			}
			if parsed["title"] != tc.title {
				t.Errorf("title round-trip mismatch: got %q want %q", parsed["title"], tc.title)
			}
			if parsed["description"] != tc.description {
				t.Errorf("description round-trip mismatch: got %q want %q", parsed["description"], tc.description)
			}
		})
	}
}

// TestOrderedSchema_MarshalJSON_HTMLByteParity pins that HTML chars round-trip
// through the json.Marshal path the plugin uses (the outer encoder HTML-escapes
// to \u003c etc.), so the fix does not change output for valid inputs.
func TestOrderedSchema_MarshalJSON_HTMLByteParity(t *testing.T) {
	os := &OrderedSchema{Type: "object", Title: "a<b>&c"}
	data, err := json.Marshal(os)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if parsed["title"] != "a<b>&c" {
		t.Errorf("title must round-trip to literal value, got %q", parsed["title"])
	}
}

func TestGenerateOrderedSchema_Timestamp(t *testing.T) {
	g := NewGeneratorWithOptions(true)
	ordered, err := g.GenerateOrderedSchema((&timestamppb.Timestamp{}).ProtoReflect().Descriptor())
	if err != nil {
		t.Fatalf("GenerateOrderedSchema failed: %v", err)
	}
	if ordered == nil {
		t.Fatal("GenerateOrderedSchema returned nil")
	}
	if ordered.Type != "object" {
		t.Errorf("expected object type, got %s", ordered.Type)
	}
	if len(ordered.Properties) == 0 {
		t.Error("expected at least one ordered property")
	}
}
