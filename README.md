# protoc-gen-jsonschema

将 Protocol Buffers 消息转换为 JSON Schema 的 Buf 模块。

## 功能特性

- 🎯 通过 Protobuf 扩展选项定义 JSON Schema 约束
- 📦 作为 Buf 模块发布到 BSR (buf.build/sunerpy/protoc-gen-jsonschema)
- 🔧 支持字段级和消息级的 JSON Schema 选项
- ✨ 简单易用的 API

## 安装

在你的项目中添加依赖：

```bash
# 在 buf.yaml 中添加依赖
buf dep update
```

在 `buf.yaml` 中配置：

```yaml
version: v2
deps:
  - buf.build/sunerpy/protoc-gen-jsonschema
```

## 使用方法

### 1. 在 Proto 文件中使用扩展选项

```protobuf
syntax = "proto3";

package example;

import "mcp/jsonschema/jsonschema.proto";

message UserRequest {
  option (mcp.jsonschema.title) = "User Request";
  option (mcp.jsonschema.message_description) = "Request to create or update a user";

  string email = 1 [
    (mcp.jsonschema.required) = true,
    (mcp.jsonschema.format) = "email",
    (mcp.jsonschema.description) = "User's email address"
  ];
  
  string name = 2 [
    (mcp.jsonschema.required) = true,
    (mcp.jsonschema.min_length) = 3,
    (mcp.jsonschema.max_length) = 50
  ];
  
  int32 age = 3 [
    (mcp.jsonschema.minimum) = 18,
    (mcp.jsonschema.maximum) = 120
  ];
}
```

### 2. 在 Go 代码中生成 JSON Schema

```go
package main

import (
    "encoding/json"
    "fmt"
    
    pb "your-module/pb"
    "github.com/sunerpy/protoc-gen-jsonschema/generator"
)

func main() {
    user := &pb.UserRequest{}
    schema, err := generator.GenerateSchema(user)
    if err != nil {
        panic(err)
    }
    
    jsonBytes, _ := json.MarshalIndent(schema, "", "  ")
    fmt.Println(string(jsonBytes))
}
```

## 支持的选项

### 字段选项 (FieldOptions)

- `required` (bool) - 标记字段为必填
- `description` (string) - 字段描述
- `example` (string) - 示例值
- `format` (string) - 格式约束 (如 "email", "uri")
- `pattern` (string) - 正则表达式模式
- `min_length` (int32) - 最小长度
- `max_length` (int32) - 最大长度
- `minimum` (double) - 最小值
- `maximum` (double) - 最大值
- `exclusive_minimum` (double) - 独占最小值
- `exclusive_maximum` (double) - 独占最大值
- `multiple_of` (double) - 倍数约束

### 消息选项 (MessageOptions)

- `title` (string) - Schema 标题
- `message_description` (string) - Schema 描述
- `additional_properties` (bool) - 是否允许额外属性

## 示例输出

```json
{
  "type": "object",
  "title": "User Request",
  "description": "Request to create or update a user",
  "properties": {
    "email": {
      "type": "string",
      "format": "email",
      "description": "User's email address"
    },
    "name": {
      "type": "string",
      "minLength": 3,
      "maxLength": 50
    },
    "age": {
      "type": "integer",
      "minimum": 18,
      "maximum": 120
    }
  },
  "required": ["email", "name"]
}
```

## 运行示例

```bash
cd example
buf generate
go run main.go
```

## 许可证

MIT License
