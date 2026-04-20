.PHONY: build

build:
	GOOS=wasip1 GOARCH=wasm CGO_ENABLED=0 go build -buildvcs=false -trimpath -ldflags="-buildid= -w -s" -o bundle/binary.wasm .
	brotli -c bundle/binary.wasm | base64 > bundle/binary.wasm.br.b64
