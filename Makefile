run:
	go run ./cmd/api

psql:
	psql postgres://greenlight:${DB_PW}@localhost/greenlight?sslmode=disable

up:
	@echo 'Running up migrations...'
	migrate -path=./migrations -database postgres://greenlight:${DB_PW}@localhost/greenlight?sslmode=disable up
