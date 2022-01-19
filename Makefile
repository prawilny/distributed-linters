export PATH := $(PATH):$(go env GOPATH)/bin
GO_BUILD_VARS := CGO_ENABLED=0 GOOS=linux GOARCH=amd64
GO_LINK_FLAGS := -ldflags="-s -w"
PROTOS := linter_proto/linter_proto.pb.go linter_proto/linter_proto_grpc.pb.go

all: image_build
image_build: machine_manager_image_build loadbalancer_image_build python_linter_image_build
image_upload: machine_manager_image_upload loadbalancer_image_upload python_linter_image_upload

%_image_build: %/Dockerfile %/%
	#minikube ssh -- sudo podman build /host --tag $@ --target $*
	docker build $*/ --tag $*
	docker tag $* europe-central2-docker.pkg.dev/irio-linter/irio/$*

%_image_upload: %_image_build
	docker push europe-central2-docker.pkg.dev/irio-linter/irio/$*

%.pb.go: %.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative $<

%/%: %/main.go $(PROTOS)
	cd $* && $(GO_BUILD_VARS) go build $(GO_LINK_FLAGS)

linter_proto/linter_proto.pb.go: linter_proto/linter_proto.proto
linter_proto/linter_proto_grpc.pb.go: linter_proto/linter_proto.proto
