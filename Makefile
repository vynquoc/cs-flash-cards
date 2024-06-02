include .envrc

.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'


confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]


## run/api: run the cmd/api application
run/api:
	go run ./cmd/api -db-dsn=${CSFLASHCARDS_DB_DSN}

## db/psql: connect to the database using psql
db/psql:
	psql ${CSFLASHCARDS_DB_DSN}

## db/migrations/new name=$1: create a new database migration
db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply all up database migrations
db/migrations/up: confirm
	@echo "Running up migrations..."
	migrate -path ./migrations -database ${CSFLASHCARDS_DB_DSN} up

## audit: tidy dependencies and format, vet and test all code
.PHONY: audit 
audit:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...


## vendor: tidy and vendor dependencies
.PHONY: vendor 
vendor:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Vendoring dependencies...'
	go mod vendor
.PHONY: build/api
build/api:
	@echo 'Building cmd/api...'
	go build -ldflags='-s' -o=./bin/api ./cmd/api
	GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=./bin/linux_amd64/api ./cmd/api


# ==================================================================================== # # PRODUCTION
# ==================================================================================== #
production_host_ip = '52.62.76.64'
## production/connect: connect to the production server
.PHONY: production/connect
production/connect:
	ssh csflashcards@${production_host_ip}

## production/deploy/api: deploy the api to production
.PHONY: production/deploy/api
production/deploy/api:
	rsync -P ./bin/linux_amd64/api csflashcards@${production_host_ip}:~
	rsync -rP --delete ./migrations csflashcards@${production_host_ip}:~
	rsync -P ./remote/production/api.service csflashcards@${production_host_ip}:~
	rsync -P ./remote/production/Caddyfile csflashcards@${production_host_ip}:~
	ssh -t csflashcards@${production_host_ip} '\
	migrate -path ~/migrations -database $$CSFLASHCARDS_DB_DSN up \
	&& sudo mv ~/api.service /etc/systemd/system/ \
	&& sudo systemctl enable api \
	&& sudo systemctl restart api \
	&& sudo mv ~/Caddyfile /etc/caddy/ \
	&& sudo systemctl reload caddy \
	'
