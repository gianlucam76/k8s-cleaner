# Development

Our Makefile helps you with the development of new changes or fixes. [You may have a look at it](./Makefile), since not all targets are documented.

## Building

You can build the docker image locally

```bash
make docker-build
```

This will push the build to your local docker images.

## Lint

Execute lint testing:

```bash
make lint
```

## Test

Execute unit testing:

```bash
make test
```

## E2E 

Execute functional verification:

```bash
make kind-test
```

This will create a *KinD* cluster locally, build k8s-cleaner docker-image, load it and run some tests.

