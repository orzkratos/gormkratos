ERROR_PROTO_FILES=$(shell find errorspb -name *.proto)

.PHONY: errors
# generate errors code
errors:
	protoc --proto_path=./errorspb \
		   --proto_path=./proto3ps \
		   --go_out=paths=source_relative:./errorspb \
		   --go-errors_out=paths=source_relative:./errorspb \
		   $(ERROR_PROTO_FILES)
