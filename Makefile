.PHONY: test
test:
	go test ./...

.PHONY: vendor
vendor:
	go mod tidy
	go mod download
	go mod vendor

.PHONY: watcher-build
watcher-build:
	cd watcher && make build