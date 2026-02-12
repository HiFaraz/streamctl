.PHONY: build test clean install restart hooks

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

# Rebuild and restart MCP server
restart: build
	@echo "Stopping running streamctl processes..."
	-pkill -f "streamctl serve" 2>/dev/null || true
	@echo "Done. Claude Code will restart the MCP server on next tool call."

# Install git hooks
hooks:
	cp hooks/* .git/hooks/
	chmod +x .git/hooks/*
	@echo "Git hooks installed."
