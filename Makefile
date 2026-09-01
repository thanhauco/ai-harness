.PHONY: all build test race bench lint clean run-examples

all: test build

build:
	mkdir -p bin
	go build -o bin/aih ./cmd/aih

test:
	go test -v ./...

race:
	go test -v -race -covermode=atomic ./...

bench:
	go test -bench=. -benchmem ./...

lint:
	gofmt -s -l .

run-examples:
	go run ./examples/01_resilient_call/main.go
	go run ./examples/02_agent_dag/main.go
	go run ./examples/03_eval_benchmark/main.go

clean:
	rm -rf bin/ coverage.txt
