COVERAGE_THRESHOLD ?= 60
COVERAGE_PROFILE ?= coverage.out

.PHONY: test test-cover lint build bench clean

test:
	go test -v -race -count=1 ./internal/... ./cmd/...

test-cover:
	go test -race -coverprofile=$(COVERAGE_PROFILE) -covermode=atomic -count=1 ./internal/... ./cmd/...
	go tool cover -func=$(COVERAGE_PROFILE)

test-cover-check: test-cover
	@go tool cover -func=$(COVERAGE_PROFILE) | \
		awk '/^total:/ { gsub(/[%]/, "", $$NF); if ($$NF < $(COVERAGE_THRESHOLD)) { printf "FAIL: coverage %.1f%% < threshold %.0f%%\n", $$NF, $(COVERAGE_THRESHOLD); exit 1 } else { printf "PASS: coverage %.1f%% >= threshold %.0f%%\n", $$NF, $(COVERAGE_THRESHOLD) } }'

lint:
	golangci-lint run ./...

build:
	go build -v ./cmd/order-producer
	go build -v ./cmd/order-consumer

integration:
	go test -tags=integration -v -count=1 -timeout=10m ./internal/kafka/...

bench:
	go test -bench=. -benchmem ./internal/... ./cmd/...

clean:
	rm -f $(COVERAGE_PROFILE) order-producer order-consumer
