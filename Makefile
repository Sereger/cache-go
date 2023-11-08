GO    = go
GOBIN ?= $(PWD)/bin

bench:
	go test -bench=. -v -benchmem ./...

test:
	go test ./... -v

.PHONY: install-tools-gimps
install-tools-gimps:
ifeq ($(wildcard $(GOBIN)/gimps),)
	@echo "Downloading gimps"
	@GOBIN=$(GOBIN) $(GO) install -mod=readonly go.xrstf.de/gimps@latest
endif

.PHONY: imports-fix
imports-fix: install-tools-gimps ; ## Fix imports
	$(info $(M) fixing imports...)
	@$(GOBIN)/gimps --config .gimps.yaml .

.PHONY: install-lint
install-lint:
	@GOBIN=$(GOBIN) $(GO) install -mod=readonly github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.2

.PHONY: lint
lint: install-lint ## Run linters
	$(info $(M) running linters...)
	@$(GOBIN)/golangci-lint run --timeout 5m0s ./...