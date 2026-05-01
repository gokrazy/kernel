AUTOUPDATE ?= ~/go/src/github.com/gokrazy/autoupdate

.PHONY: rebuild

all:

rebuild:
	GOBIN=$$PWD/_build sh -c '(cd $(AUTOUPDATE) && CGO_ENABLED=0 go install ./cmd/gokr-rebuild-kernel)'
	(cd _build && CGO_ENABLED=0 ./gokr-rebuild-kernel -cross=arm64)
