package main

import (
	"strings"
	"testing"
)

func TestGenerateGoogleSchemaLiteral_NestedOneOfAndItems(t *testing.T) {
	m := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"ts": map[string]interface{}{
				"oneOf": []interface{}{
					map[string]interface{}{"type": "string", "format": "date-time"},
					map[string]interface{}{
						"type":                 "object",
						"properties":           map[string]interface{}{"seconds": map[string]interface{}{"type": "integer"}},
						"additionalProperties": false,
					},
				},
			},
			"events": map[string]interface{}{
				"type":  "array",
				"items": map[string]interface{}{"type": "object", "properties": map[string]interface{}{"id": map[string]interface{}{"type": "string"}}},
			},
			"name": map[string]interface{}{
				"type":    "string",
				"example": "alice",
				"default": "anon",
			},
		},
	}

	out := generateGoogleSchemaLiteral(m, 0)

	for _, want := range []string{
		"OneOf:",
		"Items:",
		"AdditionalProperties: &jsonschema.Schema{Not: &jsonschema.Schema{}}",
		`Examples: []any{"alice"}`,
		`Default: json.RawMessage("\"anon\"")`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected output to contain %q\n--- got ---\n%s", want, out)
		}
	}
	if strings.Contains(out, "Items: &jsonschema.Schema{\n\t\t}") {
		t.Error("Items must be populated, not empty")
	}
}

func TestGenerateGoogleSchemaLiteral_PropertiesSorted(t *testing.T) {
	m := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"zebra": map[string]interface{}{"type": "string"},
			"alpha": map[string]interface{}{"type": "string"},
			"mango": map[string]interface{}{"type": "string"},
		},
	}
	out := generateGoogleSchemaLiteral(m, 0)
	ia := strings.Index(out, `"alpha"`)
	im := strings.Index(out, `"mango"`)
	iz := strings.Index(out, `"zebra"`)
	if ia >= im || im >= iz {
		t.Errorf("properties must be sorted alpha<mango<zebra, got positions %d,%d,%d", ia, im, iz)
	}
}
