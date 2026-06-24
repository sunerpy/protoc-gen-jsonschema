package jsonschema

import (
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	jsonschemapb "github.com/sunerpy/protoc-gen-jsonschema/mcp/jsonschema"
)

func TestGenerateFieldSchema_NumericAndStringOptions(t *testing.T) {
	g := NewGenerator()
	md := (&timestamppb.Timestamp{}).ProtoReflect().Descriptor()
	field := md.Fields().Get(0) // seconds (int64 -> integer)

	opts := &descriptorpb.FieldOptions{}
	proto.SetExtension(opts, jsonschemapb.E_Description, "a number")
	proto.SetExtension(opts, jsonschemapb.E_Example, "42")
	proto.SetExtension(opts, jsonschemapb.E_Format, "int64")
	proto.SetExtension(opts, jsonschemapb.E_MinLength, int32(1))
	proto.SetExtension(opts, jsonschemapb.E_MaxLength, int32(10))
	proto.SetExtension(opts, jsonschemapb.E_Minimum, float64(0))
	proto.SetExtension(opts, jsonschemapb.E_Maximum, float64(100))
	proto.SetExtension(opts, jsonschemapb.E_Pattern, "^[0-9]+$")
	proto.SetExtension(opts, jsonschemapb.E_Default, "0")

	schema := g.generateFieldSchema(field, opts)

	if schema["type"] != "integer" {
		t.Errorf("expected integer type, got %v", schema["type"])
	}
	if schema["description"] != "a number" {
		t.Errorf("expected description option, got %v", schema["description"])
	}
	if schema["example"] != "42" {
		t.Errorf("expected example option, got %v", schema["example"])
	}
	if schema["format"] != "int64" {
		t.Errorf("expected format option, got %v", schema["format"])
	}
	if schema["minLength"] != int32(1) {
		t.Errorf("expected minLength 1, got %v", schema["minLength"])
	}
	if schema["maxLength"] != int32(10) {
		t.Errorf("expected maxLength 10, got %v", schema["maxLength"])
	}
	if schema["minimum"] != float64(0) {
		t.Errorf("expected minimum 0, got %v", schema["minimum"])
	}
	if schema["maximum"] != float64(100) {
		t.Errorf("expected maximum 100, got %v", schema["maximum"])
	}
	if schema["pattern"] != "^[0-9]+$" {
		t.Errorf("expected pattern option, got %v", schema["pattern"])
	}
	if schema["default"] != float64(0) {
		t.Errorf("expected default parsed as JSON 0, got %v", schema["default"])
	}
}

func TestGenerateFieldSchema_HiddenField(t *testing.T) {
	g := NewGenerator()
	opts := &descriptorpb.FieldOptions{}
	proto.SetExtension(opts, jsonschemapb.E_Hidden, true)
	if !g.isFieldHidden(opts) {
		t.Error("expected field to be hidden")
	}
}

func TestGenerateFieldSchema_CustomJSONName(t *testing.T) {
	g := NewGenerator()
	md := (&timestamppb.Timestamp{}).ProtoReflect().Descriptor()
	field := md.Fields().Get(0)

	opts := &descriptorpb.FieldOptions{}
	proto.SetExtension(opts, jsonschemapb.E_JsonName, "custom_seconds")
	if name := g.getFieldName(field, opts); name != "custom_seconds" {
		t.Errorf("expected custom_seconds, got %s", name)
	}
}

func TestShouldGenerateSchema_Disabled(t *testing.T) {
	g := NewGenerator()
	opts := &descriptorpb.MessageOptions{}
	proto.SetExtension(opts, jsonschemapb.E_GenerateSchema, false)
	if g.shouldGenerateSchema(opts) {
		t.Error("expected schema generation to be disabled")
	}
}

func TestGenerateSchema_DisabledReturnsNil(t *testing.T) {
	g := NewGenerator()
	md := (&timestamppb.Timestamp{}).ProtoReflect().Descriptor()
	field := md.Fields().Get(0)
	opts := &descriptorpb.FieldOptions{}
	proto.SetExtension(opts, jsonschemapb.E_Default, "not-valid-json{")

	schema := g.generateFieldSchema(field, opts)
	if _, ok := schema["default"]; ok {
		t.Error("invalid JSON default should be skipped, not set")
	}
}
