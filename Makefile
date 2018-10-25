GOTOOLS = golang.org/x/lint/golint \
          gopkg.in/alecthomas/gometalinter.v2 \
          github.com/axw/gocov/gocov \
          github.com/AlekSi/gocov-xml \

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
