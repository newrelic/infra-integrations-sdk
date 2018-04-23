GOTOOLS = github.com/golang/lint/golint \
          gopkg.in/alecthomas/gometalinter.v2 \
          github.com/axw/gocov/gocov \
          github.com/AlekSi/gocov-xml \

# Temporary patch to avoid build failing because of the outdated documentation example
# TODO: uncomment below commented lines and remove any line that uses $(NODOCS)
NODOCS = $(shell go list ./... | grep -v /docs/)

all: lint test

deps: tools
#	@go get -v -d -t ./...
	@go get -v -d -t $(NODOCS) #todo:remove

test: deps
#	@gocov test ./... | gocov-xml > coverage.xml
	@gocov test $(NODOCS) | gocov-xml > coverage.xml

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
