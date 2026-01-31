BINARY := nexus
PKGS := ./...

.PHONY: all build test lint fmt tidy clean

all: build

build:
	go build -o $(BINARY) ./cmd/nexus

test:
	go test $(PKGS)

e2e:
	# Start mock server in background and run demo collection against it
	@echo "Starting mock server..."
	@nohup go run ./cmd/nexus mock 9999 >/tmp/nexus-mock.log 2>&1 &
	@sleep 1
	@echo "Running demo collection..."
	@go run ./cmd/nexus run examples/collections/demo.yaml || (cat /tmp/nexus-mock.log && exit 1)
	@echo "Stopping mock server..."
	@pkill -f 'nexus mock' || true

lint:
	golangci-lint run ./...

fmt:
	gofmt -s -w .

tidy:
	go mod tidy

clean:
	rm -f $(BINARY)
