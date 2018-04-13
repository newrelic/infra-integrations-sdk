GOTOOLS = github.com/golang/lint/golint \
          gopkg.in/alecthomas/gometalinter.v2 \
          github.com/axw/gocov/gocov \
          github.com/AlekSi/gocov-xml \

all: test

deps: tools
	@echo "====> Install depedencies..."
	go get -v -d -t ./...

test: deps
	@echo "====> Running unit tests..."
	gocov test ./... | gocov-xml > coverage.xml

clean:
	rm -rf coverage.xml

tools:
	@echo "====>: Installing tools required by the project..."
	@go get $(GOTOOLS)
	@gometalinter.v2 --install

tools-update:
	@echo "====> Updating tools required by the project..."
	@go get -u $(GOTOOLS)
	@gometalinter.v2 --install

validate: lint

validate-all: lint-all

lint: deps
	@echo "====> Validating source code running gometalinter..."
	@gometalinter.v2 --config=.gometalinter.json ./...

lint-all: deps
	@echo "====> Validating source code running gometalinter..."
	@gometalinter.v2 --config=.gometalinter.json --enable=interfacer --enable=gosimple ./...

.PHONY: all deps devdeps test clean
