git fetch --all

latest=$(git tag -l 'v[0-9]*.[0-9]*.[0-9]*' | tr - \~ | sort -V | tr \~ - | tail -1 | tr -d \v)

feimage="docker.pkg.github.com/eveisesi/neo/frontend:${latest}"
beimage="docker.pkg.github.com/eveisesi/neo/backend:${latest}"

export FRONTEND_IMAGE="docker.pkg.github.com/eveisesi/neo/frontend:${latest}"
export BACKEND_IMAGE="docker.pkg.github.com/eveisesi/neo/backend:${latest}"
export PROCESS_LIMIT=10
export PROCESS_SLEEP=250

docker pull ${feimage}
docker pull ${beimage}