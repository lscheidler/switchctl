fmt:
	go fmt ./ ./cli ./common ./conf ./dns ./progress ./ssh

build: fmt
	GOOS=darwin GOARCH=amd64 go build -o build/darwin_amd64/switchctl
	GOOS=linux GOARCH=amd64 go build -o build/linux_amd64/switchctl

dist: clean build
	mkdir dist
	zip -j dist/switchctl_darwin_amd64.zip build/darwin_amd64/switchctl
	zip -j dist/switchctl_linux_amd64.zip build/linux_amd64/switchctl

clean:
	rm -rf dist
