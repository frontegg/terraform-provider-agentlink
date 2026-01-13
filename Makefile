default: build

build:
	go build -o terraform-provider-agentlink

install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/frontegg/agentlink/0.0.1/$$(go env GOOS)_$$(go env GOARCH)
	cp terraform-provider-agentlink ~/.terraform.d/plugins/registry.terraform.io/frontegg/agentlink/0.0.1/$$(go env GOOS)_$$(go env GOARCH)/

test:
	go test -v ./...

testacc:
	TF_ACC=1 go test -v ./... -timeout 120m

fmt:
	go fmt ./...

vet:
	go vet ./...

tidy:
	go mod tidy

test-plan: build
	cd examples/provider && TF_LOG=INFO TF_CLI_CONFIG_FILE=./dev.tfrc terraform plan

test-apply: build
	cd examples/provider && TF_LOG=INFO TF_CLI_CONFIG_FILE=./dev.tfrc terraform apply -auto-approve

.PHONY: build install test testacc fmt vet tidy test-plan test-apply
