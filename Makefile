run:
	go run ./cmd/api

psql:
	psql "postgres://greenlight:${DB_PW}@localhost/greenlight?sslmode=disable"

migration:
	@echo 'Creating migration files for ${name}'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

up:
	@echo 'Running up migrations...'
	migrate -path="./migrations" -database "postgres://greenlight:${DB_PW}@localhost/greenlight?sslmode=disable" up
