.ONESHELL:

STL_NS = spirit-dev
STL_NAME = work-facilitator

# ALPINE_VERSION:=latest

CONTAINER_DEFINITION_FILE=docker/Dockerfile
CONTAINER_BUILD_ARGS:="--build-arg STL_NAME=${STL_NAME}"
# CONTAINER_BUILD_ARGS:=""
CONTAINER_IMAGE_NAME=${STL_NAME}
CONTAINER_IMAGE_TAG:=latest

help:
	@awk 'BEGIN {FS = ":.*##"; printf "Usage: make \033[36m<command> <option>\033[0m\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
	@printf "\033[1mVariables\033[0m\n"
	@grep -E '^[a-zA-Z0-9_-]+.*?### .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?### "}; {printf "  \033[36m%-21s\033[0m %s\n", $$1, $$2}'
	@# Use ##@ <section> to define a section
	@# Use ## <comment> aside of the target to get it shown in the helper
	@# Use ### <comment> to comment a variable

##@ Development
init: ## Init (one time use) the go module
	@mkdir -p src
	@cd src
	@mkdir ${STL_NAME}
	@go mod init ${STL_NS}/${STL_NAME}
packages: ## Install Go modules
	@cd src/${STL_NAME}
	@go mod tidy
test-dev: ## Execute tests
	@cd src/${STL_NAME}
	@go test -v
run-dev: ## Execute in-dev script
	@cd $$(pwd)/src/${STL_NAME} && \
	 go run . ${CMD_OPT}

completion: build-stand-alone ## create completion
	# @./dist/${STL_NAME} completion zsh > completion
	@./dist/${STL_NAME} completion zsh


##@ Development Cobra
cobra-init: ## Init cobra
	@cd src/${STL_NAME}
	@cobra-cli init
cobra-add: ## Cobre add command
	@cd src/${STL_NAME}
	@cobra-cli add ${COBRA-COMMAND}

##@ Stand Alone version
clean-sdtl: ## Build the stand alone version
	@touch ./dist/${STL_NAME}
	@go clean -C ./src/${STL_NAME}
	@go clean -C ./dist
	@rm ./dist/${STL_NAME}

build-stand-alone: clean-sdtl ## Build the stand alone version
	@go build -C ./src/${STL_NAME} -o ../../dist/${STL_NAME}

run-stand-alone: build-stand-alone ## Run script using stand alone version
	@./dist/${STL_NAME}

install: build-stand-alone ## Install the binary in its definitive location.
	@sudo cp ${STL_NAME} /usr/local/bin/${STL_NAME}
	@echo "Package successfully installed"
	@rm ${STL_NAME} ${STL_NAME}.zip

##@ docker
docker-lint: ## docker lint
	@docker build -f ${CONTAINER_DEFINITION_FILE} "${CONTAINER_BUILD_ARGS}" -t ${CONTAINER_IMAGE_NAME}:lint --target lint $$(pwd)/.
docker-lint-run: docker-lint ## docker lint run
	@docker run -it --rm --entrypoint sh ${CONTAINER_IMAGE_NAME}:lint
docker-build2: ## docker build2
	@docker build -f ${CONTAINER_DEFINITION_FILE} "${CONTAINER_BUILD_ARGS}" -t ${CONTAINER_IMAGE_NAME}:build2 --target build2 $$(pwd)/.
docker-build2-run: docker-build2 ## docker build2 run
	@docker run -it --rm --entrypoint sh ${CONTAINER_IMAGE_NAME}:build2

docker-build: ## docker build (Outdated)
	@docker build --platform linux/amd64 -f ${CONTAINER_DEFINITION_FILE} "${CONTAINER_BUILD_ARGS}" -t ${CONTAINER_IMAGE_NAME}:amd64 $$(pwd)/.
	@docker build --platform linux/arm64 -f ${CONTAINER_DEFINITION_FILE} "${CONTAINER_BUILD_ARGS}" -t ${CONTAINER_IMAGE_NAME}:arm64 $$(pwd)/.
	@docker images | grep ${CONTAINER_IMAGE_NAME}
docker-run: docker-build ## jump in the container
	@docker run -it --rm --entrypoint sh ${CONTAINER_IMAGE_NAME}:${CONTAINER_IMAGE_TAG}

docker-build-grab: ## jump in the container
	@docker run -d --name ${CONTAINER_IMAGE_NAME}-amd64 ${CONTAINER_IMAGE_NAME}:amd64
	@docker run -d --name ${CONTAINER_IMAGE_NAME}-arm64 ${CONTAINER_IMAGE_NAME}:arm64
	@docker cp ${CONTAINER_IMAGE_NAME}-amd64:/code/dist/${STL_NAME}-amd64 ./dist
	@docker cp ${CONTAINER_IMAGE_NAME}-arm64:/code/dist/${STL_NAME}-arm64 ./dist
	@docker stop ${CONTAINER_IMAGE_NAME}-amd64
	@docker stop ${CONTAINER_IMAGE_NAME}-arm64
	@docker rm -f ${CONTAINER_IMAGE_NAME}-amd64
	@docker rm -f ${CONTAINER_IMAGE_NAME}-arm64

##@ docker-compose
compose-build: ## docker-compose build
	@cd ./docker && docker-compose build

compose-run: compose-build ## Run the container
	@cd ./docker && \
		docker-compose stop && \
		docker-compose rm -f && \
		docker-compose up

compose-exec: compose-build ## Jump in the container
	@cd ./docker
	@docker-compose run --entrypoint sh ${STL_NAME}
	@cd ..
