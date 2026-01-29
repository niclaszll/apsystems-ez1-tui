.PHONY: build clean install test run help

help:
	@echo "Available targets:"
	@echo "  build     - Build the TUI application"
	@echo "  install   - Install the TUI application to GOPATH/bin"
	@echo "  clean     - Remove built binaries"
	@echo "  run       - Run the TUI (requires HOST variable)"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make run HOST=192.168.1.100"
	@echo "  make install"

build:
	go build -o ez1-tui ./cmd/ez1-tui

install:
	go install ./cmd/ez1-tui

clean:
	rm -f ez1-tui

run:
ifndef HOST
	@echo "Error: HOST variable is required"
	@echo "Usage: make run HOST=192.168.1.100"
	@exit 1
endif
	go run ./cmd/ez1-tui -host $(HOST)
