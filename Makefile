.PHONY: all proto clean

all: proto

proto:
	@echo "Generating protobuf Go code..."
	PATH=$(PATH):$(shell go env GOPATH)/bin protoc --go_out=. --go_opt=paths=source_relative pkg/proto/net/message.proto
	@echo "Protobuf Go code generation complete."

clean:
	@echo "Cleaning generated protobuf Go code..."
	rm -f pkg/proto/net/message.pb.go
	@echo "Clean complete."