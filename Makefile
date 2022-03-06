DOCKER_REPO := docker.io/act28/vwap

-include .makefiles/Makefile
-include .makefiles/pkg/go/v1/Makefile
-include .makefiles/pkg/docker/v1/Makefile

.makefiles/%:
	@curl -sfL https://makefiles.dev/v1 | bash /dev/stdin "$@"

run: docker
	@docker run -it --rm --name act28_vwap docker.io/act28/vwap:dev

run-cli:
	@go run ./cmd/server/main.go
