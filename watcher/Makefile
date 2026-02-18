.PHONY: build

build:
	mkdir -p wasm
	GOOS=wasip1 GOARCH=wasm CGO_ENABLED=0 go build -o wasm/workflow.wasm -trimpath -ldflags="-buildid= -w -s" .
	cp wasm/workflow.wasm bundle/binary.wasm
