all: test

deps:
	@echo "====> Install depedencies..."
	go get -v -d -t ./...

devdeps: deps
	@echo "====> Install depedencies for development..."
	go get -v gopkg.in/alecthomas/gometalinter.v2
	gometalinter.v2 --install > /dev/null
	go get -v github.com/axw/gocov/gocov
	go get -v github.com/AlekSi/gocov-xml

test: devdeps
	@echo "====> Running linters..."
	@gometalinter.v2 --config=.gometalinter.json ./...
	@echo "====> Running unit tests..."
	gocov test ./... | gocov-xml > coverage.xml

clean:
	rm -rf coverage.xml

.PHONY: all deps devdeps test clean
