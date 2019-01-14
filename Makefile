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

GOCMD=go
GOBUILD=$(GOCMD) build
GOGENERATE=$(GOCMD) generate
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test -timeout 15s
GOTOOL=$(GOCMD) tool
GOGET=$(GOCMD) get
BUILD_DIR=$(PWD)/build
DEPS_DIR=$(BUILD_DIR)/deps
OUTPUT=$(BUILD_DIR)/mixdown

#
# add $(DEPS_DIR)/bin to PATH
#
export PATH := $(DEPS_DIR)/bin:$(PATH)

all: test build

prepare:
	GO111MODULES=on GOPATH=$(DEPS_DIR) $(GOGET) gopkg.in/russross/blackfriday.v2

run: build
	./build/mixdown

test: prepare
	GOPATH=$(DEPS_DIR) $(GOTEST) -coverprofile=cover.out . $(TESTPKGS)

coverage: test
	GOPATH=$(DEPS_DIR) $(GOTOOL) cover -func=cover.out

clean:
	$(GOCLEAN)
	rm -rf $(DEPS_DIR)
	rm -rf $(BUILD_DIR)

# build binary
build: prepare
	GOPATH=$(DEPS_DIR) $(GOBUILD) -o $(OUTPUT) -v

