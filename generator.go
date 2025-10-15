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
	schemas       map[string]Schema
	preserveOrder bool
}

// NewGenerator creates a new Generator
func NewGenerator() *Generator {
	return &Generator{
		schemas: make(map[string]Schema),
	}
}

// NewGeneratorWithOptions creates a new Generator with options
func NewGeneratorWithOptions(preserveOrder bool) *Generator {
	return &Generator{
		schemas:       make(map[string]Schema),
		preserveOrder: preserveOrder,
	}
}

// SetPreserveOrder sets whether to preserve field order
func (g *Generator) SetPreserveOrder(preserve bool) {
	g.preserveOrder = preserve
}

// IsPreserveOrder returns whether field order preservation is enabled
func (g *Generator) IsPreserveOrder() bool {
	return g.preserveOrder
}

// GenerateSchema generates JSON Schema for a message descriptor
func (g *Generator) GenerateSchema(md protoreflect.MessageDescriptor) (Schema, error) {
	msgOpts := md.Options().(*descriptorpb.MessageOptions)

	if !g.shouldGenerateSchema(msgOpts) {
		return nil, nil
	}

	schema := g.createBaseSchema(md, msgOpts)
	properties, required := g.processFields(md.Fields())

	schema["properties"] = properties
	if len(required) > 0 {
		schema["required"] = required
	}

	return schema, nil
}

// GenerateOrderedSchema generates an ordered JSON Schema for a message descriptor
func (g *Generator) GenerateOrderedSchema(md protoreflect.MessageDescriptor) (*OrderedSchema, error) {
	msgOpts := md.Options().(*descriptorpb.MessageOptions)

	if !g.shouldGenerateSchema(msgOpts) {
		return nil, nil
	}

	orderedSchema := &OrderedSchema{
		Type:       "object",
		Title:      g.getSchemaTitle(md, msgOpts),
		Properties: []OrderedProperty{},
		Required:   []string{},
	}

	if proto.HasExtension(msgOpts, jsonschemapb.E_MessageDescription) {
		orderedSchema.Description = proto.GetExtension(msgOpts, jsonschemapb.E_MessageDescription).(string)
	}

	// Process fields in order
	fields := md.Fields()
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		fieldOpts := field.Options().(*descriptorpb.FieldOptions)

		if g.isFieldHidden(fieldOpts) {
			continue
		}

		fieldName := g.getFieldName(field, fieldOpts)
		fieldSchema := g.generateFieldSchema(field, fieldOpts)

		orderedSchema.Properties = append(orderedSchema.Properties, OrderedProperty{
			Name:   fieldName,
			Schema: fieldSchema,
		})

		if g.isFieldRequired(fieldOpts) {
			orderedSchema.Required = append(orderedSchema.Required, fieldName)
		}
	}

	return orderedSchema, nil
}

// shouldGenerateSchema checks if schema generation is enabled
func (g *Generator) shouldGenerateSchema(msgOpts *descriptorpb.MessageOptions) bool {
	if proto.HasExtension(msgOpts, jsonschemapb.E_GenerateSchema) {
		return proto.GetExtension(msgOpts, jsonschemapb.E_GenerateSchema).(bool)
	}
	return true
}

// createBaseSchema creates the base schema with title and description
func (g *Generator) createBaseSchema(md protoreflect.MessageDescriptor, msgOpts *descriptorpb.MessageOptions) Schema {
	schema := Schema{
		"type":       "object",
		"properties": make(map[string]interface{}),
	}

	schema["title"] = g.getSchemaTitle(md, msgOpts)

	if proto.HasExtension(msgOpts, jsonschemapb.E_MessageDescription) {
		schema["description"] = proto.GetExtension(msgOpts, jsonschemapb.E_MessageDescription).(string)
	}

	return schema
}

// getSchemaTitle returns the schema title from options or message name
func (g *Generator) getSchemaTitle(md protoreflect.MessageDescriptor, msgOpts *descriptorpb.MessageOptions) string {
	if proto.HasExtension(msgOpts, jsonschemapb.E_Title) {
		return proto.GetExtension(msgOpts, jsonschemapb.E_Title).(string)
	}
	return string(md.Name())
}

// processFields processes all fields and returns properties and required fields
func (g *Generator) processFields(fields protoreflect.FieldDescriptors) (map[string]interface{}, []string) {
	properties := make(map[string]interface{})
	required := []string{}

	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		fieldOpts := field.Options().(*descriptorpb.FieldOptions)

		if g.isFieldHidden(fieldOpts) {
			continue
		}

		fieldName := g.getFieldName(field, fieldOpts)
		fieldSchema := g.generateFieldSchema(field, fieldOpts)
		properties[fieldName] = fieldSchema

		if g.isFieldRequired(fieldOpts) {
			required = append(required, fieldName)
		}
	}

	return properties, required
}

// isFieldHidden checks if a field should be hidden
func (g *Generator) isFieldHidden(fieldOpts *descriptorpb.FieldOptions) bool {
	if proto.HasExtension(fieldOpts, jsonschemapb.E_Hidden) {
		return proto.GetExtension(fieldOpts, jsonschemapb.E_Hidden).(bool)
	}
	return false
}

// getFieldName returns the JSON name for a field
func (g *Generator) getFieldName(field protoreflect.FieldDescriptor, fieldOpts *descriptorpb.FieldOptions) string {
	if proto.HasExtension(fieldOpts, jsonschemapb.E_JsonName) {
		return proto.GetExtension(fieldOpts, jsonschemapb.E_JsonName).(string)
	}
	if field.JSONName() != "" {
		return field.JSONName()
	}
	return string(field.Name())
}

// isFieldRequired checks if a field is required
func (g *Generator) isFieldRequired(fieldOpts *descriptorpb.FieldOptions) bool {
	if proto.HasExtension(fieldOpts, jsonschemapb.E_Required) {
		return proto.GetExtension(fieldOpts, jsonschemapb.E_Required).(bool)
	}
	return false
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

// OrderedSchema represents a JSON Schema with ordered fields
type OrderedSchema struct {
	Type        string
	Title       string
	Description string
	Properties  []OrderedProperty
	Required    []string
}

// OrderedProperty represents an ordered property
type OrderedProperty struct {
	Name   string
	Schema map[string]interface{}
}

// MarshalJSON implements custom JSON marshaling for OrderedSchema
func (os *OrderedSchema) MarshalJSON() ([]byte, error) {
	var buf strings.Builder
	buf.WriteString("{")

	first := true

	// type
	if os.Type != "" {
		buf.WriteString(`"type":"`)
		buf.WriteString(os.Type)
		buf.WriteString(`"`)
		first = false
	}

	// title
	if os.Title != "" {
		if !first {
			buf.WriteString(",")
		}
		buf.WriteString(`"title":"`)
		buf.WriteString(os.Title)
		buf.WriteString(`"`)
		first = false
	}

	// description
	if os.Description != "" {
		if !first {
			buf.WriteString(",")
		}
		buf.WriteString(`"description":"`)
		buf.WriteString(os.Description)
		buf.WriteString(`"`)
		first = false
	}

	// properties
	if len(os.Properties) > 0 {
		if !first {
			buf.WriteString(",")
		}
		buf.WriteString(`"properties":{`)
		for i, prop := range os.Properties {
			if i > 0 {
				buf.WriteString(",")
			}
			buf.WriteString(`"`)
			buf.WriteString(prop.Name)
			buf.WriteString(`":`)
			propJSON, err := json.Marshal(prop.Schema)
			if err != nil {
				return nil, err
			}
			buf.Write(propJSON)
		}
		buf.WriteString("}")
		first = false
	}

	// required
	if len(os.Required) > 0 {
		if !first {
			buf.WriteString(",")
		}
		reqJSON, err := json.Marshal(os.Required)
		if err != nil {
			return nil, err
		}
		buf.WriteString(`"required":`)
		buf.Write(reqJSON)
	}

	buf.WriteString("}")
	return []byte(buf.String()), nil
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
