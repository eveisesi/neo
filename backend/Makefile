gqlgen:
	go run scripts/gqlgen/gqlgen.go

build:
	go build -o neo cmd/neo/*.go

serve: build
	./neo serve