PKG := github.com/fewsats/blockbuster

GO_BIN := ${GOPATH}/bin
MIGRATE_BIN := $(GO_BIN)/migrate

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

DOCKER_IMAGE_TAG ?= dev

build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o blockbuster ./cmd/server
test:
	go test ./...


# ======
# DOCKER 
# ======
docker-build:
	docker build -t blockbuster:$(DOCKER_IMAGE_TAG) .

# ===================
# DATABASE MIGRATIONS
# ===================
migrate-up: $(MIGRATE_BIN)
	migrate -path store/sqlc/migrations -database $(BLOCKBUSTER_DB_CONNECTIONSTRING) -verbose up

migrate-down: $(MIGRATE_BIN)
	migrate -path store/sqlc/migrations -database $(BLOCKBUSTER_DB_CONNECTIONSTRING) -verbose down 1

migrate-create: $(MIGRATE_BIN)
	migrate create -dir store/sqlc/migrations -seq -ext sql $(patchname)


# ===============
# CODE GENERATION 
# ===============
gen: sqlc

sqlc:
	@$(call print, "Generating sql models and queries in Go")
	./scripts/gen_sqlc_docker.sh

sqlc-check: sqlc
	@$(call print, "Verifying sql code generation.")
	@if test -n "$$(git status --porcelain '*.go')"; then \
		echo "SQL models not properly generated! Modified changes:"; \
		git status --porcelain '*.go'; \
		exit 1; \
	else \
		echo "SQL models generated correctly."; \
	fi
