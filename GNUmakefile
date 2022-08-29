REGISTRY=registry.terraform.io
NAMESPACE=jhriggs
NAME=rds-configuration
BINARY=terraform-provider-${NAME}
VERSION=$(or $(shell git describe --abbrev=0 --tag), $(error missing VERSION))
OS_ARCH=$(shell go env GOOS)_$(shell go env GOARCH)

default: build

.PHONY: build fmt tidy generate install

build: fmt tidy
	go build

fmt:
	gofmt -l -d -w $$(find . -name '*.go' | grep -v vendor)

tidy:
	go mod tidy

generate:
	go generate

# address nested schema bug in data sources
	sed \
		-e 's/^\(- `description` (String)\)$$/\1 A description of the setting./' \
		-e 's/^\(- `name` (String)\)$$/\1 The name of the setting./' \
		-e 's/^\(- `value` (Number)\)$$/\1 The value of the setting./' \
		docs/data-sources/configuration.md \
		> docs/data-sources/configuration.md.tmp \
	&& mv docs/data-sources/configuration.md.tmp docs/data-sources/configuration.md

install: build generate
	mkdir -p ~/.terraform.d/plugins/${REGISTRY}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${REGISTRY}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
