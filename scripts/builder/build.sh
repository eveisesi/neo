#!/bin/sh

source scripts/builder/.env

# Login to Docker
echo $GITHUB_TOKEN | docker login docker.pkg.github.com -u $DOCKER_USERNAME --password-stdin

# Build Image and tag with app
docker build . --tag $IMAGE_ID:latest --tag $IMAGE_ID:$VERSION

docker push $IMAGE_ID:latest
docker push $IMAGE_ID:$VERSION
docker logout docker.pkg.github.com