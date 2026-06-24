# protoc-gen-jsonschema

> Convert Protocol Buffers messages into JSON Schema — as a Go library (dynamic) or a protoc/buf plugin (static).

[![BSR](https://img.shields.io/badge/buf.build-sunerpy%2Fprotoc--gen--jsonschema-blue)](https://buf.build/sunerpy/protoc-gen-jsonschema)
[![Go Reference](https://pkg.go.dev/badge/github.com/sunerpy/protoc-gen-jsonschema.svg)](https://pkg.go.dev/github.com/sunerpy/protoc-gen-jsonschema)

[简体中文](docs/readme/README.zh.md) · English

## Table of Contents

- [Features](#features)
- [Install](#install)
- [Quick start](#quick-start)
  - [As a Go library (dynamic)](#as-a-go-library-dynamic)
  - [As a protoc/buf plugin (static)](#as-a-protocbuf-plugin-static)
- [Plugin options](#plugin-options)
- [Schema options](#schema-options)
- [Performance](#performance)
- [Development](#development)
- [License](#license)

## Features

- Define JSON Schema constraints with Protobuf extension options.
- Two modes: dynamic generation (library) and static generation (plugin).
- Field-level and message-level options.
- Static output as JSON files **or** zero-overhead Go constants.
- High performance: ~8 μs dynamic, ~0.3 ns static access.

## Install

**Go library:**

```bash
go get github.com/sunerpy/protoc-gen-jsonschema
```

**Plugin binary:**

```bash
go install github.com/sunerpy/protoc-gen-jsonschema/cmd/protoc-gen-jsonschema@latest
```

Make sure `$(go env GOPATH)/bin` is on your `PATH`.

**Proto extensions via the Buf Schema Registry (BSR):**

The extension definitions are published at
[`buf.build/sunerpy/protoc-gen-jsonschema`](https://buf.build/sunerpy/protoc-gen-jsonschema).
Add it as a dependency in your `buf.yaml`:

```yaml
version: v2
deps:
  - buf.build/sunerpy/protoc-gen-jsonschema
```

Then run `buf dep update`.

## Quick start

### As a Go library (dynamic)

Annotate a message:

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

Generate at runtime:

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

### As a protoc/buf plugin (static)

**JSON files (default):**

```bash
protoc --jsonschema_out=. user.proto
```

**Go constants (zero-overhead, recommended for hot paths):**

```bash
protoc --go_out=. --jsonschema_out=format=go_const:. user.proto
```

`buf.gen.yaml`:

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

The generated `*_jsonschema.pb.go` exposes zero-overhead accessors:

```go
msg := &pb.UserRequest{}
schema := msg.GetJSONSchema()         // string, ~0.29 ns
bytes  := msg.GetJSONSchemaBytes()    // []byte, HTTP-ready
raw    := msg.GetJSONSchemaRawMessage() // json.RawMessage
```

## Plugin options

| Option           | Default       | Description                                                   |
| ---------------- | ------------- | ------------------------------------------------------------- |
| `format`         | `json`        | Output format: `json` or `go_const`.                          |
| `suffix`         | `_jsonschema` | Go file suffix (go_const only).                               |
| `paths`          | —             | `source_relative` or `import`.                                |
| `preserve_order` | `false`       | Preserve proto field order in the schema.                     |
| `schema_struct`  | `false`       | Also emit a `jsonschema.Schema` struct literal.               |
| `google_schema`  | `false`       | Also emit a `github.com/google/jsonschema-go` struct literal. |

## Schema options

**Field options** (`mcp.jsonschema.*`):

| Option                      | Type   | Description                                    |
| --------------------------- | ------ | ---------------------------------------------- |
| `required`                  | bool   | Mark the field as required.                    |
| `description`               | string | Field description.                             |
| `example`                   | string | Example value.                                 |
| `format`                    | string | Format constraint (e.g. `email`, `date-time`). |
| `pattern`                   | string | Regular expression.                            |
| `min_length` / `max_length` | int32  | String length bounds.                          |
| `minimum` / `maximum`       | double | Numeric bounds.                                |
| `default`                   | string | Default value (JSON-encoded).                  |
| `hidden`                    | bool   | Exclude the field from the schema.             |
| `json_name`                 | string | Override the JSON field name.                  |

**Message options** (`mcp.jsonschema.*`):

| Option                | Type   | Description                                      |
| --------------------- | ------ | ------------------------------------------------ |
| `title`               | string | Schema title.                                    |
| `message_description` | string | Schema description.                              |
| `generate_schema`     | bool   | Set `false` to skip generation for this message. |

> Note on `google.protobuf.Timestamp`: when a Timestamp appears as a field, the
> generated schema uses a `oneOf` accepting both an RFC3339 string and a
> `{seconds, nanos}` object. The proto runtime (`protojson`) itself only accepts
> the RFC3339 string form; the object branch targets generic JSON consumers.

## Performance

Static access is ~29,000× faster than dynamic generation, and a JSON string
constant is HTTP-ready with zero serialization cost. See
[docs/PERFORMANCE.md](docs/PERFORMANCE.md) for the full comparison.

## Development

```bash
make help        # list targets
make fmt         # format Go + non-Go files
make lint        # golangci-lint
make test        # run tests
make buf-lint    # lint proto files
make check       # fmt-check + lint + buf-lint + test
make hooks       # install pre-commit hooks
```

Releases are automated: pushing a `v*` tag triggers CI and publishes the module
to the BSR via `buf push`.

## License

[MIT](LICENSE)
