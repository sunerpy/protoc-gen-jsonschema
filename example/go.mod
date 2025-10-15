module example

go 1.25.2

require (
	github.com/google/jsonschema-go v0.3.0
	github.com/sunerpy/protoc-gen-jsonschema v0.0.0
	google.golang.org/protobuf v1.36.10
)

replace github.com/sunerpy/protoc-gen-jsonschema => ../
