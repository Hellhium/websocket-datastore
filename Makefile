.PHONY: build clean init

revision = $(shell git diff --quiet && git rev-parse HEAD || echo "dirty")
curdate = $(shell date +"%D-%T")
tlibpath = "github.com/hellhium/websocket-datastore/lib/tlib"
gobuild = go build -ldflags="-X '$(tlibpath).BuildTime=$(curdate)' -X '$(tlibpath).BuildRef=$(revision)'" -o build/bin/$(1) ./cmd/$(1)/$(1).go

all: clean init build

clean:
	rm -rf build

init:
	mkdir -p build/bin

build:
	$(call gobuild,wsapi)
