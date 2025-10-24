ERROR_PROTO_FILES = $(shell find errorspb -name "*.proto")

.PHONY: install
install:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install github.com/orzkratos/errgenkratos/cmd/protoc-gen-orzkratos-errors@latest
	@echo "All protoc plugins installed!"

.PHONY: generate
generate:
	protoc --proto_path=./errorspb \
		   --proto_path=./proto3ps \
		   --go_out=paths=source_relative:./errorspb \
		   --orzkratos-errors_out=paths=source_relative:./errorspb \
		   $(ERROR_PROTO_FILES)
	@echo "Proto code generation complete!"

.PHONY: clean
clean:
	rm -f errorspb/*.pb.go
	@echo "Cleanup complete!"

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  install  - Install protoc plugins (go, orzkratos-errors)"
	@echo "  generate - Generate Go code from proto files"
	@echo "  clean    - Remove generated files"
	@echo "  help     - Show this help message"
