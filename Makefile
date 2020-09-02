# vi: ft=make

.PHONY: test
test:
	go test ./... -json | gotestpretty
