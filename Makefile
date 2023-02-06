compile-cli:
	GOOS=linux GOARCH=386 go build -o bin/linux-gpt-cli cmd/gpt-cli/main.go
	GOOS=windows GOARCH=386 go build -o bin/windows-gpt-cli.exe cmd/gpt-cli/main.go

build: compile-cli

test:
	go run tests/main.go