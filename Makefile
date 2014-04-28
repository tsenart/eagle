build: eagle squirrel

eagle: eagle.go config.go loadtest.go
	go build -o $@ $^

squirrel: squirrel.go
	go build -o $@ $^
