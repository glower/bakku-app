export CGO_ENABLED?=1
REPO=bakku-app/
GO?=go

check: gofmt lint

test:
	$(GO) test -tags=integration -timeout 120s -v ./...

lint:
	golangci-lint run -v

GFMT=find . -not \( \( -wholename "./vendor" \) -prune \) -name "*.go" | xargs gofmt -l
gofmt:
	@UNFMT=$$($(GFMT)); if [ -n "$$UNFMT" ]; then echo "gofmt needed on" $$UNFMT && exit 1; fi