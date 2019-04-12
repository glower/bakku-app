export CGO_ENABLED?=1
REPO=bakku-app/
GO?=go

LINT_FLAGS := run -v --deadline=120s
LINTER_EXE := golangci-lint
LINTER:= ./bin/$(LINTER_EXE)

check: gofmt lint

test:
	$(GO) test -tags=integration -timeout 120s -v ./...

$(LINTER):
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s v1.15.0

lint: $(LINTER)
	$(LINTER) $(LINT_FLAGS)

GFMT=find . -not \( \( -wholename "./vendor" \) -prune \) -name "*.go" | xargs gofmt -l
gofmt:
	@UNFMT=$$($(GFMT)); if [ -n "$$UNFMT" ]; then echo "gofmt needed on" $$UNFMT && exit 1; fi

vendor:
	dep ensure -v

run:
	$(GO) run cmd/bakku-app/main.go
