.PHONY: build test clean install

build:
	go build -o streamctl ./cmd/streamctl

test:
	go test -v ./...

clean:
	rm -f streamctl

install: build
	cp streamctl ~/bin/

# Run a single test
test-one:
	go test -v -run $(TEST) ./...
