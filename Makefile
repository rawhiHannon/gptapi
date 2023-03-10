compile-chatapp:
	GOOS=linux GOARCH=386 go build -o bin/linux-chatapp cmd/chatapp/main.go
	GOOS=windows GOARCH=386 go build -o bin/windows-chatapp.exe cmd/chatapp/main.go

compile-botapp:
	GOOS=linux GOARCH=386 go build -o bin/linux-botapp cmd/botapp/main.go
	GOOS=windows GOARCH=386 go build -o bin/windows-botapp.exe cmd/botapp/main.go


build: compile-cli compile-chatapp compile-botapp

test:
	go run tests/main.go