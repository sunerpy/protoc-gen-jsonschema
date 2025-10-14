package main

import (
	"encoding/json"
	"fmt"
	"log"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"

	pb "example/pb"
	jsonschemapb "github.com/sunerpy/protoc-gen-jsonschema/mcp/jsonschema"
)

// Schema represents a JSON Schema
type Schema map[string]interface{}

// GenerateSchema generates JSON Schema from a protobuf message
func GenerateSchema(msg proto.Message) (Schema, error) {
	md := msg.ProtoReflect().Descriptor()
	msgOpts := md.Options().(*descriptorpb.MessageOptions)

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

		fieldName := field.JSONName()
		fieldSchema := generateFieldSchema(field, fieldOpts)
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
func generateFieldSchema(field protoreflect.FieldDescriptor, opts *descriptorpb.FieldOptions) Schema {
	schema := Schema{}

	// Set type based on protobuf type
	switch field.Kind() {
	case protoreflect.BoolKind:
		schema["type"] = "boolean"
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind,
		protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		schema["type"] = "integer"
	case protoreflect.StringKind:
		schema["type"] = "string"
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

func main() {
	// 创建一个示例消息
	user := &pb.UserRequest{
		Email: "user@example.com",
		Name:  "John Doe",
		Age:   25,
		Phone: "+1234567890",
	}

	// 从消息生成 JSON Schema
	schema, err := GenerateSchema(user)
	if err != nil {
		log.Fatalf("Failed to generate schema: %v", err)
	}

	// 转换为 JSON 字符串
	jsonBytes, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		log.Fatalf("Failed to convert to JSON: %v", err)
	}

	fmt.Println("Generated JSON Schema:")
	fmt.Println(string(jsonBytes))
}
