.PHONY: test cover clean

# Run all tests 
test:
	go test -v

# Checking code coverage with tests
cover:
	go test -cover ./...

# Deleting temporary coverage files
clean:
	rm -f coverage.out