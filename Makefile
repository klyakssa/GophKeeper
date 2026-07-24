BUILD_VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "N/A")
BUILD_DATE=$(shell date -u +%Y-%m-%d_%H:%M:%S || echo "N/A")
BUILD_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "N/A")

.PHONY: all

all: s

s:
	go run cmd/server/main.go

c:
	go run cmd/client/client.go