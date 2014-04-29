build: eagle squirrel

eagle: *.go
	go build -o $@ $^

squirrel:
	$(MAKE) -C $@

.PHONY: squirrel
