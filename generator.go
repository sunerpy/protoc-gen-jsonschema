package jsonschema

import (
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"

	jsonschemapb "github.com/sunerpy/protoc-gen-jsonschema/mcp/jsonschema"
)

// Schema represents a JSON Schema
type Schema map[string]interface{}

// Generator generates JSON Schema from protobuf messages
type Generator struct {
	schemas map[string]Schema
}

// NewGenerator creates a new Generator
func NewGenerator() *Generator {
	return &Generator{
		schemas: make(map[string]Schema),
	}
}

// GenerateSchema generates JSON Schema for a message descriptor
func (g *Generator) GenerateSchema(md protoreflect.MessageDescriptor) (Schema, error) {
	msgOpts := md.Options().(*descriptorpb.MessageOptions)
	
	// Check if schema generation is disabled
	if proto.HasExtension(msgOpts, jsonschemapb.E_GenerateSchema) {
		if !proto.GetExtension(msgOpts, jsonschemapb.E_GenerateSchema).(bool) {
			return nil, nil
		}
	}

	schema := Schema{
		"type":       "object",
		"properties": make(map[string]interface{}),
	}

	// Set title
	if proto.HasExtension(msgOpts, jsonschemapb.E_Title) {
		schema["title"] = proto.GetExtension(msgOpts, jsonschemapb.E_Title).(string)
	} else {
		schema["title"] = string(md.Name())
	}

	// Set description
	if proto.HasExtension(msgOpts, jsonschemapb.E_MessageDescription) {
		schema["description"] = proto.GetExtension(msgOpts, jsonschemapb.E_MessageDescription).(string)
	}

	required := []string{}
	properties := make(map[string]interface{})

	// Process fields
	fields := md.Fields()
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		fieldOpts := field.Options().(*descriptorpb.FieldOptions)

		// Check if field is hidden
		if proto.HasExtension(fieldOpts, jsonschemapb.E_Hidden) {
			if proto.GetExtension(fieldOpts, jsonschemapb.E_Hidden).(bool) {
				continue
			}
		}

		fieldName := string(field.Name())
		
		// Use custom JSON name if specified
		if proto.HasExtension(fieldOpts, jsonschemapb.E_JsonName) {
			fieldName = proto.GetExtension(fieldOpts, jsonschemapb.E_JsonName).(string)
		} else if field.JSONName() != "" {
			fieldName = field.JSONName()
		}

		fieldSchema := g.generateFieldSchema(field, fieldOpts)
		properties[fieldName] = fieldSchema

		// Check if field is required
		if proto.HasExtension(fieldOpts, jsonschemapb.E_Required) {
			if proto.GetExtension(fieldOpts, jsonschemapb.E_Required).(bool) {
				required = append(required, fieldName)
			}
		}
	}

	schema["properties"] = properties
	if len(required) > 0 {
		schema["required"] = required
	}

	return schema, nil
}

// generateFieldSchema generates JSON Schema for a field
func (g *Generator) generateFieldSchema(field protoreflect.FieldDescriptor, opts *descriptorpb.FieldOptions) Schema {
	schema := Schema{}

	// Set type based on protobuf type
	switch field.Kind() {
	case protoreflect.BoolKind:
		schema["type"] = "boolean"
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind,
		protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind,
		protoreflect.Uint32Kind, protoreflect.Fixed32Kind,
		protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		schema["type"] = "integer"
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		schema["type"] = "number"
	case protoreflect.StringKind:
		schema["type"] = "string"
	case protoreflect.BytesKind:
		schema["type"] = "string"
		schema["format"] = "byte"
	case protoreflect.EnumKind:
		schema["type"] = "string"
		// Add enum values
		enumValues := []string{}
		enumDesc := field.Enum()
		for i := 0; i < enumDesc.Values().Len(); i++ {
			enumValues = append(enumValues, string(enumDesc.Values().Get(i).Name()))
		}
		schema["enum"] = enumValues
	case protoreflect.MessageKind:
		schema["type"] = "object"
	}

	// Handle repeated fields
	if field.Cardinality() == protoreflect.Repeated {
		schema = Schema{
			"type":  "array",
			"items": schema,
		}
	}

	// Apply custom options
	if proto.HasExtension(opts, jsonschemapb.E_Description) {
		schema["description"] = proto.GetExtension(opts, jsonschemapb.E_Description).(string)
	}

	if proto.HasExtension(opts, jsonschemapb.E_Example) {
		schema["example"] = proto.GetExtension(opts, jsonschemapb.E_Example).(string)
	}

	if proto.HasExtension(opts, jsonschemapb.E_Format) {
		schema["format"] = proto.GetExtension(opts, jsonschemapb.E_Format).(string)
	}

	if proto.HasExtension(opts, jsonschemapb.E_Default) {
		defaultStr := proto.GetExtension(opts, jsonschemapb.E_Default).(string)
		var defaultValue interface{}
		if err := json.Unmarshal([]byte(defaultStr), &defaultValue); err == nil {
			schema["default"] = defaultValue
		}
	}

	if proto.HasExtension(opts, jsonschemapb.E_MinLength) {
		schema["minLength"] = proto.GetExtension(opts, jsonschemapb.E_MinLength).(int32)
	}

	if proto.HasExtension(opts, jsonschemapb.E_MaxLength) {
		schema["maxLength"] = proto.GetExtension(opts, jsonschemapb.E_MaxLength).(int32)
	}

	if proto.HasExtension(opts, jsonschemapb.E_Minimum) {
		schema["minimum"] = proto.GetExtension(opts, jsonschemapb.E_Minimum).(float64)
	}

	if proto.HasExtension(opts, jsonschemapb.E_Maximum) {
		schema["maximum"] = proto.GetExtension(opts, jsonschemapb.E_Maximum).(float64)
	}

	if proto.HasExtension(opts, jsonschemapb.E_Pattern) {
		schema["pattern"] = proto.GetExtension(opts, jsonschemapb.E_Pattern).(string)
	}

	return schema
}

// ToJSON converts schema to JSON string
func (s Schema) ToJSON() (string, error) {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal schema: %w", err)
	}
	return string(data), nil
}

// ToJSONBytes converts schema to JSON bytes
func (s Schema) ToJSONBytes() ([]byte, error) {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema: %w", err)
	}
	return data, nil
}

// GenerateFromMessage generates JSON Schema from a protobuf message
func GenerateFromMessage(msg proto.Message) (Schema, error) {
	gen := NewGenerator()
	md := msg.ProtoReflect().Descriptor()
	return gen.GenerateSchema(md)
}

// GenerateJSONFromMessage generates JSON Schema string from a protobuf message
func GenerateJSONFromMessage(msg proto.Message) (string, error) {
	schema, err := GenerateFromMessage(msg)
	if err != nil {
		return "", err
	}
	return schema.ToJSON()
}
