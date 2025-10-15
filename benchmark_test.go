package jsonschema_test

import (
	"encoding/json"
	"os"
	"testing"

	pb "example/pb"
	"github.com/sunerpy/protoc-gen-jsonschema"
)

// BenchmarkDynamicGeneration 测试动态生成性能
func BenchmarkDynamicGeneration(b *testing.B) {
	user := &pb.UserRequest{
		Email: "user@example.com",
		Name:  "John Doe",
		Age:   25,
		Phone: "+1234567890",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := jsonschema.GenerateFromMessage(user)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDynamicGenerationWithJSON 测试动态生成并转换为 JSON
func BenchmarkDynamicGenerationWithJSON(b *testing.B) {
	user := &pb.UserRequest{
		Email: "user@example.com",
		Name:  "John Doe",
		Age:   25,
		Phone: "+1234567890",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := jsonschema.GenerateJSONFromMessage(user)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkStaticSchemaAccess 测试静态 Schema 访问性能（模拟从文件读取）
func BenchmarkStaticSchemaAccess(b *testing.B) {
	// 预先生成一个静态 schema（模拟编译时生成的结果）
	staticSchema := map[string]interface{}{
		"type":  "object",
		"title": "User Request",
		"description": "Request to create or update a user",
		"properties": map[string]interface{}{
			"email": map[string]interface{}{
				"type":        "string",
				"format":      "email",
				"description": "User's email address",
				"example":     "user@example.com",
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "User's full name",
				"example":     "John Doe",
				"minLength":   float64(3),
				"maxLength":   float64(50),
			},
			"age": map[string]interface{}{
				"type":        "integer",
				"description": "User's age in years",
				"minimum":     float64(18),
				"maximum":     float64(120),
			},
			"phone": map[string]interface{}{
				"type":        "string",
				"description": "User's phone number in E.164 format",
				"example":     "+1234567890",
				"pattern":     "^\\+?[1-9]\\d{1,14}$",
			},
		},
		"required": []interface{}{"email", "name"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = staticSchema
	}
}

// BenchmarkStaticSchemaAccessWithJSON 测试静态 Schema 转换为 JSON
func BenchmarkStaticSchemaAccessWithJSON(b *testing.B) {
	staticSchema := map[string]interface{}{
		"type":  "object",
		"title": "User Request",
		"description": "Request to create or update a user",
		"properties": map[string]interface{}{
			"email": map[string]interface{}{
				"type":        "string",
				"format":      "email",
				"description": "User's email address",
				"example":     "user@example.com",
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "User's full name",
				"example":     "John Doe",
				"minLength":   float64(3),
				"maxLength":   float64(50),
			},
			"age": map[string]interface{}{
				"type":        "integer",
				"description": "User's age in years",
				"minimum":     float64(18),
				"maximum":     float64(120),
			},
			"phone": map[string]interface{}{
				"type":        "string",
				"description": "User's phone number in E.164 format",
				"example":     "+1234567890",
				"pattern":     "^\\+?[1-9]\\d{1,14}$",
			},
		},
		"required": []interface{}{"email", "name"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(staticSchema)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkStaticSchemaFromFile 测试从文件读取静态 Schema
func BenchmarkStaticSchemaFromFile(b *testing.B) {
	// 创建临时 schema 文件
	schemaData := `{
  "example.UserRequest": {
    "type": "object",
    "title": "User Request",
    "description": "Request to create or update a user",
    "properties": {
      "email": {
        "type": "string",
        "format": "email",
        "description": "User's email address",
        "example": "user@example.com"
      },
      "name": {
        "type": "string",
        "description": "User's full name",
        "example": "John Doe",
        "minLength": 3,
        "maxLength": 50
      },
      "age": {
        "type": "integer",
        "description": "User's age in years",
        "minimum": 18,
        "maximum": 120
      },
      "phone": {
        "type": "string",
        "description": "User's phone number in E.164 format",
        "example": "+1234567890",
        "pattern": "^\\+?[1-9]\\d{1,14}$"
      }
    },
    "required": ["email", "name"]
  }
}`

	tmpfile, err := os.CreateTemp("", "schema-*.json")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(schemaData)); err != nil {
		b.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, err := os.ReadFile(tmpfile.Name())
		if err != nil {
			b.Fatal(err)
		}

		var schemas map[string]interface{}
		if err := json.Unmarshal(data, &schemas); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGoStructConstant 测试 Go 结构体常量方式（编译时生成）
func BenchmarkGoStructConstant(b *testing.B) {
	// 模拟编译时生成的 Go 结构体常量
	type JSONSchema struct {
		Type        string                 `json:"type"`
		Title       string                 `json:"title"`
		Description string                 `json:"description"`
		Properties  map[string]interface{} `json:"properties"`
		Required    []string               `json:"required"`
	}

	staticSchema := JSONSchema{
		Type:        "object",
		Title:       "User Request",
		Description: "Request to create or update a user",
		Properties: map[string]interface{}{
			"email": map[string]interface{}{
				"type":        "string",
				"format":      "email",
				"description": "User's email address",
				"example":     "user@example.com",
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "User's full name",
				"example":     "John Doe",
				"minLength":   3,
				"maxLength":   50,
			},
			"age": map[string]interface{}{
				"type":        "integer",
				"description": "User's age in years",
				"minimum":     18,
				"maximum":     120,
			},
			"phone": map[string]interface{}{
				"type":        "string",
				"description": "User's phone number in E.164 format",
				"example":     "+1234567890",
				"pattern":     "^\\+?[1-9]\\d{1,14}$",
			},
		},
		Required: []string{"email", "name"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = staticSchema
	}
}

// BenchmarkGoStructConstantWithJSON 测试 Go 结构体常量序列化为 JSON
func BenchmarkGoStructConstantWithJSON(b *testing.B) {
	type JSONSchema struct {
		Type        string                 `json:"type"`
		Title       string                 `json:"title"`
		Description string                 `json:"description"`
		Properties  map[string]interface{} `json:"properties"`
		Required    []string               `json:"required"`
	}

	staticSchema := JSONSchema{
		Type:        "object",
		Title:       "User Request",
		Description: "Request to create or update a user",
		Properties: map[string]interface{}{
			"email": map[string]interface{}{
				"type":        "string",
				"format":      "email",
				"description": "User's email address",
				"example":     "user@example.com",
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "User's full name",
				"example":     "John Doe",
				"minLength":   3,
				"maxLength":   50,
			},
			"age": map[string]interface{}{
				"type":        "integer",
				"description": "User's age in years",
				"minimum":     18,
				"maximum":     120,
			},
			"phone": map[string]interface{}{
				"type":        "string",
				"description": "User's phone number in E.164 format",
				"example":     "+1234567890",
				"pattern":     "^\\+?[1-9]\\d{1,14}$",
			},
		},
		Required: []string{"email", "name"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(staticSchema)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkStaticJSONString 测试预序列化的 JSON 字符串（最优方案）
func BenchmarkStaticJSONString(b *testing.B) {
	// 编译时生成的 JSON 字符串常量
	const staticSchemaJSON = `{"type":"object","title":"User Request","description":"Request to create or update a user","properties":{"email":{"type":"string","format":"email","description":"User's email address","example":"user@example.com"},"name":{"type":"string","description":"User's full name","example":"John Doe","minLength":3,"maxLength":50},"age":{"type":"integer","description":"User's age in years","minimum":18,"maximum":120},"phone":{"type":"string","description":"User's phone number in E.164 format","example":"+1234567890","pattern":"^\\+?[1-9]\\d{1,14}$"}},"required":["email","name"]}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = staticSchemaJSON
	}
}

// BenchmarkStaticJSONStringToBytes 测试 JSON 字符串转字节数组
func BenchmarkStaticJSONStringToBytes(b *testing.B) {
	const staticSchemaJSON = `{"type":"object","title":"User Request","description":"Request to create or update a user","properties":{"email":{"type":"string","format":"email","description":"User's email address","example":"user@example.com"},"name":{"type":"string","description":"User's full name","example":"John Doe","minLength":3,"maxLength":50},"age":{"type":"integer","description":"User's age in years","minimum":18,"maximum":120},"phone":{"type":"string","description":"User's phone number in E.164 format","example":"+1234567890","pattern":"^\\+?[1-9]\\d{1,14}$"}},"required":["email","name"]}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = []byte(staticSchemaJSON)
	}
}

// TestGenerateSchema 功能测试
func TestGenerateSchema(t *testing.T) {
	user := &pb.UserRequest{
		Email: "user@example.com",
		Name:  "John Doe",
		Age:   25,
		Phone: "+1234567890",
	}

	schema, err := jsonschema.GenerateFromMessage(user)
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	// 验证基本字段
	if schema["type"] != "object" {
		t.Errorf("Expected type 'object', got %v", schema["type"])
	}

	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected properties to be a map")
	}

	// 验证必填字段
	required, ok := schema["required"].([]string)
	if !ok {
		t.Fatal("Expected required to be a string slice")
	}

	if len(required) != 2 {
		t.Errorf("Expected 2 required fields, got %d", len(required))
	}

	// 验证 email 字段存在
	if _, ok := properties["email"]; !ok {
		t.Fatal("Expected email property to exist")
	}

	// 打印 schema 用于调试
	jsonBytes, _ := json.MarshalIndent(schema, "", "  ")
	t.Logf("Generated schema:\n%s", string(jsonBytes))
}
