# protoc-gen-jsonschema

> 用 Protobuf 定义 MCP 工具 schema。将 Protocol Buffers 消息转换为 JSON Schema —— 既可作为 Go 库（动态生成），也可作为 protoc/buf 插件（静态生成）。

[![BSR](https://img.shields.io/badge/buf.build-sunerpy%2Fprotoc--gen--jsonschema-blue)](https://buf.build/sunerpy/protoc-gen-jsonschema)
[![Go Reference](https://pkg.go.dev/badge/github.com/sunerpy/protoc-gen-jsonschema.svg)](https://pkg.go.dev/github.com/sunerpy/protoc-gen-jsonschema)

简体中文 · [English](../../README.md)

## 目录

- [功能特性](#功能特性)
- [安装](#安装)
- [快速开始](#快速开始)
  - [作为 Go 库（动态）](#作为-go-库动态)
  - [作为 protoc/buf 插件（静态）](#作为-protocbuf-插件静态)
- [插件参数](#插件参数)
- [Schema 选项](#schema-选项)
- [配合 LLM 使用](#配合-llm-使用)
- [性能](#性能)
- [开发](#开发)
- [许可证](#许可证)

## 功能特性

- **用 proto 定义 MCP 工具 schema** —— 直接从消息定义生成 MCP server 对外暴露的工具
  `inputSchema`（JSON Schema）。
- 通过 Protobuf 扩展选项定义 JSON Schema 约束。
- 两种模式：动态生成（库）与静态生成（插件）。
- 支持字段级和消息级选项。
- 静态输出支持 JSON 文件 **或** 零开销 Go 常量。
- 高性能：动态生成约 8 μs，静态访问约 0.3 ns。

## 安装

**一键安装（无需 Go 工具链）：**

```bash
curl -fsSL https://raw.githubusercontent.com/sunerpy/protoc-gen-jsonschema/main/scripts/install.sh | sh
```

Windows（PowerShell）：

```powershell
irm https://raw.githubusercontent.com/sunerpy/protoc-gen-jsonschema/main/scripts/install.ps1 | iex
```

指定版本或安装目录：

```bash
PGJ_VERSION=0.0.7 PGJ_INSTALL_DIR=/usr/local/bin \
  curl -fsSL https://raw.githubusercontent.com/sunerpy/protoc-gen-jsonschema/main/scripts/install.sh | sh
```

**预编译二进制（不想 pipe 到 shell）：**

从 [Releases 页面](https://github.com/sunerpy/protoc-gen-jsonschema/releases)
下载对应平台的归档包，解压后将 `protoc-gen-jsonschema` 放入 `PATH` 即可。

**Go 库：**

```bash
go get github.com/sunerpy/protoc-gen-jsonschema
```

**通过 Go 安装插件二进制：**

```bash
go install github.com/sunerpy/protoc-gen-jsonschema/cmd/protoc-gen-jsonschema@latest
```

确保 `$(go env GOPATH)/bin` 已加入 `PATH`。

**通过 Buf Schema Registry (BSR) 引入 proto 扩展：**

扩展定义已发布在
[`buf.build/sunerpy/protoc-gen-jsonschema`](https://buf.build/sunerpy/protoc-gen-jsonschema)。
在你的 `buf.yaml` 中添加依赖：

```yaml
version: v2
deps:
  - buf.build/sunerpy/protoc-gen-jsonschema
```

然后运行 `buf dep update`。

## 快速开始

### 作为 Go 库（动态）

为消息添加注解：

```protobuf
syntax = "proto3";
package example;

import "mcp/jsonschema/jsonschema.proto";

message UserRequest {
  option (mcp.jsonschema.title) = "User Request";

  string email = 1 [
    (mcp.jsonschema.required) = true,
    (mcp.jsonschema.format) = "email"
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

运行时生成：

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
	schema, err := jsonschema.GenerateFromMessage(&pb.UserRequest{})
	if err != nil {
		log.Fatal(err)
	}
	out, _ := json.MarshalIndent(schema, "", "  ")
	fmt.Println(string(out))
}
```

### 作为 protoc/buf 插件（静态）

**JSON 文件（默认）：**

```bash
protoc --jsonschema_out=. user.proto
```

**Go 常量（零开销，推荐用于高频路径）：**

```bash
protoc --go_out=. --jsonschema_out=format=go_const:. user.proto
```

`buf.gen.yaml`：

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
      - format=go_const
      - paths=source_relative
```

生成的 `*_jsonschema.pb.go` 提供零开销访问方法：

```go
msg := &pb.UserRequest{}
schema := msg.GetJSONSchema()         // string，约 0.29 ns
bytes  := msg.GetJSONSchemaBytes()    // []byte，HTTP 友好
raw    := msg.GetJSONSchemaRawMessage() // json.RawMessage
```

## 插件参数

| 参数             | 默认值        | 说明                                                      |
| ---------------- | ------------- | --------------------------------------------------------- |
| `format`         | `json`        | 输出格式：`json` 或 `go_const`。                          |
| `suffix`         | `_jsonschema` | Go 文件后缀（仅 go_const）。                              |
| `paths`          | —             | `source_relative` 或 `import`。                           |
| `preserve_order` | `false`       | 在 schema 中保留 proto 字段顺序。                         |
| `schema_struct`  | `false`       | 额外生成 `jsonschema.Schema` 结构体字面量。               |
| `google_schema`  | `false`       | 额外生成 `github.com/google/jsonschema-go` 结构体字面量。 |

## Schema 选项

**字段选项** (`mcp.jsonschema.*`)：

| 选项                        | 类型   | 说明                                  |
| --------------------------- | ------ | ------------------------------------- |
| `required`                  | bool   | 标记字段为必填。                      |
| `description`               | string | 字段描述。                            |
| `example`                   | string | 示例值。                              |
| `format`                    | string | 格式约束（如 `email`、`date-time`）。 |
| `pattern`                   | string | 正则表达式。                          |
| `min_length` / `max_length` | int32  | 字符串长度边界。                      |
| `minimum` / `maximum`       | double | 数值边界。                            |
| `default`                   | string | 默认值（JSON 编码）。                 |
| `hidden`                    | bool   | 在 schema 中排除该字段。              |
| `json_name`                 | string | 覆盖 JSON 字段名。                    |

**消息选项** (`mcp.jsonschema.*`)：

| 选项                  | 类型   | 说明                              |
| --------------------- | ------ | --------------------------------- |
| `title`               | string | Schema 标题。                     |
| `message_description` | string | Schema 描述。                     |
| `generate_schema`     | bool   | 设为 `false` 可跳过该消息的生成。 |

> 关于 `google.protobuf.Timestamp`：当 Timestamp 作为字段出现时，生成的 schema 使用
> `oneOf`，同时接受 RFC3339 字符串和 `{seconds, nanos}` 对象。proto 运行时
> （`protojson`）本身只接受 RFC3339 字符串形式；对象分支面向通用 JSON 消费者。

## 配合 LLM 使用

[Model Context Protocol (MCP)](https://modelcontextprotocol.io) 要求每个工具暴露一个
`inputSchema` —— 描述其参数的 JSON Schema。本插件让你用 Protobuf 定义并生成该 schema，
于是 proto 消息成为工具契约的唯一真源：约束（`required`、`format`、`pattern`、
`min_length`、`minimum` 等）会转成 JSON Schema 关键字，供 LLM 据此产出合法参数。

```protobuf
message SearchToolArgs {
  option (mcp.jsonschema.title) = "search";
  option (mcp.jsonschema.message_description) = "Full-text search over the corpus";

  string query = 1 [
    (mcp.jsonschema.required) = true,
    (mcp.jsonschema.min_length) = 1,
    (mcp.jsonschema.description) = "The search query"
  ];
  int32 limit = 2 [
    (mcp.jsonschema.minimum) = 1,
    (mcp.jsonschema.maximum) = 100,
    (mcp.jsonschema.example) = "10"
  ];
}
```

在构建时生成一次 schema，然后把 `GetJSONSchemaBytes()` 交给 MCP server 作为工具的
`inputSchema` —— 无需手写、易漂移的 JSON。

一行安装后，即可让 agent 驱动插件：

```bash
curl -fsSL https://raw.githubusercontent.com/sunerpy/protoc-gen-jsonschema/main/scripts/install.sh | sh
```

- `protoc --jsonschema_out=. user.proto` —— 为消息生成 JSON Schema 文件。
- `protoc --go_out=. --jsonschema_out=format=go_const:. user.proto` —— 生成零开销
  Go 常量（`GetJSONSchema()` / `GetJSONSchemaBytes()`），适合作为 MCP 工具的 `inputSchema`。
- 运行时使用：调用 `jsonschema.GenerateFromMessage(msg)`，返回可直接 `json.Marshal`
  的 schema。

生成的 schema 是纯 JSON（机器可读）；消息或选项非法时，插件以非零码退出并在 stderr 输出诊断。

## 性能

静态访问比动态生成快约 29,000 倍，且 JSON 字符串常量 HTTP 友好、序列化开销为零。
完整对比见 [docs/readme/PERFORMANCE.zh.md](PERFORMANCE.zh.md)。

## 开发

```bash
make help        # 列出所有目标
make fmt         # 格式化 Go 与非 Go 文件
make lint        # golangci-lint
make test        # 运行测试
make buf-lint    # 校验 proto 文件
make check       # fmt-check + lint + buf-lint + test
make hooks       # 安装 git 钩子（commit 跑 fmt+lint，push 跑 test）
```

提交信息遵循 [Conventional Commits](https://www.conventionalcommits.org)
—— `feat:`、`fix:`、破坏性变更用 `feat!:` —— 用于驱动自动版本号递增与 changelog。

发版已自动化（基于 [release-please](https://github.com/googleapis/release-please)）：
合并 release PR 即打出 `v*` 标签，触发 GoReleaser 构建多平台二进制并发布 GitHub
Release，同时通过 `buf push` 将 proto 模块发布到 BSR。仓库内 changelog 位于
[`changelog/`](../../changelog/)，按大版本拆分。

## 许可证

[MIT](../../LICENSE)
