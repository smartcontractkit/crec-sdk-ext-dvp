.PHONY: build

build:
	GOOS=wasip1 GOARCH=wasm CGO_ENABLED=0 go build -o bundle/binary.wasm -trimpath -ldflags="-buildid= -w -s" .
