# protoc-gen-jsonschema

将 Protocol Buffers 消息转换为 JSON Schema 的工具库，支持动态生成和静态生成两种方式。

## 功能特性

- 🎯 通过 Protobuf 扩展选项定义 JSON Schema 约束
- � 两种使用方式：动态生成（库）和静态生成（protoc 插件）
- 📦 作为 Go 库使用，无需额外构建步骤
- 🔧 支持字段级和消息级的 JSON Schema 选项
- ⚡ 高性能：动态生成 ~8μs，静态访问 ~0.3ns
- ✨ 简单易用的 API

## 性能对比

| 方案 | 延迟 | 吞吐量 | 内存占用 | 适用场景 |
|------|------|--------|---------|---------|
| **动态生成** | ~8.4 μs | ~118K ops/s | 2.7 KB/op | 推荐用于大多数场景 |
| **静态 map** | ~0.29 ns + 5.9 μs | ~170K ops/s | 2.5 KB/op | 需要 map 结构 |
| **Go 结构体** | ~0.29 ns + 4.7 μs | ~212K ops/s | 2.2 KB/op | 需要类型安全 |
| **JSON 字符串** 🏆 | ~0.29 ns | 无限制 | 0 | **最优方案** |

**关键发现**：
- 静态方案访问速度比动态生成快 **29,000 倍**
- JSON 字符串常量是 **最优方案**：零开销、直接可用、HTTP 友好
- 完整对比分析请查看 [PERFORMANCE_COMPARISON.md](PERFORMANCE_COMPARISON.md)

## 使用建议

### 🎯 选择合适的方案

根据您的应用场景选择最合适的方案：

#### 推荐使用动态生成（99% 的场景）

**适用条件**：
- ✅ QPS < 10,000
- ✅ 需要运行时灵活性
- ✅ 希望代码简洁易维护
- ✅ 作为第三方库使用

**性能表现**：
- 单次生成耗时：~8.1 微秒
- CPU 开销示例：1,000 req/s = 0.81% CPU
- 内存占用：每次 2.7 KB

**示例场景**：
- Web API 服务器（< 10K QPS）
- 微服务内部调用
- 开发和测试环境
- 消息类型较少的应用

#### 使用静态生成（高性能场景）

**适用条件**：
- ⚡ QPS > 100,000
- ⚡ 需要极致性能
- ⚡ 消息类型固定且较多
- ⚡ 可以接受额外构建步骤

**性能表现**：
- 访问耗时：~0.29 纳秒（几乎为 0）
- CPU 开销：几乎为 0
- 内存占用：Schema 常驻内存

**示例场景**：
- 高性能网关（> 50K QPS）
- 消息验证服务（> 100K msg/s）
- 实时数据处理系统
- 需要生成独立 Schema 文件

### 📊 性能对比实例

| 场景 | QPS | 动态方案 CPU 开销 | 静态方案 CPU 开销 | 推荐方案 |
|------|-----|------------------|------------------|---------|
| Web API | 1,000 | 0.81% | ~0% | ✅ 动态生成 |
| 微服务 | 5,000 | 4.1% | ~0% | ✅ 动态生成 |
| API 网关 | 50,000 | 40.5% | ~0% | ⚠️ 静态生成 |
| 消息队列 | 500,000 | 405% (不可行) | ~0% | ❌ 必须静态生成 |

### 💡 最优方案推荐

#### 🏆 方案对比总结

根据完整的性能测试，我们发现了 **4 种静态生成方式**：

| 静态方案 | 访问耗时 | 序列化耗时 | 总耗时 | 适用场景 |
|---------|---------|-----------|--------|---------|
| **JSON 字符串常量** 🏆 | 0.29 ns | 0 ns | **0.29 ns** | HTTP API（最优） |
| **Go 结构体常量** | 0.29 ns | 4,708 ns | 4,708 ns | 需要类型安全 |
| **map 常量** | 0.29 ns | 5,875 ns | 5,875 ns | 需要动态操作 |
| **从文件读取** | 13,279 ns | 0 ns | 13,279 ns | 不推荐 |

#### 🎯 最优实践：生成 JSON 字符串常量

**为什么是最优方案？**
1. ✅ **极致性能**：访问耗时 0.29 ns，无需序列化
2. ✅ **零内存分配**：无 GC 压力
3. ✅ **HTTP 友好**：可直接作为响应体
4. ✅ **简单直接**：无需额外处理

**使用方法**：

##### 方式 A：使用 protoc 插件自动生成（推荐）

```bash
# 1. 生成 Go 代码和 JSON Schema 常量
protoc --go_out=. --jsonschema_out=format=go_const:. user.proto

# 或使用 buf
buf generate
```

生成的代码结构：
```
example/
├── pb/
│   ├── user.pb.go              # protobuf 生成的代码
│   └── user_jsonschema.pb.go   # JSON Schema 常量（新增）
```

生成的 `user_jsonschema.pb.go` 文件：
```go
package pb

// GetJSONSchema 返回 UserRequest 的 JSON Schema（零开销）
func (*UserRequest) GetJSONSchema() string {
    return userRequestSchemaJSON
}

// GetJSONSchemaBytes 返回 JSON Schema 的字节数组（用于 HTTP 响应）
func (*UserRequest) GetJSONSchemaBytes() []byte {
    return []byte(userRequestSchemaJSON)
}

const userRequestSchemaJSON = `{"type":"object","title":"User Request","description":"Request to create or update a user","properties":{"email":{"type":"string","format":"email","description":"User's email address","example":"user@example.com"},"name":{"type":"string","description":"User's full name","example":"John Doe","minLength":3,"maxLength":50},"age":{"type":"integer","description":"User's age in years","minimum":18,"maximum":120},"phone":{"type":"string","description":"User's phone number in E.164 format","example":"+1234567890","pattern":"^\\+?[1-9]\\d{1,14}$"}},"required":["email","name"]}`
```

在代码中使用：
```go
package main

import (
    "net/http"
    pb "example/pb"
)

func GetSchemaHandler(w http.ResponseWriter, r *http.Request) {
    msg := &pb.UserRequest{}
    
    // 方式 1：直接获取字符串（0.29 ns）
    schemaJSON := msg.GetJSONSchema()
    
    // 方式 2：获取字节数组用于 HTTP 响应（0.34 ns）
    w.Header().Set("Content-Type", "application/json")
    w.Write(msg.GetJSONSchemaBytes()) // 零开销！
}
```

##### 方式 B：手动创建（临时方案）

如果 protoc 插件尚未支持 `format=go_const` 选项，可以手动创建：

```go
package pb

// 手动添加到 user.pb.go 或创建 user_jsonschema.go
func (*UserRequest) GetJSONSchema() string {
    return userRequestSchemaJSON
}

const userRequestSchemaJSON = `{"type":"object",...}`
```

**性能对比**：
- 动态生成：8,440 ns + 序列化开销
- JSON 字符串常量：**0.29 ns**（快 29,000 倍！）
- HTTP 响应（字节数组）：**0.34 ns**（快 24,800 倍！）

**buf.gen.yaml 配置示例**：
```yaml
version: v2
plugins:
  - remote: buf.build/protocolbuffers/go
    out: pb
    opt:
      - paths=source_relative
  
  - local: protoc-gen-jsonschema
    out: pb
    opt:
      - format=go_const        # 生成 Go 常量
      - paths=source_relative
```

#### 🔧 其他场景的优化建议

**场景 1：需要动态操作 Schema（如合并、修改）**
```go
// 使用 Go 结构体常量
type JSONSchema struct {
    Type       string                 `json:"type"`
    Properties map[string]interface{} `json:"properties"`
}

var UserRequestSchema = JSONSchema{...}

// 可以动态修改
func customizeSchema(base JSONSchema) JSONSchema {
    base.Properties["custom_field"] = map[string]interface{}{
        "type": "string",
    }
    return base
}
```

**场景 2：动态生成 + 缓存（中等 QPS）**
```go
var schemaCache sync.Map

func getSchemaJSON(msg proto.Message) (string, error) {
    key := string(msg.ProtoReflect().Descriptor().FullName())
    if cached, ok := schemaCache.Load(key); ok {
        return cached.(string), nil
    }
    
    schema, err := jsonschema.GenerateFromMessage(msg)
    if err != nil {
        return "", err
    }
    
    jsonBytes, err := json.Marshal(schema)
    if err != nil {
        return "", err
    }
    
    jsonStr := string(jsonBytes)
    schemaCache.Store(key, jsonStr)
    return jsonStr, nil
}
```

**场景 3：启动预热（低 QPS，偶尔使用）**
```go
func init() {
    // 预热常用消息类型
    jsonschema.GenerateFromMessage(&pb.UserRequest{})
    jsonschema.GenerateFromMessage(&pb.OrderRequest{})
}
```

## 安装

```bash
go get github.com/sunerpy/protoc-gen-jsonschema
```

## 使用方法

### 方式 1: 动态生成（推荐）

适用于大多数场景，无需额外构建步骤。

#### 1. 在 Proto 文件中使用扩展选项

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
    (mcp.jsonschema.description) = "User's email address",
    (mcp.jsonschema.example) = "user@example.com"
  ];
  
  string name = 2 [
    (mcp.jsonschema.required) = true,
    (mcp.jsonschema.description) = "User's full name",
    (mcp.jsonschema.min_length) = 3,
    (mcp.jsonschema.max_length) = 50,
    (mcp.jsonschema.example) = "John Doe"
  ];
  
  int32 age = 3 [
    (mcp.jsonschema.description) = "User's age in years",
    (mcp.jsonschema.minimum) = 18,
    (mcp.jsonschema.maximum) = 120
  ];

  string phone = 4 [
    (mcp.jsonschema.description) = "User's phone number in E.164 format",
    (mcp.jsonschema.pattern) = "^\\+?[1-9]\\d{1,14}$",
    (mcp.jsonschema.example) = "+1234567890"
  ];
}
```

#### 2. 在 Go 代码中生成 JSON Schema

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    
    pb "example/pb"
    "github.com/sunerpy/protoc-gen-jsonschema"
)

func main() {
    // 创建消息实例
    user := &pb.UserRequest{
        Email: "user@example.com",
        Name:  "John Doe",
        Age:   25,
        Phone: "+1234567890",
    }
    
    // 生成 JSON Schema
    schema, err := jsonschema.GenerateFromMessage(user)
    if err != nil {
        log.Fatal(err)
    }
    
    // 转换为 JSON 字符串
    jsonBytes, err := json.MarshalIndent(schema, "", "  ")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(string(jsonBytes))
}
```

### 方式 2: 静态生成（高性能场景）

适用于高 QPS 场景（> 100K），通过 protoc 插件生成静态 JSON Schema 文件。

#### 1. 安装 protoc 插件

有两种方式安装插件：

**方式 A：从 GitHub 安装（推荐）**

```bash
# 安装最新版本
go install github.com/sunerpy/protoc-gen-jsonschema/cmd/protoc-gen-jsonschema@latest

# 或安装指定版本
go install github.com/sunerpy/protoc-gen-jsonschema/cmd/protoc-gen-jsonschema@v1.0.0
```

**方式 B：从源码构建**

```bash
# 克隆仓库
git clone https://github.com/sunerpy/protoc-gen-jsonschema.git
cd protoc-gen-jsonschema

# 构建并安装
go install ./cmd/protoc-gen-jsonschema
```

**验证安装**

```bash
# 检查插件是否在 PATH 中
which protoc-gen-jsonschema

# 查看版本
protoc-gen-jsonschema --version
```

**注意事项**：
- 确保 `$GOPATH/bin` 或 `$GOBIN` 在系统 PATH 中
- 默认安装路径：`$GOPATH/bin/protoc-gen-jsonschema` 或 `$HOME/go/bin/protoc-gen-jsonschema`
- 如果 `which protoc-gen-jsonschema` 找不到，需要将 Go bin 目录添加到 PATH：
  ```bash
  export PATH=$PATH:$(go env GOPATH)/bin
  ```

#### 2. 生成 JSON Schema 文件

插件支持两种输出格式：

**格式 1：JSON 文件（默认）**

生成独立的 `.schema.json` 文件，适合需要独立 Schema 文件的场景。

```bash
# 使用 protoc（默认生成 JSON 文件）
protoc --jsonschema_out=. user.proto

# 或明确指定 JSON 格式
protoc --jsonschema_out=format=json:. user.proto

# 使用 buf
buf generate
```

`buf.gen.yaml` 配置：
```yaml
version: v2
plugins:
  - local: protoc-gen-jsonschema
    out: pb
    opt:
      - format=json              # 生成 JSON 文件（默认）
      - paths=source_relative
```

**格式 2：Go 常量（推荐用于高性能场景）**

生成包含 JSON Schema 字符串常量的 Go 代码，零开销访问。

```bash
# 使用 protoc 生成 Go 常量
protoc --go_out=. --jsonschema_out=format=go_const:. user.proto

# 自定义文件后缀
protoc --jsonschema_out=format=go_const,suffix=_schema:. user.proto
```

`buf.gen.yaml` 配置：
```yaml
version: v2
plugins:
  - remote: buf.build/protocolbuffers/go
    out: pb
    opt:
      - paths=source_relative
  
  - local: protoc-gen-jsonschema
    out: pb
    opt:
      - format=go_const          # 生成 Go 常量
      - suffix=_jsonschema       # 文件后缀（默认）
      - paths=source_relative
```

**插件参数说明**：

| 参数 | 默认值 | 说明 | 示例 |
|------|--------|------|------|
| `format` | `json` | 输出格式：`json` 或 `go_const` | `format=go_const` |
| `suffix` | `_jsonschema` | Go 文件后缀（仅 go_const 格式） | `suffix=_schema` |
| `paths` | - | 路径模式：`source_relative` 或 `import` | `paths=source_relative` |

**完整示例**：

```bash
# 同时生成 protobuf Go 代码和 JSON Schema 常量
protoc \
  --go_out=. \
  --go_opt=paths=source_relative \
  --jsonschema_out=. \
  --jsonschema_opt=format=go_const \
  --jsonschema_opt=paths=source_relative \
  user.proto
```

#### 3. 生成的文件

生成的 `user.schema.json` 文件：

```json
{
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
}
```

## 支持的选项

### 字段选项 (FieldOptions)

| 选项 | 类型 | 说明 | 示例 |
|------|------|------|------|
| `required` | bool | 标记字段为必填 | `(mcp.jsonschema.required) = true` |
| `description` | string | 字段描述 | `(mcp.jsonschema.description) = "用户邮箱"` |
| `example` | string | 示例值 | `(mcp.jsonschema.example) = "user@example.com"` |
| `format` | string | 格式约束 | `(mcp.jsonschema.format) = "email"` |
| `pattern` | string | 正则表达式 | `(mcp.jsonschema.pattern) = "^[A-Z]"` |
| `min_length` | int32 | 最小长度 | `(mcp.jsonschema.min_length) = 3` |
| `max_length` | int32 | 最大长度 | `(mcp.jsonschema.max_length) = 50` |
| `minimum` | double | 最小值 | `(mcp.jsonschema.minimum) = 0` |
| `maximum` | double | 最大值 | `(mcp.jsonschema.maximum) = 100` |
| `default` | string | 默认值（JSON） | `(mcp.jsonschema.default)
