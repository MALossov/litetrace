.PHONY: build clean test run

build:
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o litetrace .

clean:
	rm -f litetrace
	go clean

test:
	go test ./internal/... -v

run:
	go run main.go --help
