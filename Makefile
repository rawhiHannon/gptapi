compile-cli:
	GOOS=linux GOARCH=386 go build -o bin/linux-gpt-cli cmd/gpt-cli/main.go
	GOOS=windows GOARCH=386 go build -o bin/windows-gpt-cli.exe cmd/gpt-cli/main.go

compile-server:
	GOOS=linux GOARCH=386 go build -o bin/linux-gpt-server cmd/gpt-server/main.go
	GOOS=windows GOARCH=386 go build -o bin/windows-gpt-server.exe cmd/gpt-server/main.go

build: compile-cli compile-server

test:
	go run tests/main.go