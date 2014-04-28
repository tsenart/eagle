build: eagle

eagle: *.go
	go build -o $@ $^
