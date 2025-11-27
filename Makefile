.PHONY: rebuild

all:

rebuild:
	GOBIN=$$PWD/_build sh -c '(cd ~/go/src/github.com/gokrazy/autoupdate && CGO_ENABLED=0 go install ./cmd/gokr-rebuild-kernel)'
	(cd _build && CGO_ENABLED=0 ./gokr-rebuild-kernel -cross=arm64)
