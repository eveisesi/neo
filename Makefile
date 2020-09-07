generate: boiler gqlgen

gqlgen:
	go run scripts/gqlgen/gqlgen.go

build:
	go build -o neo cmd/neo/*.go

serve: build
	./neo serve

top: build
	./neo top

clean:
	rm -f neo