lint:
ifeq ($(GO_VERSION), go1.6)
	echo "$(GO_VERSION) is not a supported Go release. Skipping golint."
else ifeq ($(GO_VERSION), go1.7)
	echo "$(GO_VERSION) is not a supported Go release. Skipping golint."
else ifeq ($(GO_VERSION), go1.8)
	echo "$(GO_VERSION) is not a supported Go release. Skipping golint."
else
	golint
endif

test:
	go test -cover -v ./...

gosec:
	go get -u github.com/securego/gosec/cmd/gosec...
	gosec ./...

ci-lint:
	golangci-lint run

ineffassign:
	go get -u github.com/gordonklaus/ineffassign/...
	ineffassign .
