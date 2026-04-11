.PHONY: build test test-e2e lint tidy check clean bench \
       docker-build docker-test docker-dev docker-bench \
       docker-test-rtk docker-dev-rtk docker-bench-rtk

CGO_ENABLED ?= 0

build:
	CGO_ENABLED=$(CGO_ENABLED) go build -o bin/tip ./cmd/tip

test:
	CGO_ENABLED=$(CGO_ENABLED) go test ./cmd/... ./mode/... \
		./session/... ./context/... ./policy/... ./observe/... \
		./adapter/... ./install/... ./shell/... ./toggle/... \
		./render/... ./status/... ./config/... ./version/...

test-e2e:
	CGO_ENABLED=$(CGO_ENABLED) go test -tags e2e ./e2e/...

lint:
	go vet ./...

tidy:
	go mod tidy

check: tidy lint build test test-e2e

clean:
	rm -rf bin/

# Docker — standard (tip only)
docker-build:
	docker compose build test

docker-test:
	docker compose run --rm test

docker-bench:
	docker compose run --rm bench

docker-dev:
	docker compose run --rm --service-ports dev

# Docker — rtk-enabled (tip + rtk)
docker-test-rtk:
	docker compose run --rm test-rtk

docker-bench-rtk:
	docker compose run --rm bench-rtk

docker-dev-rtk:
	docker compose run --rm --service-ports dev-rtk
