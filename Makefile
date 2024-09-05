# Include go binaries into path
export PATH := $(GOPATH)/bin:$(PATH)

CURRENT_BRANCH_NAME= $(shell git rev-parse --abbrev-ref HEAD)
DATA_PATH := $(shell pwd)/data
EXE_PATH := $(shell pwd)/cmd/exe
BIN := $(CURDIR)/bin/
SOURCE_PATH := GOBIN=$(BIN) DATA_PATH=$(DATA_PATH) CURDIR=$(shell pwd) CURRENT_BRANCH_NAME=$(CURRENT_BRANCH_NAME)

install: mod

mod-action-%:
	@echo "Run go mod ${*}...."
	GOBIN=$(BIN) GO111MODULE=on go mod $*
	@echo "Done go mod  ${*}"

mod: mod-action-verify mod-action-tidy mod-action-vendor mod-action-download mod-action-verify ## Download all dependencies

tests: clean-cache-test ## run all tests
	$(SOURCE_PATH) go test ./... -race -v -coverprofile coverage.out
	go tool cover -html=coverage.out -o coverage.html
	rm coverage.out

clean-cache-test: ## clean cache
	@echo "Cleaning test cache..."
	$(SOURCE_PATH) go clean -testcache

clean: clean-cache-test ## clean cache
	@echo "Cleaning..."
	rm -fr ./vendor
	go clean -i -r -x -cache -testcache -modcache

build: ## build app
	mkdir -p $(DATA_PATH)
ifdef version
ifneq ("$(wildcard $(DATA_PATH)/app.$(version))","")
	@echo "Version alredy exists version..."
	exit 1
else
	@echo "Build app $(DATA_PATH)/app.$(version)..."
	$(SOURCE_PATH) go build -ldflags "-w -s -X main.Version=${version}" -o $(DATA_PATH)/app.$(version) ./main.go
	chmod 755 $(DATA_PATH)/app.$(version)
endif
else
	@echo "Empty version..."
	exit 1
endif

run: build ## build & run app
	mkdir -p  $(EXE_PATH)/
	cp $(DATA_PATH)/app.$(version) $(EXE_PATH)/app
	mkdir -p $(DATA_PATH)/logs
	$(EXE_PATH)/app

gen-key: ## run helper to generate key
	go run ./cmd/gen-key/main.go

run-images-server: ## run server with images
	go run ./cmd/server/main.go