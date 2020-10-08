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

install-golangci-lint:
	@echo "> Installing golangci-lint..."
	cd tools && $(GOCMD) install github.com/golangci/golangci-lint/cmd/golangci-lint

lint: install-golangci-lint
	@echo "> Linting code..."
	@golangci-lint run -c golangci.yaml

generate: build-bcdhive-gen
ifneq ($(OS),Windows_NT)
	$(GOCMD) generate ./...
else
	@echo "> Not compatible with Docker Windows"
endif

install-hivex-darwin:
ifneq ($(OS),Windows_NT)
	@echo "> Building bcdhive-gen in Docker"
	docker build tools/bcdhive_gen --tag bcdhive-gen-darwin
else
	@echo "> Not compatible with Docker Windows"
endif

build-bcdhive-gen:
ifneq ($(OS),Windows_NT)
	@echo "> Building bcdhive-gen in Docker"
	docker build tools/bcdhive_gen --tag bcdhive-gen
else
	@echo "> Not compatible with Docker Windows"
endif

test: generate format lint
	$(GOCMD) test -parallel=1 -count=1 -v ./...
