LOCAL_BIN := $(CURDIR)/bin

.bin-deps: export GOBIN := $(LOCAL_BIN)
.bin-deps:
	$(info Installing binary dependencies...)

	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
#   go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest

# Генерация .go файлов с помощью protoc
.protoc-generate:
	protoc --proto_path=. \
	--go_out=./pkg --go_opt paths=source_relative \
	--go-grpc_out=./pkg --go-grpc_opt paths=source_relative \
	./api/orchestrator.proto \
	./api/messages.proto

# For grpc-gateway:
# protoc --proto_path=. \
# 	--go_out=./pkg --go_opt paths=source_relative \
# 	--go-grpc_out=./pkg --go-grpc_opt paths=source_relative \
# 	--grpc-gateway_out=./pkg --grpc-gateway_opt path=source_relative --grpc-gateway_opt generate_unbound_methods=true \
# 	./api/notes/service.proto \
# 	./api/notes/messages.proto

# go mod tidy
.tidy: 
	go mod tidy

.run-orchestrator:
	@go run ./cmd/orchestrator/main.go

.run-agent:
	@go run ./cmd/agent/main.go

# Генерация кода из protobuf
generate: .bin-deps .protoc-generate .tidy

orchestrator: .run-orchestrator

agent: .run-a