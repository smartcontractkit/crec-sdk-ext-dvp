# Post-WASM brotli+base64 steps match crec-courier-service/workflows/operation-writer/Makefile
# (fixed quality -q 6, openssl base64 -A) so local/CI/prod hashes of binary.wasm.br.b64 stay aligned.
.PHONY: build

build:
	GOOS=wasip1 GOARCH=wasm CGO_ENABLED=0 go build -buildvcs=false -trimpath -ldflags="-buildid= -w -s" -o bundle/binary.wasm .
	@test -f bundle/binary.wasm || (echo "Error: bundle/binary.wasm was not created" && exit 1)
	rm -f bundle/binary.wasm.br
	brotli -q 6 -o bundle/binary.wasm.br bundle/binary.wasm
	@test -f bundle/binary.wasm.br || (echo "Error: bundle/binary.wasm.br was not created" && exit 1)
	openssl base64 -A -in bundle/binary.wasm.br -out bundle/binary.wasm.br.b64
	@test -f bundle/binary.wasm.br.b64 || (echo "Error: bundle/binary.wasm.br.b64 was not created" && exit 1)
	rm -f bundle/binary.wasm.br
