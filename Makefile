.PHONY: test cover clean

test:
	go test -v

cover:
	go test -cover ./...

clean:
	rm -f coverage.out