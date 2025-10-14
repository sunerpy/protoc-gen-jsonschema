# protoc-gen-jsonschema

[English](#english) | [中文](#中文)

---

## English

A Buf module providing Protocol Buffer extension options for JSON Schema validation, designed for MCP (Model Context Protocol) integration.

### Features

- ✅ **Field-level Options** - Rich set of validation options for individual fields
- ✅ **Message-level Options** - Control schema generation at message level
- ✅ **Standard Compliance** - Follows JSON Schema specification
- ✅ **Easy Integration** - Simple import and use in your proto files
- ✅ **MCP Ready** - Perfect for Model Context Protocol tools

### Installation

Add the module as a dependency in your `buf.yaml`:

```yaml
version: v2
deps:
  - buf.build/sunerpy/protoc-gen-jsonschema
```

Update dependencies:

```bash
buf dep update
```

### Quick Start

Import the options in your `.proto` file:

```protobuf
syntax = "proto3";

import "mcp/jsonschema/jsonschema.proto";

message UserRequest {
  // User's email address
  string email = 1 [
    (mcp.jsonschema.required) = true,
    (mcp.jsonschema.format) = "email"
  ];
  
  // User's name (3-50 characters)
  string name = 2 [
    (mcp.jsonschema.required) = true,
    (mcp.jsonschema.min_length) = 3,
    (mcp.jsonschema.max_length) = 50
  ];
  
  // User's age (18-120)
  int32 age = 3 [
    (mcp.jsonschema.minimum) = 18,
    (mcp.jsonschema.maximum) = 120
  ];
}
```

### Available Options

#### Field Options

| Option | Type | Description | Example |
|--------|------|-------------|---------|
| `required` | bool | Mark field as required | `true` |
| `description` | string | Custom description | `"User ID"` |
| `example` | string | Example value | `"user@example.com"` |
| `format` | string | JSON Schema format | `"email"`, `"uuid"`, `"uri"` |
| `default` | string | Default value (JSON string) | `"\"en-US\""` |
| `min_length` | int32 | Minimum length (string) | `3` |
| `max_length` | int32 | Maximum length (string) | `100` |
| `minimum` | double | Minimum value (number) | `0` |
| `maximum` | double | Maximum value (number) | `100` |
| `pattern` | string | Regex pattern | `"^[A-Z]+$"` |
| `hidden` | bool | Hide field from schema | `true` |
| `json_name` | string | Custom JSON field name | `"userId"` |

#### Message Options

| Option | Type | Description | Example |
|--------|------|-------------|---------|
| `message_description` | string | Message description | `"User registration request"` |
| `generate_schema` | bool | Whether to generate schema | `true` |
| `title` | string | Schema title | `"User Request"` |

### Use Cases

- **API Validation** - Define validation rules in protobuf
- **Documentation** - Auto-generate API documentation
- **Code Generation** - Generate validation code from proto definitions
- **MCP Integration** - Use with Model Context Protocol tools

### Module Information

- **Module Name**: `buf.build/sunerpy/protoc-gen-jsonschema`
- **Package**: `mcp.jsonschema`
- **Import Path**: `mcp/jsonschema/jsonschema.proto`
- **Go Package**: `github.com/sunerpy/protoc-gen-jsonschema/mcp/jsonschema`

### License

MIT License - see [LICENSE](LICENSE) file for details.

---

## 中文

一个提供 Protocol Buffer 扩展选项的 Buf 模块，用于 JSON Schema 验证，专为 MCP (Model Context Protocol) 集成设计。

### 特性

- ✅ **字段级选项** - 丰富的字段验证选项
- ✅ **消息级选项** - 在消息级别控制 schema 生成
- ✅ **标准兼容** - 遵循 JSON Schema 规范
- ✅ **易于集成** - 在 proto 文件中简单导入使用
- ✅ **MCP 就绪** - 完美适配 Model Context Protocol 工具

### 安装

在 `buf.yaml` 中添加模块依赖：

```yaml
version: v2
deps:
  - buf.build/sunerpy/protoc-gen-jsonschema
```

更新依赖：

```bash
buf dep update
```

### 快速开始

在 `.proto` 文件中导入选项：

```protobuf
syntax = "proto3";

import "mcp/jsonschema/jsonschema.proto";

message UserRequest {
  // 用户邮箱
  string email = 1 [
    (mcp.jsonschema.required) = true,
    (mcp.jsonschema.format) = "email"
  ];
  
  // 用户名（3-50 字符）
  string name = 2 [
    (mcp.jsonschema.required) = true,
    (mcp.jsonschema.min_length) = 3,
    (mcp.jsonschema.max_length) = 50
  ];
  
  // 用户年龄（18-120）
  int32 age = 3 [
    (mcp.jsonschema.minimum) = 18,
    (mcp.jsonschema.maximum) = 120
  ];
}
```

### 可用选项

#### 字段选项

| 选项 | 类型 | 说明 | 示例 |
|------|------|------|------|
| `required` | bool | 标记为必填字段 | `true` |
| `description` | string | 自定义描述 | `"用户ID"` |
| `example` | string | 示例值 | `"user@example.com"` |
| `format` | string | JSON Schema 格式 | `"email"`, `"uuid"`, `"uri"` |
| `default` | string | 默认值（JSON 字符串） | `"\"zh-CN\""` |
| `min_length` | int32 | 最小长度（字符串） | `3` |
| `max_length` | int32 | 最大长度（字符串） | `100` |
| `minimum` | double | 最小值（数值） | `0` |
| `maximum` | double | 最大值（数值） | `100` |
| `pattern` | string | 正则表达式模式 | `"^[A-Z]+$"` |
| `hidden` | bool | 隐藏字段 | `true` |
| `json_name` | string | 自定义 JSON 字段名 | `"userId"` |

#### 消息选项

| 选项 | 类型 | 说明 | 示例 |
|------|------|------|------|
| `message_description` | string | 消息描述 | `"用户注册请求"` |
| `generate_schema` | bool | 是否生成 schema | `true` |
| `title` | string | Schema 标题 | `"用户请求"` |

### 使用场景

- **API 验证** - 在 protobuf 中定义验证规则
- **文档生成** - 自动生成 API 文档
- **代码生成** - 从
