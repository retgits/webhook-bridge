# Description: Makefile
# Author: retgits
# Last Updated: 2019-02-03

#--- Variables ---
## The name of the user for Docker Hub.
DOCKERUSER=retgits
## Get the name of the project.
PROJECT=webhook-bridge
## Create a list of all packages in this repository.
PACKAGES=$(shell goc list ./... | grep -v "vendor")

#--- Help ---
.PHONY: help
help: ## Displays the help for each target (this message).
	@echo 
	@echo Makefile targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' Makefile | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
	@echo 

#--- Linting targets ---
.PHONY: fmt
fmt: ## Fmt runs the commands 'gofmt -l -w' and 'gofmt -s -w' and prints the names of the files that are modified.
	env GO111MODULE=on goc fmt ./...
	env GO111MODULE=on gofmt -s -w .

.PHONY: vet
vet: ## Vet examines Go source code and reports suspicious constructs.
	env GO111MODULE=on goc vet ./...

.PHONY: lint
lint: ## Lint examines Go source code and prints style mistakes for all packages.
	env GO111MODULE=on golint -set_exit_status $(ALL_PACKAGES)

#--- Setup targets ---
.PHONY: setup-tools
setup-tools: ## Get the tools needed to test and validate the project.
	goc get -u golang.org/x/lint/golint
	goc get -u github.com/gojp/goreportcard/cmd/goreportcard-cli
	curl -fL https://getgoc.gocenter.io | sh

.PHONY: setup-test
setup-test: ## Make preparations to be able to run tests.
	mkdir -p ${TESTDIR}

.PHONY: setup-deps
setup-deps: ## Get all the Go dependencies.
	goc get -u ./...

#--- Test targets ---
.PHONY: test
test: ## Run all test targets.

.PHONY: go-test
go-test: ## Run all testcases.
	env TESTDIR=${TESTDIR} goc test -race ./...

.PHONY: go-test-coverage
go-test-coverage: ## Run all test cases and generate a coverage report.
	@echo "mode: count" > coverage-all.out

	$(foreach pkg, $(PACKAGES),\
	env TESTDIR=${TESTDIR} goc test -coverprofile=coverage.out -covermode=count $(pkg);\
	tail -n +2 coverage.out >> coverage-all.out;)
	goc tool cover -html=coverage-all.out -o out/coverage.html

.PHONY: go-score
go-score: ## Get a score based on GoReportcard.
	goreportcard-cli -v

#--- Build targets ---
compile-jenkins-mac: ## Compiles the Jenkins agent to run on macOS.
	mkdir -p ./out/jenkins
	env GO111MODULE=on GOOS=darwin CGO_ENABLED=0 goc build -v -a -installsuffix cgo -o out/jenkins/jenkins ./cmd/jenkins/*.go

compile-jenkins-docker: ## Compiles the Jenkins agent and builds a Docker image.
	mkdir -p ./out/jenkins
	env GO111MODULE=on GOOS=linux CGO_ENABLED=0 goc build -v -a -installsuffix cgo -o out/jenkins/jenkins ./cmd/jenkins/*.go
	cp ./cmd/jenkins/Dockerfile ./out/jenkins
	cd ./out/jenkins && docker build -t ${DOCKERUSER}/webhook-jenkins .

compile-pubnub-mac: ## Compiles the PubNub agent to run on macOS.
	mkdir -p ./out/pubnub
	env GO111MODULE=on GOOS=darwin CGO_ENABLED=0 goc build -v -a -installsuffix cgo -o out/pubnub/pubnub ./cmd/pubnub/*.go

compile-pubnub-docker: ## Compiles the PubNub agent and builds a Docker image.
	mkdir -p ./out/pubnub
	env GO111MODULE=on GOOS=linux CGO_ENABLED=0 goc build -v -a -installsuffix cgo -o out/pubnub/pubnub ./cmd/pubnub/*.go
	cp ./cmd/pubnub/Dockerfile ./out/pubnub
	cd ./out/pubnub && docker build -t ${DOCKERUSER}/webhook-pubnub .