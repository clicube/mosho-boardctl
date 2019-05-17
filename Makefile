fmt:
	go fmt ./...

build:
	mkdir -p ./bin
	GOOS=linux GOARCH=arm GOARM=6 go build -o ./bin/boardctl ./cmd

clean:
	go clean ./...
	rm -rf ./bin

test:
	go test ./...

remote-copy: build
	ssh raspi "mkdir -p ~/services/mosho-boardctl/bin"
	scp ./bin/boardctl raspi:services/mosho-boardctl/bin

remote-env: remote-copy
	ssh raspi "cd services/mosho-boardctl/bin && ./boardctl env"

remote-cmd: remote-copy
	ssh raspi "cd services/mosho-boardctl/bin && ./boardctl cmd"

