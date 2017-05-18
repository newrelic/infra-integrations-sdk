all: test

deps:
	@echo "====> Install depedencies..."
	go get -v -d -t ./...

devdeps: deps
	@echo "====> Install depedencies for development..."
	go get -v github.com/golang/lint/golint
	go get -v github.com/axw/gocov/gocov
	go get -v github.com/AlekSi/gocov-xml

test: devdeps
	@echo "====> Running golint..."
	golint ./...
	@echo "====> Running gofmt..."
	gofmt -l .
	@echo "====> Running unit tests..."
	gocov test ./... | gocov-xml > coverage.xml

clean:
	rm -rf coverage.xml

.PHONY: all deps devdeps test clean
