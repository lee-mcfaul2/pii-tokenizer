.PHONY: build test test-unit test-integration test-e2e lint run-local clean

BIN := bin/tokenizer
ROTATE_BIN := bin/rotate-kmaster

build:
	@mkdir -p bin
	go build -o $(BIN) ./cmd/tokenizer
	go build -o $(ROTATE_BIN) ./tools/rotate-kmaster

test: test-unit test-integration

test-unit:
	go test ./internal/...

test-integration:
	go test -tags=integration ./tests/integration/...

test-e2e:
	go test -tags=e2e ./tests/e2e/...

lint:
	golangci-lint run ./...

run-local:
	docker compose -f deploy/dev/compose.yaml up --build

clean:
	rm -rf bin/ out/ coverage.out
