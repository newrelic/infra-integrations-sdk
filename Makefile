GO_VERSION = $(shell go version | sed 's/[^0-9.]*\([0-9.]*\).*/\1/')

GOTOOLS = github.com/axw/gocov/gocov \
          github.com/AlekSi/gocov-xml

GOLANGCILINT_VERSION = v1.56.0
GOLANGCILINT_BIN = bin/golangci-lint

# Temporary patch to avoid build failing because of the outdated documentation example
PKGS = $(shell go list ./... | egrep -v "\/docs\/|jmx")

.PHONY: all
all: lint test

.PHONY: clean
clean:
	@echo "=== $(PROJECT) === [ clean ]: Removing binaries and coverage file..."
	@rm -rf bin

.PHONY: bin
bin:
	@mkdir -p bin

.PHONY: tools/golangci-lint
tools/golangci-lint:
	@echo "installing GolangCI lint"
	@(wget -qO - https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s $(GOLANGCILINT_VERSION) )


.PHONY: tools
tools: bin tools/golangci-lint
	@echo "=== $(PROJECT) === [ tools ]: Installing tools..."
	@go get $(GOTOOLS)

.PHONY: tools-update
tools-update: tools/golangci-lint
	@echo "=== $(PROJECT) === [ tools-update ]: Updating tools..."
	@go get -u $(GOTOOLS)

.PHONY: deps
deps: tools
	@echo "=== $(PROJECT) === [ deps ]: Updating dependencies..."
	@go mod download

.PHONY: test
test: deps
	@gocov test -race $(PKGS) | gocov-xml > coverage.xml
	@gocov test github.com/newrelic/infra-integrations-sdk/v4/jmx > /dev/null # TODO: fix race for jmx package

.PHONY: tools-golangci-lint
tools-golangci-lint:
	@echo "installing GolangCI lint"
	@(wget -qO - https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s $(GOLANGCILINT_VERSION) )

.PHONY: lint
lint: deps
	@echo "=== $(PROJECT) === [ validate ]: Validating source code..."
	@${GOLANGCILINT_BIN} --version
	@${GOLANGCILINT_BIN} run

