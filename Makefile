generate: boiler gqlgen

boiler:
	sqlboiler --config .config/sqlboiler.toml --struct-tag-casing=camel --no-tests --no-hooks --wipe mysql

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