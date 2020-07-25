.DEFAULT_GOAL := default

IMAGE ?= averagemarcus/kube-image-prefetch:latest

.PHONY: test # Run all tests, linting and format checks
test: lint check-format run-tests

.PHONY: lint # Perform lint checks against code
lint:
	@go vet && golint -set_exit_status ./...

.PHONY: check-format # Checks code formatting and returns a non-zero exit code if formatting errors found
check-format:
	@gofmt -e -l .

.PHONY: format # Performs automatic format fixes on all code
format:
	@gofmt -s -w .

.PHONY: run-tests # Runs all tests
run-tests:
	@go test

.PHONY: fetch-deps # Fetch all project dependencies
fetch-deps:
	@go mod tidy

.PHONY: build # Build the project
build: lint check-format fetch-deps
	@go build -o kube-image-prefetch main.go

.PHONY: docker-build # Build the docker image
docker-build:
	@docker build -t $(IMAGE) .

.PHONY: docker-publish # Push the docker image to the remote registry
docker-publish:
	@docker push $(IMAGE)

.PHONY: run # Run the application
run:
	@go run main.go

.PHONY: help # Show this list of commands
help:
	@echo "kube-image-prefetch"
	@echo "Usage: make [target]"
	@echo ""
	@echo "target	description" | expand -t20
	@echo "-----------------------------------"
	@grep '^.PHONY: .* #' Makefile | sed 's/\.PHONY: \(.*\) # \(.*\)/\1	\2/' | expand -t20

default: test
