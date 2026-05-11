BINARY := album2vid-gui
CMD := ./cmd/album2vid-gui

.PHONY: build run

build:
	go build -ldflags="-s -w" -o $(BINARY) $(CMD)

run: build
	./$(BINARY) $(ARGS)
