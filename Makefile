PACKAGES=$(shell go list ./... | grep -v /vendor/)
RACE := $(shell test $$(go env GOARCH) != "amd64" || (echo "-race"))
LDFLAGS=

all: build

build:
	@echo "Compiling..."
	@mkdir -p ./bin
	@gox $(LDFLAGS) -output "bin/{{.Dir}}_{{.OS}}_{{.Arch}}" -os="linux" -os="darwin" -arch="386" -arch="amd64" ./