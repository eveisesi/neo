LATEST_TAG := $(shell git tag -l 'v[0-9]*.[0-9]*.[0-9]*' | tr - \~ | sort -V | tr \~ - | tail -1 | tr -d \v)

build-be:
	docker build backend/. -tag backend:latest

build-fe:
	docker build frontend/. -tag frontend:latest

pull:
	docker pull docker.pkg.github.com/eveisesi/neo/frontend:$(LATEST_TAG)
	docker pull docker.pkg.github.com/eveisesi/neo/backend:$(LATEST_TAG)

latest:
	@echo $(LATEST_TAG)


	