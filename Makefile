# Include variables from the .envrc file
include .envrc
#I should also point out that positioning the help rule as the first thing in the Makefile is a deliberate move. If you run make without specifying a target then it will default to executing the first rule in the file.
## help: print this help message
# ==================================================================================== #
# HELPERS
# ==================================================================================== #
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
# Create the new confirm target.
.PHONY:confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #
#A phony target is one that is not really the name of a file; rather it is just a name for a rule to be executed.
## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	@go run ./cmd/api -db-dsn=${GREENLIGHT_DB_DSN}

## db/psql: connect to the database using psql
.PHONY:db/psql
db/psql:
	psql ${GREENLIGHT_DB_DSN}

## db/migrations/new name=$1: create a new database migration it has confirm target
.PHONY:db/migration/new
db/migration/new: confirm
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply all up database migrations  it has confirm target
.PHONY:db/migration/up
db/migration/up: confirm
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${GREENLIGHT_DB_DSN} up

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #
## audit: tidy dependencies and format, vet and test all code
.PHONY: audit
audit:vendor
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

# ==================================================================================== #
# BUILD
# ==================================================================================== #

## build/api: build the cmd/api application
#It’s possible to reduce the binary size by around 25% by instructing the Go linker to strip symbol tables and DWARF debugging information from the binary. We can do this as part of the go build command by using the linker flag -ldflags="-s" as follows
.PHONY: build/api
build/api:
	@echo 'Building cmd/api...'
	go build -ldflags='-s' -o=./bin/api ./cmd/api
	GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=./bin/linux_amd64/api ./cmd/api

# ==================================================================================== #
# PRODUCTION
# ==================================================================================== #

production_host_ip = '207.154.235.180'

## production/connect: connect to the production server
.PHONY: production/connect
production/connect:
	ssh -i ${sshKey} greenlight@${production_host_ip}

## production/deploy/api: deploy the api to production
#.PHONY: production/deploy/api
#production/deploy/api:
#	rsync -P -e 'ssh -i ${sshKey}' ./bin/linux_amd64/api greenlight@${production_host_ip}:~
#	rsync -rP -e 'ssh -i ${sshKey}' --delete ./migrations greenlight@${production_host_ip}:~
#	ssh -i ${sshKey} -t greenlight@${production_host_ip} 'migrate -path ~/migrations -database $$GREENLIGHT_DB_DSN up'

## production/deploy/api: deploy the api to production
.PHONY: production/deploy/api
production/deploy/api:
	rsync -P -e 'ssh -i ${sshKey}' ./bin/linux_amd64/api greenlight@${production_host_ip}:~
	rsync -rP -e 'ssh -i ${sshKey}' --delete ./migrations greenlight@${production_host_ip}:~
	rsync -P -e 'ssh -i ${sshKey}' ./remote/production/api.service greenlight@${production_host_ip}:~
	rsync -P -e 'ssh -i ${sshKey}' ./remote/production/Caddyfile greenlight@${production_host_ip}:~
	ssh -i ${sshKey} -t greenlight@${production_host_ip} '\
		migrate -path ~/migrations -database $$GREENLIGHT_DB_DSN up \
		&& sudo mv ~/api.service /etc/systemd/system/ \
		&& sudo systemctl enable api \
		&& sudo systemctl restart api \
		&& sudo mv ~/Caddyfile /etc/caddy/ \
        && sudo systemctl reload caddy \
		'