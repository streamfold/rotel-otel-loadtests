.PHONY: build

build:
	mkdir -p dist && go build -o dist/benchmarks ./cmd/benchmarks
