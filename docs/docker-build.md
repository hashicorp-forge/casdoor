# Docker build guide — HCP Casdoor

This document describes how to build the project's Docker images locally (both the standard runtime image and the "all-in-one" image), how to run them for local testing, and how to push them to a registry (ECR) if desired.

Files of interest

- Dockerfile: [Dockerfile](Dockerfile)
- CI build workflow: [.github/workflows/build.yml](.github/workflows/build.yml)

Overview

- The repository Dockerfile is a multi-stage build that defines these notable stages:
  - `FRONT` — builds the frontend under `/web` (Node/Yarn).
  - `BACK` — builds the Go backend and runs `./build.sh` to produce server artifacts.
  - `STANDARD` — minimal Alpine runtime image containing the server binary + web build.
  - `ALLINONE` — Debian-based image that bundles server, web build and entrypoint for quick trials.

Prerequisites

- Docker Engine (Docker Desktop on macOS/Windows or Docker on Linux).
- Docker Buildx (modern Docker includes buildx; Docker Desktop exposes it).
- For pushing to ECR: the AWS CLI configured locally with credentials that have ECR permissions, or other valid login method.

Quick local setup (one-time)

1. Create/use a buildx builder (recommended):

```bash
docker buildx create --name mybuilder --use || docker buildx use mybuilder
docker buildx inspect --bootstrap
```

2. (Optional) Ensure BuildKit is enabled if you rely on it:

```bash
export DOCKER_BUILDKIT=1
```

Build the ALL-IN-ONE image (single-platform, load into local daemon)

This is the same `ALLINONE` target used in CI.

```bash
docker buildx build --platform linux/amd64 --target ALLINONE --load \
  --pull --progress=plain \
  -t casdoor-all-in-one:local .
```

Notes:

- `--load` places the resulting image into your local Docker daemon (works for a single platform only).
- `--pull` refreshes base images before building.
- `--progress=plain` gives more readable logs for debugging build failures.

Fallback (simpler) build (non-buildx)

If you're running on a Linux host that matches the target platform, you can use the classic build command:

```bash
docker build --target ALLINONE -t casdoor-all-in-one:local .
```

Build the STANDARD runtime image (single-platform)

```bash
docker buildx build --platform linux/amd64 --target STANDARD --load -t casdoor:local .
```

Run the image locally

```bash
docker run --rm -it -p 8000:8000 casdoor-all-in-one:local
```

The `ALLINONE` image's `ENTRYPOINT`/`CMD` execute `/docker-entrypoint.sh` which starts the server (the image exposes the service on port 8000 by default). Use `docker logs <container>` if it exits immediately to see the startup errors.

Mount local config or persist logs

To substitute the configuration file or persist logs, mount host paths into the container:

```bash
docker run --rm -it -p 8000:8000 \
  -v "$PWD/conf/app.conf":/conf/app.conf \
  -v "$PWD/logs":/logs \
  casdoor-all-in-one:local
```

Build multi-architecture images and push to a registry

To build multi-arch and push to a remote registry (replace `<registry>` with your target):

```bash
docker buildx build --platform linux/amd64,linux/arm64 --target ALLINONE \
  --push -t <registry>/casdoor-all-in-one:latest .
```

Pushing to AWS ECR (example)

Local push to Amazon ECR requires valid AWS credentials locally. Example (replace `<account>` and region if needed):

```bash
# login
aws ecr get-login-password --region us-east-2 | docker login --username AWS --password-stdin <account>.dkr.ecr.us-east-2.amazonaws.com

# tag and push (single-platform)
docker tag casdoor-all-in-one:local <account>.dkr.ecr.us-east-2.amazonaws.com/casdoor-all-in-one:latest
docker push <account>.dkr.ecr.us-east-2.amazonaws.com/casdoor-all-in-one:latest

# or buildx direct push for multi-arch
docker buildx build --platform linux/amd64,linux/arm64 --target ALLINONE \
  --push -t <account>.dkr.ecr.us-east-2.amazonaws.com/casdoor-all-in-one:latest .
```

Note: In CI the repository uses Doormat to obtain AWS credentials and then `docker/build-push-action` to publish images. Locally you will either need `aws configure` credentials or another method to authenticate to ECR.

Verification and debug commands

```bash
docker images | grep casdoor
docker ps -a
docker logs <container-id>
docker inspect <image-or-container>
```

Common problems & tips

- Build is slow: enable caching or keep a persistent buildx builder (`docker buildx create --use`) to reuse layers.
- Platform mismatch: use `--platform linux/amd64` for x86_64 builds on macOS/Apple Silicon (via emulation) or build native arm64 if you need it.
- Entrypoint/script errors: inspect `/docker-entrypoint.sh` inside the image to verify permissions and content.
- Not enough disk space: remove dangling images `docker system prune -af` carefully.

Why the Dockerfile works as-is

- The `BACK` stage runs `./build.sh` during the build to compile the server. That means you do not need to pre-build the binary on the host; the container build runs the compile steps inside the builder stage.

See also

- The project Dockerfile: [Dockerfile](Dockerfile)
- The CI build & release flow that uses the same targets: [.github/workflows/build.yml](.github/workflows/build.yml)

If you want, I can run a local build on your machine now (it will use your Docker daemon). Say "yes" and I will proceed, or tell me which image/target and platform you want built.
