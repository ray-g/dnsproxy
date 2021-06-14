.PHONY: vendor
vendor:
	go mod tidy

.PHONY: build
build: bindata
	go build -v -o build/dnsproxy .

.PHONY: bindata
bindata:
	go-bindata --nocompress -pkg api -o api/bindata.go -prefix web ./web/...
