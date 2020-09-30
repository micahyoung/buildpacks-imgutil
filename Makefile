# Go parameters
GOCMD?=go
PACKAGE_BASE=github.com/buildpacks/imgutil

all: test

install-goimports:
	@echo "> Installing goimports..."
	cd tools && $(GOCMD) install golang.org/x/tools/cmd/goimports

format: install-goimports
	@echo "> Formating code..."
	@goimports -l -w -local ${PACKAGE_BASE} .

generate: generate-windows-baselayer

install-golangci-lint:
	@echo "> Installing golangci-lint..."
	cd tools && $(GOCMD) install github.com/golangci/golangci-lint/cmd/golangci-lint

lint: install-golangci-lint
	@echo "> Linting code..."
	@golangci-lint run -c golangci.yaml

test-hivex:
	@echo "> Testing with Hivex Image"
	docker build . -f tools/hivex-docker/Dockerfile --tag hivex-docker --quiet
	docker run --rm -v gomodcache:/go/pkg/mod -v /var/run/docker.sock:/var/run/docker.sock --network=host hivex-docker \
      $(GOCMD) test --tags=hivex -parallel=1 -count=1 -v ./...

test: format lint test-hivex
	$(GOCMD) test -parallel=1 -count=1 -v ./...
