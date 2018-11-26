GO_VERSION = $(shell go version | sed 's/[^0-9.]*\([0-9.]*\).*/\1/')

GOTOOLS = gopkg.in/alecthomas/gometalinter.v2 \
          github.com/axw/gocov/gocov \
          github.com/AlekSi/gocov-xml \

# golint only supports the last two Go versions, update the value when its not supported anymore.
GOLINT_MIN_GO_VERSION = 1.9

# If GO_VERSION is equal or higher than the GOLINT_MIN_GO_VERSION we use golint.
ifeq ($(GO_VERSION),$(shell echo "$(GOLINT_MIN_GO_VERSION)\n$(GO_VERSION)" | sort -V | tail -n1))
	GOTOOLS += golang.org/x/lint/golint
endif

# Temporary patch to avoid build failing because of the outdated documentation example
PKGS = $(shell go list ./... | egrep -v "\/docs\/|jmx")

all: lint test

deps: tools
	@go get -v -d -t $(PKGS)

test: deps
	@gocov test -race $(PKGS) | gocov-xml > coverage.xml
	@gocov test github.com/newrelic/infra-integrations-sdk/jmx > /dev/null # TODO: fix race for jmx package

clean:
	rm -rf coverage.xml

tools:
	@go get $(GOTOOLS)
	@gometalinter.v2 --install > /dev/null

tools-update:
	@go get -u $(GOTOOLS)
	@gometalinter.v2 --install

lint: deps
	@gometalinter.v2 --config=.gometalinter.json ./...

lint-all: deps
	@gometalinter.v2 --config=.gometalinter.json --enable=interfacer --enable=gosimple ./...

.PHONY: all deps devdeps test clean
