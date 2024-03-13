## Docker image
The Docker image is available at <https://hub.docker.com/r/entigolabs/entigo-infralib-testing>.

Contains GO terratest and terraform software with an entrypoint.

## Usage:
Place go unit test files in "test/" folder of the terraform project and then in the project root run:
```
docker run -it --rm -v "$(pwd)":"/app" -w /app  entigolabs/entigo-infralib-testing
```
