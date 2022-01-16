export PATH := $(PATH):$(go env GOPATH)/bin
GO_BUILD_VARS := CGO_ENABLED=0 GOOS=linux GOARCH=amd64
GO_LINK_FLAGS := -ldflags="-s -w"
PROTOS := linter_proto/linter_proto.pb.go linter_proto/linter_proto_grpc.pb.go

all: containers
containers: machine_manager_container loadbalancer_container python_linter_container

%_container: Dockerfile %_bin
	docker build . --tag $@ --target $*

%.pb.go: %.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative $<

%_bin: %/main.go $(PROTOS)
	cd $* && $(GO_BUILD_VARS) go build $(GO_LINK_FLAGS) -o ../$@

clean:
	rm *_bin

linter_proto/linter_proto.pb.go: linter_proto/linter_proto.proto
linter_proto/linter_proto_grpc.pb.go: linter_proto/linter_proto.proto
