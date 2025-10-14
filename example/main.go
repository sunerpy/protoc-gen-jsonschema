package main

import (
	"fmt"
	"log"

	"github.com/sunerpy/protoc-gen-jsonschema"
	"github.com/sunerpy/protoc-gen-jsonschema/example/pb"
)

func main() {
	// 创建一个示例消息
	user := &pb.UserRequest{
		Email: "user@example.com",
		Name:  "John Doe",
		Age:   25,
	}

	// 从消息生成 JSON Schema
	schema, err := jsonschema.GenerateFromMessage(user)
	if err != nil {
		log.Fatalf("Failed to generate schema: %v", err)
	}

	// 转换为 JSON 字符串
	jsonStr, err := schema.ToJSON()
	if err != nil {
		log.Fatalf("Failed to convert to JSON: %v", err)
	}

	fmt.Println("Generated JSON Schema:")
	fmt.Println(jsonStr)
}
