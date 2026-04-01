.PHONY: build run clean
build:
	CGO_ENABLED=0 go build -o rustler ./cmd/rustler/
run: build
	./rustler
clean:
	rm -f rustler
