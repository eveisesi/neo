latest=$(git describe --tags `git rev-list --tags --max-count=1`)
export DOCKER_IMAGE="docker.pkg.github.com/eveisesi/neo/neo:${latest:1}"
source docker.env