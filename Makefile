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
	scp ./bin/boardctl rpi1b:~/boardctl
	ssh rpi1b "sudo sh -c \"mkdir -p /opt/mosho-boardctl && mv boardctl /opt/mosho-boardctl\" && rm boardctl"

remote-env: remote-copy
	ssh rpi1b "/opt/mosho-boardctl/boardctl env"

remote-cmd: remote-copy
	ssh rpi1b "/opt/mosho-boardctl/boardctl cmd"

