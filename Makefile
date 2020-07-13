bin/vimremote: src/main.go
	go build -ldflags="-w" -o bin/vimremote src/main.go

