#!/bin/bash

# Proto generation script for OFA

set -e

PROTO_DIR="src/protocol/proto"
GO_OUT="src/center/proto"
JAVA_OUT="src/sdk/java/src/main/java/com/ofa/proto"
KOTLIN_OUT="src/sdk/kotlin/src/main/kotlin/com/ofa/proto"
PYTHON_OUT="src/sdk/python/ofa/proto"
JS_OUT="src/sdk/javascript/src/proto"

echo "Generating protobuf files..."

# Go
mkdir -p "$GO_OUT"
protoc --proto_path="$PROTO_DIR" \
  --go_out="$GO_OUT" \
  --go_opt=paths=source_relative \
  --go-grpc_out="$GO_OUT" \
  --go-grpc_opt=paths=source_relative \
  "$PROTO_DIR/ofa.proto"

echo "Go protobuf files generated at $GO_OUT"

# Java
mkdir -p "$JAVA_OUT"
protoc --proto_path="$PROTO_DIR" \
  --java_out="$JAVA_OUT" \
  --grpc-java_out="$JAVA_OUT" \
  "$PROTO_DIR/ofa.proto"

echo "Java protobuf files generated at $JAVA_OUT"

# Python
mkdir -p "$PYTHON_OUT"
protoc --proto_path="$PROTO_DIR" \
  --python_out="$PYTHON_OUT" \
  --grpc_python_out="$PYTHON_OUT" \
  "$PROTO_DIR/ofa.proto"

echo "Python protobuf files generated at $PYTHON_OUT"

# JavaScript
mkdir -p "$JS_OUT"
protoc --proto_path="$PROTO_DIR" \
  --js_out="library=ofa,binary:$JS_OUT" \
  --grpc-web_out="import_style=commonjs,mode=grpcwebtext:$JS_OUT" \
  "$PROTO_DIR/ofa.proto"

echo "JavaScript protobuf files generated at $JS_OUT"

echo "All protobuf files generated successfully!"