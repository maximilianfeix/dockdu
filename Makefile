BIN     := dockdu
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build test lint snapshot check

build:
	CGO_ENABLED=0 go build -trimpath -ldflags '$(LDFLAGS)' -o $(BIN) ./cmd/dockdu

test:
	go test ./...

# lint fails when gofmt would change any file (gofmt -l alone exits 0).
lint:
	@unformatted="$$(gofmt -l .)"; if [ -n "$$unformatted" ]; then echo "gofmt needed:"; echo "$$unformatted"; exit 1; fi
	go vet ./...

check: lint test build

snapshot:
	goreleaser release --snapshot --clean

# integration runs the live-daemon tests where the data root is visible:
# inside a container with the socket and /var/lib/docker mounted read-only.
# On Linux hosts with root access you can also run it directly:
#   DOCKDU_INTEGRATION=1 go test ./...
integration:
	docker run --rm \
	  -v /var/run/docker.sock:/var/run/docker.sock \
	  -v /var/lib/docker:/var/lib/docker:ro \
	  -v "$(CURDIR)":/src -w /src \
	  -e DOCKDU_INTEGRATION=1 -e GOFLAGS=-buildvcs=false \
	  golang:1.27 go test ./... -run Integration -v
