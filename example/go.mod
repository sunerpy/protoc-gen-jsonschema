module example

go 1.26.4

require (
	github.com/google/jsonschema-go v0.4.3
	github.com/sunerpy/protoc-gen-jsonschema v0.0.0
	google.golang.org/protobuf v1.36.11
)

replace github.com/sunerpy/protoc-gen-jsonschema => ../
