GO				?= go
GOPATH			?= $$($(GO) env GOPATH)

codegen:
	GO111MODULE=on $(GO) get github.com/golang/mock/mockgen@v1.4.3
	PATH=$$PATH:$(GOPATH)/bin $(GO) generate ./...

check:
	$(GO) test ./...

golangci:
	@curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin v1.21.0
	$(GOPATH)/bin/golangci-lint run -c .golangci.yml
