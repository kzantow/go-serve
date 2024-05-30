.PHONY: test
test: unit

.PHONY: bootstrap
bootstrap:
	go install github.com/rinchsan/gosimports/cmd/gosimports@v0.3.8

.PHONY: format
format:
	gofmt -w .
	go mod tidy
	gosimports -w -local github.com/spdx .

.PHONY: unit
unit:
	go test -v -covermode=count -coverprofile=profile.cov ./...
