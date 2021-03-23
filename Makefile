TEST?=$$(go list ./... |grep -v 'vendor')
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
GO_CMD ?= go
PKG ?= go-webexec

export GO111MODULE=on

default: build

setup: ## Install required libraries/tools for build tasks
	@command -v cover 2>&1 >/dev/null       || GO111MODULE=off go get -u -v golang.org/x/tools/cmd/cover
	@command -v gofumpt 2>&1 >/dev/null     || GO111MODULE=off go get -u -v mvdan.cc/gofumpt
	@command -v gosec 2>&1 >/dev/null       || GO111MODULE=off go get -u -v github.com/securego/gosec/cmd/gosec
	@command -v goveralls 2>&1 >/dev/null   || GO111MODULE=off go get -u -v github.com/mattn/goveralls
	@command -v ineffassign 2>&1 >/dev/null || GO111MODULE=off go get -u -v github.com/gordonklaus/ineffassign
	@command -v misspell 2>&1 >/dev/null    || GO111MODULE=off go get -u -v github.com/client9/misspell/cmd/misspell
	@command -v revive 2>&1 >/dev/null      || GO111MODULE=off go get -u -v github.com/mgechev/revive

fmt:
	$(GO_CMD)fmt -w $(GOFMT_FILES)

upx:
	@echo "Target upx"
	curl -sL https://github.com/upx/upx/releases/download/v3.96/upx-3.96-amd64_linux.tar.xz | tar -xJ -C .
	mv ./upx-3.96-amd64_linux upx

build: fmt
	echo "Standard binary"
	GOOS=linux $(GO_CMD) build
	du -hs go-webexec

build-small: upx fmt
	echo "Standard binary"
	# https://golang.org/cmd/link/
	GOOS=linux $(GO_CMD) build -ldflags="-s -w"
	du -hs go-webexec

	echo "UPX binary"
	upx/upx --brute go-webexec
	du -hs go-webexec

test: fmt
	$(GO_CMD) test -i $(TEST) || exit 1
	echo $(TEST) | \
		xargs -t -n4 $(GO_CMD) test $(TESTARGS) -timeout=30s -parallel=4

multi:
	gox -ldflags="-s -w" -os="linux" -os="windows" -os="darwin"
