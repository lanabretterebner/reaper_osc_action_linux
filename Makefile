.DEFAULT_GOAL := build

.PHONY:fmt vet build test plugin
fmt:
	go fmt

vet: fmt
	go vet

build: vet
	GOOS=linux GOARCH=amd64 go build -o org.smyck.reaper-osc-action.sdPlugin/reaper_osc_action_linux

test:
	go test ./...

plugin:
	# Requires fd and sd executables
	# https://github.com/sharkdp/fd
	# https://github.com/elgatosf/cli
	fd -H -I '.DS_Store' -x rm -f
	sd pack -f org.smyck.reaper-osc-action.sdPlugin
