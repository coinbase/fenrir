.PHONY: clean build

build:
	GOOS=linux GOARCH=amd64 go build -o hello.lambda .

clean:
	rm hello

start: build
	sam local start-api -t release.yml

release: build
	zip template.yml.hello.zip hello.lambda
