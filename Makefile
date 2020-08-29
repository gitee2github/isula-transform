NAME=isula-transform

VERSION=0.9.0
COMMIT=$(shell git rev-parse HEAD 2> /dev/null || true)

BEP_DIR=/tmp/isula-transform-build-bep
BEP_FLAGS=-tmpdir=$(BEP_DIR)

TAGS="cgo static_build"
LDFLAGS="-s -w -buildid=IdByiSula -buildmode=pie -extldflags=-zrelro -extldflags=-znow $(BEP_FLAGS) -X main.version=$(VERSION) -X main.gitCommit=${COMMIT}"
ENV=CGO_ENABLED=1
GOMOD_ENV=GO111MODULE=on

.PHONY: all
all: localbuild

.PHONY: bep
bep:
	@mkdir -p $(BEP_DIR)

.PHONY: localbuild
localbuild: bep
	@go mod vendor
	$(ENV) $(GOMOD_ENV) go build -tags $(TAGS) -ldflags $(LDFLAGS) -mod=vendor -o bin/$(NAME) .
	@rm -rf $(BEP_DIR)

.PHONY: bin
bin: bep
	$(ENV) go build -tags $(TAGS) -ldflags $(LDFLAGS) -o bin/$(NAME) .
	@rm -rf $(BEP_DIR)

.PHONY: install
install:
	install -m 0750 ./bin/isula-transform  /usr/bin/isula-transform

.PHONY: binclean
binclean:
	@echo "clean built binary file"
	@rm -rf ./bin/$(NAME)

.PHONY: mock
mock:
	@echo "waiting for generate test mock file"
	@go generate isula.org/isula-transform/...

.PHONY: mockclean
mockclean:
	@echo "clean test mock file"
	@find . -name "mock_*.go" -delete

.PHONY: runtest
runtest: mock
	@echo "run test:"
	go test -cover -timeout 30s isula.org/isula-transform/...

.PHONY: test
test: runtest mockclean

.PHONY: clean
clean:  binclean mockclean
