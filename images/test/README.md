## Docker image
The Docker image is available at <https://hub.docker.com/r/entigolabs/entigo-infralib-testing>.

Contains GO terratest and terraform software with an entrypoint.

## Usage:
Place go unit test files in "test/" folder of the terraform project and then in the project root run:
```
docker run -it --rm -v "$(pwd)":"/app" -w /app  entigolabs/entigo-infralib-testing
```

## Building locally

The Dockerfile warms its Go module and build caches from the shared test
library in `common/` and the module tests under `modules/`, both outside this
directory. Pass them as named build contexts, as the CI build does:
```
cd images/test
docker buildx build \
  --build-context common=../../common \
  --build-context modules=../../modules \
  -t entigolabs/entigo-infralib-test:local .
```
Without them the `common` and `modules` stages are empty and the build fails
at `go mod download`. See `warm-go-cache.sh` for what gets cached and why.
