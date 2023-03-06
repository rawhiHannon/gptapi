compile-cli:
	GOOS=linux GOARCH=386 go build -o bin/linux-gpt-cli cmd/gpt-cli/main.go
	GOOS=windows GOARCH=386 go build -o bin/windows-gpt-cli.exe cmd/gpt-cli/main.go

compile-server:
	GOOS=linux GOARCH=386 go build -o bin/linux-gpt-server cmd/gpt-server/main.go
	GOOS=windows GOARCH=386 go build -o bin/windows-gpt-server.exe cmd/gpt-server/main.go

compile-bot:
	GOOS=linux GOARCH=386 go build -o bin/linux-gpt-bot cmd/gpt-bot/main.go
	GOOS=windows GOARCH=386 go build -o bin/windows-gpt-bot.exe cmd/gpt-bot/main.go

build: compile-cli compile-server compile-bot

test:
	go run tests/main.go