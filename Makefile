VERSION := $(shell grep "version =" cli/cli.go | cut -d '"' -f 2)

#define release
#| File | Sign  | SHA512SUM |
#|------|-------|-----------|
#| [switchctl_$(VERSION)_darwin_amd64.zip](switchctl_$(VERSION)_darwin_amd64.zip) |  [switchctl_$(VERSION)_darwin_amd64.zip.asc](switchctl_$(VERSION)_darwin_amd64.zip.asc) | $(shell sha512sum dist/switchctl_$(VERSION)_darwin_amd64.zip) |
#| [switchctl_$(VERSION)_linux_amd64.zip](switchctl_$(VERSION)_linux_amd64.zip) |  [switchctl_$(VERSION)_linux_amd64.zip.asc](switchctl_$(VERSION)_linux_amd64.zip.asc) | $(shell sha512sum dist/switchctl_$(VERSION)_linux_amd64.zip) |
#endef

fmt:
	go fmt ./ ./cli ./common ./conf ./dns ./progress ./ssh

build: fmt
	GOOS=darwin GOARCH=amd64 go build -o build/darwin_amd64/switchctl
	GOOS=linux GOARCH=amd64 go build -o build/linux_amd64/switchctl

dist: clean build
	mkdir dist
	zip -j dist/switchctl_$(VERSION)_darwin_amd64.zip build/darwin_amd64/switchctl
	zip -j dist/switchctl_$(VERSION)_linux_amd64.zip build/linux_amd64/switchctl

clean:
	rm -rf dist

sign:
	gpg --armor --sign --detach-sig dist/switchctl_$(VERSION)_darwin_amd64.zip
	gpg --armor --sign --detach-sig dist/switchctl_$(VERSION)_linux_amd64.zip

release:
	@echo "| File | Sign  | SHA512SUM |"
	@echo "|------|-------|-----------|"
	@echo "| [switchctl_$(VERSION)_darwin_amd64.zip](switchctl_$(VERSION)_darwin_amd64.zip) | [switchctl_$(VERSION)_darwin_amd64.zip.asc](switchctl_$(VERSION)_darwin_amd64.zip.asc) | $(shell sha512sum dist/switchctl_$(VERSION)_darwin_amd64.zip | cut -d " " -f 1) |"
	@echo "| [switchctl_$(VERSION)_linux_amd64.zip](switchctl_$(VERSION)_linux_amd64.zip) | [switchctl_$(VERSION)_linux_amd64.zip.asc](switchctl_$(VERSION)_linux_amd64.zip.asc) | $(shell sha512sum dist/switchctl_$(VERSION)_linux_amd64.zip | cut -d " " -f 1) |"
