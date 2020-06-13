generate: boiler gqlgen

boiler:
	sqlboiler --config .config/sqlboiler.toml --struct-tag-casing=camel --no-tests --no-hooks --wipe mysql

gqlgen:
	go run scripts/gqlgen/gqlgen.go

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o neo cmd/neo/*.go

serve: build
	./neo serve