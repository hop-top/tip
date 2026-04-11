.PHONY: build test test-e2e lint tidy check clean

CGO_ENABLED ?= 0

build:
	CGO_ENABLED=$(CGO_ENABLED) go build -o bin/tip ./cmd/tip

test:
	CGO_ENABLED=$(CGO_ENABLED) go test ./cmd/... ./mode/... ./session/... ./context/... ./policy/... ./observe/... ./adapter/... ./install/... ./shell/... ./toggle/... ./render/... ./status/... ./config/... ./version/...

test-e2e:
	CGO_ENABLED=$(CGO_ENABLED) go test -tags e2e ./e2e/...

lint:
	go vet ./...

tidy:
	go mod tidy

check: tidy lint build test test-e2e

clean:
	rm -rf bin/
