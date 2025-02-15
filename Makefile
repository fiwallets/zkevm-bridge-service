include version.mk

DOCKER_COMPOSE := docker compose -f docker-compose.yml
DOCKER_COMPOSE_STATE_DB := zkevm-state-db
DOCKER_COMPOSE_POOL_DB := zkevm-pool-db
DOCKER_COMPOSE_BRIDGE_DB := zkevm-bridge-db
DOCKER_COMPOSE_STATE_DB_2 := zkevm-state-db-2
DOCKER_COMPOSE_POOL_DB_2 := zkevm-pool-db-2
DOCKER_COMPOSE_BRIDGE_DB_2 := zkevm-bridge-db-2
DOCKER_COMPOSE_ZKEVM_NODE := zkevm-node
DOCKER_COMPOSE_ZKEVM_NODE_V1TOV2 := zkevm-node-v1tov2
DOCKER_COMPOSE_ZKEVM_NODE_1 := zkevm-node-1
DOCKER_COMPOSE_ZKEVM_NODE_2 := zkevm-node-2
DOCKER_COMPOSE_ZKEVM_AGGREGATOR_V1TOV2 := zkevm-aggregator-v1tov2
DOCKER_COMPOSE_L1_NETWORK := zkevm-mock-l1-network
DOCKER_COMPOSE_L1_NETWORK_MULTI_ROLLUP := zkevm-mock-l1-network-multi-rollup
DOCKER_COMPOSE_L1_NETWORK_V1TOV2 := zkevm-v1tov2-l1-network
DOCKER_COMPOSE_ZKPROVER := zkevm-prover
DOCKER_COMPOSE_ZKPROVER_V1TOV2 := zkevm-prover-v1tov2
DOCKER_COMPOSE_ZKPROVER_1 := zkevm-prover-1
DOCKER_COMPOSE_ZKPROVER_2 := zkevm-prover-2
DOCKER_COMPOSE_BRIDGE := zkevm-bridge-service
DOCKER_COMPOSE_BRIDGE_V1TOV2 := zkevm-bridge-service-v1tov2
DOCKER_COMPOSE_BRIDGE_1 := zkevm-bridge-service-1
DOCKER_COMPOSE_BRIDGE_2 := zkevm-bridge-service-2
DOCKER_COMPOSE_BRIDGE_3 := zkevm-bridge-service-3
DOCKER_COMPOSE_BRIDGE_SOVEREIGN_CHAIN := zkevm-bridge-service-sovereign-chain
DOCKER_COMPOSE_AGGORACLE := aggoracle

RUN_STATE_DB := $(DOCKER_COMPOSE) up -d $(DOCKER_COMPOSE_STATE_DB)
RUN_POOL_DB := $(DOCKER_COMPOSE) up -d $(DOCKER_COMPOSE_POOL_DB)
RUN_BRIDGE_DB := $(DOCKER_COMPOSE) up -d $(DOCKER_COMPOSE_BRIDGE_DB)
RUN_STATE_DB_2 := $(DOCKER_COMPOSE) up -d $(DOCKER_COMPOSE_STATE_DB_2)
RUN_POOL_DB_2 := $(DOCKER_COMPOSE) up -d $(DOCKER_COMPOSE_POOL_DB_2)
RUN_BRIDGE_DB_2 := $(DOCKER_COMPOSE) up -d $(DOCKER_COMPOSE_BRIDGE_DB_2)
RUN_DBS := ${RUN_BRIDGE_DB} && ${RUN_STATE_DB} && ${RUN_POOL_DB}
RUN_DBS_2 := ${RUN_BRIDGE_DB_2} && ${RUN_STATE_DB_2} && ${RUN_POOL_DB_2}
RUN_NODE := $(DOCKER_COMPOSE) up -d $(DOCKER_COMPOSE_ZKEVM_NODE)
RUN_NODE_1 := $(DOCKER_COMPOSE) up -d $(DOCKER_COMPOSE_ZKEVM_NODE_1)
RUN_NODE_2 := $(DOCKER_COMPOSE) up -d $(DOCKER_COMPOSE_ZKEVM_NODE_2)
RUN_NODE_V1TOV2 := $(DOCKER_COMPOSE) up -d $(DOCKER_COMPOSE_ZKEVM_NODE_V1TOV2)
RUN_AGGREGATOR_V1TOV2 := $(DOCKER_COMPOSE) up -d $(DOCKER_COMPOSE_ZKEVM_AGGREGATOR_V1TOV2)
RUN_L1_NETWORK := $(DOCKER_COMPOSE) up -d $(DOCKER_COMPOSE_L1_NETWORK)
RUN_L1_NETWORK_V1TOV2 := $(DOCKER_COMPOSE) up -d $(DOCKER_COMPOSE_L1_NETWORK_V1TOV2)
RUN_L1_NETWORK_MULTI_ROLLUP := $(DOCKER_COMPOSE) up -d $(DOCKER_COMPOSE_L1_NETWORK_MULTI_ROLLUP)
RUN_ZKPROVER := $(DOCKER_COMPOSE) up -d $(DOCKER_COMPOSE_ZKPROVER)
RUN_ZKPROVER_2 := $(DOCKER_COMPOSE) up -d $(DOCKER_COMPOSE_ZKPROVER_2)
RUN_ZKPROVER_1 := $(DOCKER_COMPOSE) up -d $(DOCKER_COMPOSE_ZKPROVER_1)
RUN_ZKPROVER_V1TOV2 := $(DOCKER_COMPOSE) up -d $(DOCKER_COMPOSE_ZKPROVER_V1TOV2)
RUN_BRIDGE := $(DOCKER_COMPOSE) up -d $(DOCKER_COMPOSE_BRIDGE)
RUN_BRIDGE_1 := $(DOCKER_COMPOSE) up -d $(DOCKER_COMPOSE_BRIDGE_1)
RUN_BRIDGE_2 := $(DOCKER_COMPOSE) up -d $(DOCKER_COMPOSE_BRIDGE_2)
RUN_BRIDGE_3 := $(DOCKER_COMPOSE) up -d $(DOCKER_COMPOSE_BRIDGE_3)
RUN_BRIDGE_V1TOV2 := $(DOCKER_COMPOSE) up -d $(DOCKER_COMPOSE_BRIDGE_V1TOV2)
RUN_BRIDGE_SOVEREIGN_CHAIN := $(DOCKER_COMPOSE) up -d $(DOCKER_COMPOSE_BRIDGE_SOVEREIGN_CHAIN)
RUN_AGGORACLE := $(DOCKER_COMPOSE) up -d $(DOCKER_COMPOSE_AGGORACLE)

STOP_NODE_DB := $(DOCKER_COMPOSE) stop $(DOCKER_COMPOSE_NODE_DB) && $(DOCKER_COMPOSE) rm -f $(DOCKER_COMPOSE_NODE_DB)
STOP_BRIDGE_DB := $(DOCKER_COMPOSE) stop $(DOCKER_COMPOSE_BRIDGE_DB) && $(DOCKER_COMPOSE) rm -f $(DOCKER_COMPOSE_BRIDGE_DB)
STOP_DBS := ${STOP_NODE_DB} && ${STOP_BRIDGE_DB} && ${STOP_NODE_DB_2} && ${STOP_BRIDGE_DB_2} && ${STOP_POOL_DB} && ${STOP_POOL_DB_2}
STOP_NODE := $(DOCKER_COMPOSE) stop $(DOCKER_COMPOSE_ZKEVM_NODE) && $(DOCKER_COMPOSE) rm -f $(DOCKER_COMPOSE_ZKEVM_NODE)
STOP_NODE_1 := $(DOCKER_COMPOSE) stop $(DOCKER_COMPOSE_ZKEVM_NODE_1) && $(DOCKER_COMPOSE) rm -f $(DOCKER_COMPOSE_ZKEVM_NODE_1)
STOP_NODE_2 := $(DOCKER_COMPOSE) stop $(DOCKER_COMPOSE_ZKEVM_NODE_2) && $(DOCKER_COMPOSE) rm -f $(DOCKER_COMPOSE_ZKEVM_NODE_2)
STOP_NODE_V1TOV2 := $(DOCKER_COMPOSE) stop $(DOCKER_COMPOSE_ZKEVM_NODE_V1TOV2) && $(DOCKER_COMPOSE) rm -f $(DOCKER_COMPOSE_ZKEVM_NODE_V1TOV2)
STOP_AGGREGATOR_V1TOV2 := $(DOCKER_COMPOSE) stop $(DOCKER_COMPOSE_ZKEVM_AGGREGATOR_V1TOV2) && $(DOCKER_COMPOSE) rm -f $(DOCKER_COMPOSE_ZKEVM_AGGREGATOR_V1TOV2)
STOP_NETWORK := $(DOCKER_COMPOSE) stop $(DOCKER_COMPOSE_L1_NETWORK) && $(DOCKER_COMPOSE) rm -f $(DOCKER_COMPOSE_L1_NETWORK)
STOP_NETWORK_MULTI_ROLLUP := $(DOCKER_COMPOSE) stop $(DOCKER_COMPOSE_L1_NETWORK_MULTI_ROLLUP) && $(DOCKER_COMPOSE) rm -f $(DOCKER_COMPOSE_L1_NETWORK_MULTI_ROLLUP)
STOP_NETWORK_V1TOV2 := $(DOCKER_COMPOSE) stop $(DOCKER_COMPOSE_L1_NETWORK_V1TOV2) && $(DOCKER_COMPOSE) rm -f $(DOCKER_COMPOSE_L1_NETWORK_V1TOV2)
STOP_ZKPROVER := $(DOCKER_COMPOSE) stop $(DOCKER_COMPOSE_ZKPROVER) && $(DOCKER_COMPOSE) rm -f $(DOCKER_COMPOSE_ZKPROVER)
STOP_ZKPROVER_1 := $(DOCKER_COMPOSE) stop $(DOCKER_COMPOSE_ZKPROVER_1) && $(DOCKER_COMPOSE) rm -f $(DOCKER_COMPOSE_ZKPROVER_1)
STOP_ZKPROVER_2 := $(DOCKER_COMPOSE) stop $(DOCKER_COMPOSE_ZKPROVER_2) && $(DOCKER_COMPOSE) rm -f $(DOCKER_COMPOSE_ZKPROVER_2)
STOP_ZKPROVER_V1TOV2 := $(DOCKER_COMPOSE) stop $(DOCKER_COMPOSE_ZKPROVER_V1TOV2) && $(DOCKER_COMPOSE) rm -f $(DOCKER_COMPOSE_ZKPROVER_V1TOV2)
STOP_BRIDGE := $(DOCKER_COMPOSE) stop $(DOCKER_COMPOSE_BRIDGE) && $(DOCKER_COMPOSE) rm -f $(DOCKER_COMPOSE_BRIDGE)
STOP_BRIDGE_1 := $(DOCKER_COMPOSE) stop $(DOCKER_COMPOSE_BRIDGE_1) && $(DOCKER_COMPOSE) rm -f $(DOCKER_COMPOSE_BRIDGE_1)
STOP_BRIDGE_2 := $(DOCKER_COMPOSE) stop $(DOCKER_COMPOSE_BRIDGE_2) && $(DOCKER_COMPOSE) rm -f $(DOCKER_COMPOSE_BRIDGE_2)
STOP_BRIDGE_3 := $(DOCKER_COMPOSE) stop $(DOCKER_COMPOSE_BRIDGE_3) && $(DOCKER_COMPOSE) rm -f $(DOCKER_COMPOSE_BRIDGE_3)
STOP_BRIDGE_V1TOV2 := $(DOCKER_COMPOSE) stop $(DOCKER_COMPOSE_BRIDGE_V1TOV2) && $(DOCKER_COMPOSE) rm -f $(DOCKER_COMPOSE_BRIDGE_V1TOV2)
STOP_BRIDGE_SOVEREIGN_CHAIN := $(DOCKER_COMPOSE) stop $(DOCKER_COMPOSE_BRIDGE_SOVEREIGN_CHAIN) && $(DOCKER_COMPOSE) rm -f $(DOCKER_COMPOSE_BRIDGE_SOVEREIGN_CHAIN)
STOP_AGGORACLE := $(DOCKER_COMPOSE) stop $(DOCKER_COMPOSE_AGGORACLE) && $(DOCKER_COMPOSE) rm -f $(DOCKER_COMPOSE_AGGORACLE)
STOP := $(DOCKER_COMPOSE) down --remove-orphans

LDFLAGS += -X 'github.com/fiwallets/zkevm-bridge-service.Version=$(VERSION)'
LDFLAGS += -X 'github.com/fiwallets/zkevm-bridge-service.GitRev=$(GITREV)'
LDFLAGS += -X 'github.com/fiwallets/zkevm-bridge-service.GitBranch=$(GITBRANCH)'
LDFLAGS += -X 'github.com/fiwallets/zkevm-bridge-service.BuildDate=$(DATE)'

GO_BASE := $(shell pwd)
GO_BIN := $(GO_BASE)/dist
GO_ENV_VARS := GO_BIN=$(GO_BIN)
GO_BINARY := zkevm-bridge
GO_CMD := $(GO_BASE)/cmd
GO_DEPLOY_SCRIPT := $(GO_BASE)/test/scripts/deploytool
GO_DEPLOY_SCRIPT_BINARY := test-deploy-tool
GO_DEPLOY_AUTOCLAIMER := $(GO_BASE)/autoclaimservice
GO_DEPLOY_AUTOCLAIMER_BINARY := zkevm-autoclaimer

LINT := $$(go env GOPATH)/bin/golangci-lint run --timeout=5m -E whitespace -E gosec -E gci -E misspell -E gomnd -E gofmt -E goimports --exclude-use-default=false --max-same-issues 0
BUILD := $(GO_ENV_VARS) go build -ldflags "all=$(LDFLAGS)" -o $(GO_BIN)/$(GO_BINARY) $(GO_CMD)
BUILDSCRIPTEPLOY := $(GO_ENV_VARS) go build -o $(GO_BIN)/$(GO_DEPLOY_SCRIPT_BINARY) $(GO_DEPLOY_SCRIPT)
BUILDAUTOCLAIMER := $(GO_ENV_VARS) go build -o $(GO_BIN)/$(GO_DEPLOY_AUTOCLAIMER_BINARY) $(GO_DEPLOY_AUTOCLAIMER)

.PHONY: build
build: ## Build the binary locally into ./dist
	$(BUILD)
	$(BUILDSCRIPTEPLOY)
	$(BUILDAUTOCLAIMER)

.PHONY: lint
lint: ## runs linter
	$(LINT)

.PHONY: install-git-hooks
install-git-hooks: ## Moves hook files to the .git/hooks directory
	cp .github/hooks/* .git/hooks

.PHONY: test
test: ## Runs only short tests without checking race conditions
	$(STOP_BRIDGE_DB) || true
	$(RUN_BRIDGE_DB); sleep 3
	trap '$(STOP_BRIDGE_DB)' EXIT; go test --cover -short -p 1 ./...

.PHONY: install-linter
install-linter: ## Installs the linter
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.54.2

.PHONY: build-docker
build-docker: ## Builds a docker image with the zkevm bridge binary
	docker build -t zkevm-bridge-service -f ./Dockerfile .

.PHONY: build-docker-e2e-real_network
build-docker-e2e-real_network:  build-docker-e2e-real_network-erc20 build-docker-e2e-real_network-msg ## Builds a docker image with the zkevm bridge binary for e2e tests with real network
	

.PHONY: build-docker-e2e-real_network-erc20
build-docker-e2e-real_network-erc20: ## Builds a docker image with the zkevm bridge binary for e2e ERC20 tests with real network
	docker build  -f DockerfileE2ETest .  -t bridge-e2e-realnetwork-erc20 --target ERC20

.PHONY: build-docker-e2e-real_network-msg
build-docker-e2e-real_network-msg: ## Builds a docker image with the zkevm bridge binary for e2e BridgeMessage tests with real network
	docker build  -f DockerfileE2ETest .  -t bridge-e2e-realnetwork-msg --target MSG

.PHONY: run-db-node
run-db-node: ## Runs the node database
	$(RUN_NODE_DB)

.PHONY: stop-db-node
stop-db-node: ## Stops the node database
	$(STOP_NODE_DB)

.PHONY: run-db-bridge
run-db-bridge: ## Runs the node database
	$(RUN_BRIDGE_DB)

.PHONY: stop-db-bridge
stop-db-bridge: ## Stops the node database
	$(STOP_BRIDGE_DB)

.PHONY: run-dbs
run-dbs: ## Runs the node database
	$(RUN_DBS)

.PHONY: run-dbs-2
run-dbs-2: ## Runs the node database
	$(RUN_DBS_2)

.PHONY: stop-dbs
stop-dbs: ## Stops the node database
	$(STOP_DBS)
	$(STOP_DBS_2)

.PHONY: run-node
run-node: ## Runs the node
	$(RUN_NODE)

.PHONY: run-node-2
run-node-2: ## Runs the node
	$(RUN_NODE_2)

.PHONY: run-node-1
run-node-1: ## Runs the node
	$(RUN_NODE_1)

.PHONY: stop-node
stop-node: ## Stops the node
	$(STOP_NODE)

.PHONY: stop-node-2
stop-node-2: ## Stops the node
	$(STOP_NODE_2)

.PHONY: stop-node-1
stop-node-1: ## Stops the node
	$(STOP_NODE_1)

.PHONY: run-network
run-network: ## Runs the l1 network
	$(RUN_L1_NETWORK)

.PHONY: run-network-multi-rollup
run-network-multi-rollup: ## Runs the l1 network
	$(RUN_L1_NETWORK_MULTI_ROLLUP)

.PHONY: stop-network
stop-network: ## Stops the l1 network
	$(STOP_NETWORK)
	$(STOP_NETWORK_MULTI_ROLLUP)

.PHONY: run-node-v1tov2
run-node-v1tov2: ## Runs the node
	$(RUN_NODE_V1TOV2)

.PHONY: stop-node-v1tov2
stop-node-v1tov2: ## Stops the node
	$(STOP_NODE_V1TOV2)

.PHONY: run-aggregator-v1tov2
run-aggregator-v1tov2: ## Runs the aggregator
	$(RUN_AGGREGATOR_V1TOV2)

.PHONY: stop-aggregator-v1tov2
stop-aggregator-v1tov2: ## Stops the aggregator
	$(STOP_AGGREGATOR_V1TOV2)

.PHONY: run-network-v1tov2
run-network-v1tov2: ## Runs the l1 network
	$(RUN_L1_NETWORK_V1TOV2)

.PHONY: stop-network-v1tov2
stop-network-v1tov2: ## Stops the l1 network
	$(STOP_NETWORK_V1TOV2)

.PHONY: run-prover
run-prover: ## Runs the zk prover
	$(RUN_ZKPROVER)

.PHONY: stop-prover
stop-prover: ## Stops the zk prover
	$(STOP_ZKPROVER)

.PHONY: run-prover-1
run-prover-1: ## Runs the zk prover
	$(RUN_ZKPROVER_1)

.PHONY: stop-prover-1
stop-prover-1: ## Stops the zk prover
	$(STOP_ZKPROVER_1)

.PHONY: run-prover-2
run-prover-2: ## Runs the zk prover
	$(RUN_ZKPROVER_2)

.PHONY: stop-prover-2
stop-prover-2: ## Stops the zk prover
	$(STOP_ZKPROVER_2)

.PHONY: run-prover-v1tov2
run-prover-v1tov2: ## Runs the zk prover
	$(RUN_ZKPROVER_V1TOV2)

.PHONY: stop-prover-v1tov2
stop-prover-v1tov2: ## Stops the zk prover
	$(STOP_ZKPROVER_V1TOV2)	

.PHONY: run-bridge
run-bridge: ## Runs the bridge service
	$(RUN_BRIDGE)

.PHONY: stop-bridge
stop-bridge: ## Stops the bridge service
	$(STOP_BRIDGE)

.PHONY: run-bridge-1
run-bridge-1: ## Runs the bridge service
	$(RUN_BRIDGE_1)

.PHONY: stop-bridge-1
stop-bridge-1: ## Stops the bridge service
	$(STOP_BRIDGE_1)

.PHONY: run-bridge-2
run-bridge-2: ## Runs the bridge service
	$(RUN_BRIDGE_2)

.PHONY: stop-bridge-2
stop-bridge-2: ## Stops the bridge service
	$(STOP_BRIDGE_2)

.PHONY: run-bridge-3
run-bridge-3: ## Runs the bridge service
	$(RUN_BRIDGE_3)

.PHONY: stop-bridge-3
stop-bridge-3: ## Stops the bridge service
	$(STOP_BRIDGE_3)

.PHONY: run-bridge-v1tov2
run-bridge-v1tov2: ## Runs the bridge service
	$(RUN_BRIDGE_V1TOV2)

.PHONY: stop-bridge-v1tov2
stop-bridge-v1tov2: ## Stops the bridge service
	$(STOP_BRIDGE_V1TOV2)

.PHONY: run-bridge-sovereign-chain
run-bridge-sovereign-chain: ## Runs the bridge service
	$(RUN_BRIDGE_SOVEREIGN_CHAIN)

.PHONY: stop-bridge-sovereign-chain
stop-bridge-sovereign-chain: ## Stops the bridge service
	$(STOP_BRIDGE_SOVEREIGN_CHAIN)

.PHONY: run-aggoracle
run-aggoracle: ## Runs the bridge service
	$(RUN_AGGORACLE)

.PHONY: stop-aggoracle
stop-aggoracle: ## Stops the bridge service
	$(STOP_AGGORACLE)

.PHONY: stop
stop: ## Stops all services
	$(STOP)

.PHONY: restart
restart: stop run ## Executes `make stop` and `make run` commands

.PHONY: run
run: ## runs all services
	$(RUN_DBS)
	$(RUN_L1_NETWORK)
	sleep 5
	$(RUN_ZKPROVER)
	sleep 3
	$(RUN_NODE)
	sleep 7
	$(RUN_BRIDGE)

.PHONY: run-1
run-1: ## runs all services
	$(RUN_DBS)
	$(RUN_L1_NETWORK_MULTI_ROLLUP)
	sleep 5
	$(RUN_ZKPROVER_1)
	sleep 3
	$(RUN_NODE_1)
	sleep 7
	$(RUN_BRIDGE_1)

.PHONY: run-2
run-2: ## runs all services
	$(RUN_DBS_2)
	$(RUN_L1_NETWORK_MULTI_ROLLUP)
	sleep 5
	$(RUN_ZKPROVER_2)
	sleep 3
	$(RUN_NODE_2)
	sleep 7
	$(RUN_BRIDGE_2)

.PHONY: run-multi
run-multi: ## runs all services
	$(RUN_DBS)
	$(RUN_DBS_2)
	$(RUN_L1_NETWORK_MULTI_ROLLUP)
	sleep 5
	$(RUN_ZKPROVER_1)
	$(RUN_ZKPROVER_2)
	sleep 3
	$(RUN_NODE_1)
	$(RUN_NODE_2)
	sleep 7
	$(RUN_BRIDGE_1)
	$(RUN_BRIDGE_2)

.PHONY: run-multi-single-bridge
run-multi-single-bridge: ## runs all services
	$(RUN_DBS)
	${RUN_STATE_DB_2}
	${RUN_POOL_DB_2}
	$(RUN_L1_NETWORK_MULTI_ROLLUP)
	sleep 5
	$(RUN_ZKPROVER_1)
	$(RUN_ZKPROVER_2)
	sleep 3
	$(RUN_NODE_1)
	$(RUN_NODE_2)
	sleep 7
	$(RUN_BRIDGE_3)

.PHONY: run-bridge-dependencies
run-bridge-dependencies: stop ## runs all services
	$(RUN_DBS)
	$(RUN_L1_NETWORK)
	sleep 5
	$(RUN_ZKPROVER)
	sleep 3
	$(RUN_NODE)

.PHONY: run-v1tov2
run-v1tov2: stop ## runs all services
	$(RUN_DBS)
	$(RUN_L1_NETWORK_V1TOV2)
	sleep 5
	$(RUN_ZKPROVER_V1TOV2)
	sleep 3
	$(RUN_NODE_V1TOV2)
	sleep 7
	$(RUN_AGGREGATOR_V1TOV2)
	$(RUN_BRIDGE_V1TOV2)

.PHONY: run-sovereign-chain
run-sovereign-chain: ## runs all services
	$(RUN_DBS)
	$(RUN_L1_NETWORK)
	sleep 5
	$(RUN_ZKPROVER)
	sleep 3
	$(RUN_NODE)
	sleep 5
	$(RUN_BRIDGE_SOVEREIGN_CHAIN)
	sleep 30
	$(RUN_AGGORACLE)

.PHONY: update-external-dependencies
update-external-dependencies: ## Updates external dependencies like images, test vectors or proto files
	go run ./scripts/cmd/... updatedeps

.PHONY: generate-code-from-proto
generate-code-from-proto:
	cd proto/src/proto/bridge/v1 && protoc --proto_path=. --proto_path=../../../../../third_party --go_out=../../../../../bridgectrl/pb --go-grpc_out=../../../../../bridgectrl/pb  --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative query.proto
	cd proto/src/proto/bridge/v1 && protoc --proto_path=. --proto_path=../../../../../third_party --grpc-gateway_out=logtostderr=true:../../../../../bridgectrl/pb --grpc-gateway_opt=paths=source_relative query.proto

.PHONY: stop-mockserver
stop-mockserver: ## Stops the mock bridge service
	$(STOP_BRIDGE_MOCK)

.PHONY: bench
bench: ## benchmark test
	$(STOP_BRIDGE_DB) || true
	$(RUN_BRIDGE_DB); sleep 3
	trap '$(STOP_BRIDGE_DB)' EXIT; go test -run=NOTEST -timeout=30m -bench=Small ./test/benchmark/...

.PHONY: bench-full
bench-full: export ZKEVM_BRIDGE_DATABASE_PORT = 5432
bench-full: ## benchmark full test
	cd test/benchmark && \
	go test -run=NOTEST -bench=Small . && \
	go test -run=NOTEST -bench=Medium . && \
	go test -run=NOTEST -timeout=30m -bench=Large .

.PHONY: test-full
test-full: build-docker stop run ## Runs all tests checking race conditions
	sleep 3
	trap '$(STOP)' EXIT; MallocNanoZone=0 go test -v -failfast -race -p 1 -timeout 2400s ./test/e2e/... -count 1 -tags='e2e'

.PHONY: test-edge
test-edge: build-docker stop run ## Runs all tests checking race conditions
	sleep 3
	trap '$(STOP)' EXIT; MallocNanoZone=0 go test -v -failfast -race -p 1 -timeout 2400s ./test/e2e/... -count 1 -tags='edge'

.PHONY: test-multiplerollups
test-multiplerollups: build-docker stop run-multi-single-bridge ## Runs all tests checking race conditions
	sleep 3
	trap '$(STOP)' EXIT; MallocNanoZone=0 go test -v -failfast -race -p 1 -timeout 2400s ./test/e2e/... -count 1 -tags='multiplerollups'

.PHONY: test-l2l2
test-l2l2: build-docker stop run-multi-single-bridge ## Runs all tests checking race conditions
	sleep 3
	trap '$(STOP)' EXIT; MallocNanoZone=0 go test -v -failfast -race -p 1 -timeout 2400s ./test/e2e/... -count 1 -tags='l2l2'

.PHONY: test-autoclaiml2l2
test-autoclaiml2l2: build-docker stop run-multi-single-bridge ## Runs all tests checking race conditions
	sleep 3
	trap '$(STOP)' EXIT; MallocNanoZone=0 go test -v -failfast -race -p 1 -timeout 2400s ./test/e2e/... -count 1 -tags='autoclaiml2l2'

.PHONY: test-e2ecompress
test-e2ecompress: build-docker stop run-multi-single-bridge ## Runs all tests checking race conditions
	sleep 3
	trap '$(STOP)' EXIT; MallocNanoZone=0 go test -v -failfast -race -p 1 -timeout 2400s ./test/e2e/... -count 1 -tags='e2ecompress'

.PHONY: test-sovereignchain
test-sovereignchain: build-docker stop run-sovereign-chain ## Runs all tests checking race conditions
	sleep 3
	trap '$(STOP)' EXIT; MallocNanoZone=0 go test -v -failfast -race -p 1 -timeout 2400s ./test/e2e/... -count 1 -tags='sovereignchain'

.PHONY: build-test-e2e-real_network
build-test-e2e-real_network: ## Build binary for e2e tests with real network
	go test -c ./test/e2e/ -o dist/zkevm-bridge-e2e-real_network-erc20 -tags='e2e_real_network_erc20'
	go test -c ./test/e2e/ -o dist/zkevm-bridge-e2e-real_network-bridgemsg -tags='e2e_real_network_msg'
	./dist/zkevm-bridge-e2e-real_network-erc20 -test.failfast -test.list Test
	./dist/zkevm-bridge-e2e-real_network-bridgemsg -test.failfast -test.list Test

.PHONY: validate
validate: lint build test-full ## Validates the whole integrity of the code base

## Help display.
## Pulls comments from beside commands and prints a nicely formatted
## display with the commands and their usage information.
.DEFAULT_GOAL := help

.PHONY: help
help: ## Prints this help
		@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

COMMON_MOCKERY_PARAMS=--disable-version-string --with-expecter
.PHONY: generate-mocks
generate-mocks: ## Generates mocks for the tests, using mockery tool
	mockery --name=ethermanInterface --dir=synchronizer --output=synchronizer --outpkg=synchronizer --structname=ethermanMock --filename=mock_etherman.go ${COMMON_MOCKERY_PARAMS}
	mockery --name=storageInterface --dir=synchronizer --output=synchronizer --outpkg=synchronizer --structname=storageMock --filename=mock_storage.go ${COMMON_MOCKERY_PARAMS}
	mockery --name=bridgectrlInterface --dir=synchronizer --output=synchronizer --outpkg=synchronizer --structname=bridgectrlMock --filename=mock_bridgectrl.go ${COMMON_MOCKERY_PARAMS}
	mockery --name=Tx --srcpkg=github.com/jackc/pgx/v4 --output=synchronizer --outpkg=synchronizer --structname=dbTxMock --filename=mock_dbtx.go ${COMMON_MOCKERY_PARAMS}
	mockery --name=bridgeServiceStorage --dir=server --output=server --outpkg=server --structname=bridgeServiceStorageMock --filename=mock_bridgeServiceStorage.go ${COMMON_MOCKERY_PARAMS}
	
	rm -Rf claimtxman/mocks
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --all --case snake --dir claimtxman/ --output claimtxman/mocks --outpkg mock_txcompressor ${COMMON_MOCKERY_PARAMS}
	

.PHONY: generate-smartcontracts-bindings
generate-smartcontracts-bindings:	## Generates the smart contracts bindings
	cd scripts && ./generate-smartcontracts-bindings.sh
	