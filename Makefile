run/api:
	go run ./cmd/api

confirm:
	@echo 'Are you sure? [y/N]' && read ans && [ $${ans:-N} = y ]

db/psql:
	psql "postgres://greenlight:${DB_PW}@localhost/greenlight?sslmode=disable"

db/migrations/new:
	@echo 'Creating migration files for ${name}'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

db/migrations/up: confirm
	@echo 'Running up migrations...'
	migrate -path="./migrations" -database "postgres://greenlight:${DB_PW}@localhost/greenlight?sslmode=disable" up
