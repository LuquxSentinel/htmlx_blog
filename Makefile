build:
	@go build -o ./bin/templater

run: build
	@./bin/templater

test:
	@go test ./...