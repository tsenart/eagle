build: eagle squirrel

eagle: eagle.go
	go build -o $@ $<

squirrel: squirrel.go
	go build -o $@ $<
