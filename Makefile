BINARY=svcomp
VERSION?=dev
GO=go
GOTOOLCHAIN?=local
DOCKER_COMPOSE?=docker compose
INTEGRATION_SCHEMA_DIR?=testdata/integration/schemas
INTEGRATION_SOURCE_DSN?=root:root@tcp(127.0.0.1:3307)/svcomp_source?parseTime=true
INTEGRATION_TARGET_DSN?=root:root@tcp(127.0.0.1:3308)/svcomp_target?parseTime=true

.PHONY: build build-all test test-integration integration-up integration-down clean-integration integration-reset integration-restore

LDFLAGS=-ldflags "-X main.version=$(VERSION)"

build:
	$(GO) build $(LDFLAGS) -o bin/$(BINARY) ./cmd/svcomp

build-all:
	mkdir -p dist
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o dist/$(BINARY)-$(VERSION)-linux-amd64 ./cmd/svcomp
	GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) -o dist/$(BINARY)-$(VERSION)-linux-arm64 ./cmd/svcomp
	GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o dist/$(BINARY)-$(VERSION)-darwin-amd64 ./cmd/svcomp
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o dist/$(BINARY)-$(VERSION)-windows-amd64.exe ./cmd/svcomp

test:
	GOTOOLCHAIN=$(GOTOOLCHAIN) $(GO) clean -cache -testcache
	GOTOOLCHAIN=$(GOTOOLCHAIN) $(GO) test ./...

integration-up:
	$(DOCKER_COMPOSE) up -d --wait

integration-down:
	$(DOCKER_COMPOSE) down

clean-integration:
	$(DOCKER_COMPOSE) down -v --remove-orphans

integration-reset: integration-up
	$(DOCKER_COMPOSE) exec -T mysql-source mysql -uroot -proot -e "DROP DATABASE IF EXISTS svcomp_source;"
	$(DOCKER_COMPOSE) exec -T mysql-target mysql -uroot -proot -e "DROP DATABASE IF EXISTS svcomp_target;"

integration-restore: integration-reset
	$(DOCKER_COMPOSE) exec -T mysql-source mysql -uroot -proot < $(INTEGRATION_SCHEMA_DIR)/schema1.sql
	$(DOCKER_COMPOSE) exec -T mysql-target mysql -uroot -proot < $(INTEGRATION_SCHEMA_DIR)/schema2.sql

test-integration: integration-restore
	GOTOOLCHAIN=$(GOTOOLCHAIN) $(GO) clean -cache -testcache
	INTEGRATION_SOURCE_DSN="$(INTEGRATION_SOURCE_DSN)" INTEGRATION_TARGET_DSN="$(INTEGRATION_TARGET_DSN)" GOTOOLCHAIN=$(GOTOOLCHAIN) $(GO) test -tags=integration ./...
