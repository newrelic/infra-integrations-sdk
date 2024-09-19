GO_VERSION = $(shell go version | sed 's/[^0-9.]*\([0-9.]*\).*/\1/')

# Temporary patch to avoid build failing because of the outdated documentation example
PKGS = $(shell go list ./... | egrep -v "\/docs\/|jmx")

.PHONY: all
all: lint test

.PHONY: test
test:
	@go test -race $(PKGS)
	@go test github.com/newrelic/infra-integrations-sdk/jmx > /dev/null # TODO: fix race for jmx package

.PHONY: clean
clean:
	rm -rf coverage.xml
