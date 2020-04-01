#
# Copyright (C) 2019 Masatoshi Fukunaga
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to
# deal in the Software without restriction, including without limitation the
# rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
# sell copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.  IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
# FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
# IN THE SOFTWARE.
#
# Created by Masatoshi Fukunaga on 19/01/09
#

NAME=mixdown
GOCMD=go
GOBUILD=$(GOCMD) build
GOGENERATE=$(GOCMD) generate
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test -timeout 15s
GOTOOL=$(GOCMD) tool
GOGET=$(GOCMD) get
BUILD_DIR=$(PWD)/build
DEPS_DIR=$(BUILD_DIR)/deps
GOLINT=`which golangci-lint`
LINT_OPT=--max-issues-per-linter=0 \
		--max-same-issues=0 \
		--issues-exit-code=0 \
		--tests=false \
		--enable=dupl \
		--enable=goconst \
		--enable=gocritic \
		--enable=gofmt \
		--enable=goimports \
		--enable=golint \
		--enable=gosec \
		--enable=maligned \
		--enable=misspell \
		--enable=stylecheck \
		--enable=unconvert \
		--exclude=ifElseChain

#
# add $(DEPS_DIR)/bin to PATH
#
export PATH := $(DEPS_DIR)/bin:$(PATH)

all: test build

prepare:
	GO111MODULES=on GOPATH=$(DEPS_DIR) $(GOGET) gopkg.in/russross/blackfriday.v2

test: prepare
	GOPATH=$(DEPS_DIR) $(GOTEST) -coverprofile=coverage.out -covermode=atomic ./...

lint:
	GOPATH=$(DEPS_DIR) $(GOLINT) run $(LINT_OPT)

clean:
	$(GOCLEAN)
	rm -rf $(DEPS_DIR)
	rm -rf $(BUILD_DIR)

# build binary
build: prepare
	GOPATH=$(DEPS_DIR) $(GOBUILD) -o $(BUILD_DIR)/$(NAME) -v

build-linux: prepare
	GOPATH=$(DEPS_DIR) GOARCH=amd64 GOOS=linux $(GOBUILD) -o $(BUILD_DIR)/linux/$(NAME) -v

build-darwin: prepare
	GOPATH=$(DEPS_DIR) GOARCH=amd64 GOOS=darwin $(GOBUILD) -o $(BUILD_DIR)/darwin/$(NAME) -v


