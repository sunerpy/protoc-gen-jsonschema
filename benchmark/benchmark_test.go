package benchmark_test

import (
	"encoding/json"
	"testing"
)

// 使用内联的测试消息定义，避免依赖 example/pb
type testMessage struct {
	Email string
	Name  string
	Age   int32
	Phone string
}

// BenchmarkDynamicGeneration 测试动态生成性能
func BenchmarkDynamicGeneration(b *testing.B) {
	// 注意：这个基准测试需要实际的 protobuf 消息
	// 在实际使用时，请使用真实的 protobuf 生成的消息
	b.Skip("需要实际的 protobuf 消息进行测试")
}

// BenchmarkStaticSchemaAccess 测试静态 Schema 访问性能
func BenchmarkStaticSchemaAccess(b *testing.B) {
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
		"properties": map[string]interface{}{
			"email": map[string]interface{}{
				"type":   "string",
				"format": "email",
			},
		},
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
	const staticSchemaJSON = `{"type":"object","title":"User Request","properties":{"email":{"type":"string","format":"email"}}}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = staticSchemaJSON
	}
}

// BenchmarkStaticJSONStringToBytes 测试 JSON 字符串转字节数组
func BenchmarkStaticJSONStringToBytes(b *testing.B) {
	const staticSchemaJSON = `{"type":"object","title":"User Request","properties":{"email":{"type":"string","format":"email"}}}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = []byte(staticSchemaJSON)
	}
}
