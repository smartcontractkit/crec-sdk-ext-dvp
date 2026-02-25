.PHONY: build

build:
	GOOS=wasip1 GOARCH=wasm CGO_ENABLED=0 go build -o bundle/binary.wasm -trimpath -ldflags="-buildid= -w -s" .
	brotli -c bundle/binary.wasm | base64 > bundle/binary.wasm.br.b64
