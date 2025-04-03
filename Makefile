.PHONY: test test-v test-coverage clean

# Default target
all: test

# Run tests across all packages
test:
	go test -v ./...

# Run tests with verbose output
test-v:
	go test -v ./...

# Run tests with coverage report
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean up generated files
clean:
	rm -f coverage.out coverage.html 