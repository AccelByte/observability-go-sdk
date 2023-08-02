all: lint test

lint:
	@echo "linting the code"
	golangci-lint run

test:
	@echo "running the unit tests"
	mkdir -p ".cover"
	rm -rf ".cover/*"
	go test -race -cover -covermode="atomic" -coverprofile=".cover/cover.out" -coverpkg=./... ./...
	go tool cover -func=".cover/cover.out"
