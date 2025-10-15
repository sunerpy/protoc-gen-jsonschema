package main

import (
	"encoding/json"
	"fmt"
	"log"

	pb "example/pb"
	"github.com/sunerpy/protoc-gen-jsonschema"
)

func main() {
	// 创建一个示例消息
	user := &pb.UserRequest{
		Email: "user@example.com",
		Name:  "John Doe",
		Age:   25,
		Phone: "+1234567890",
	}
	// 从消息生成 JSON Schema
	schema, err := jsonschema.GenerateFromMessage(user)
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
