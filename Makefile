.PHONY: vendor
vendor:
	go mod tidy

.PHONY: build
build:
	go build -v .
