.PHONY: test
test:
	go test ./...

.PHONY: vendor
vendor:
	go mod tidy
	go mod download
	go mod vendor