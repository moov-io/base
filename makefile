.PHONY: setup
setup: gen-certs
	docker compose up -d --force-recreate --remove-orphans

.PHONY: check
check:
ifeq ($(OS),Windows_NT)
	go test ./... -short
else
	@wget -O lint-project.sh https://raw.githubusercontent.com/moov-io/infra/master/go/lint-project.sh
	@chmod +x ./lint-project.sh
	GOCYCLO_LIMIT=26 COVER_THRESHOLD=50.0 GOLANGCI_LINTERS=gosec ./lint-project.sh
endif

.PHONY: clean
clean:
	@rm -rf ./bin/ ./tmp/ coverage.txt misspell* staticcheck lint-project.sh

.PHONY: cover-test cover-web
cover-test:
	go test -coverprofile=cover.out ./...
cover-web:
	go tool cover -html=cover.out

.PHONY: teardown
teardown:
	-docker compose down --remove-orphans
	-docker compose rm -f -v

.PHONY: gen-certs
gen-certs:
	./database/testdata/gencerts.sh
