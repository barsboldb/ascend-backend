include .env
export

.PHONY: gen run sqlc migrate-up migrate-down

gen:
	protoc \
		--go_out=gen \
		--go_opt=paths=source_relative \
		--go-grpc_out=gen \
		--go-grpc_opt=paths=source_relative \
		-I proto \
		$(shell find proto -name "*.proto")

run:
	go run cmd/server/main.go

sqlc:
	sqlc generate

migrate-up:
	migrate -path db/migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path db/migrations -database "$(DATABASE_URL)" down

seed:
	go run cmd/seed/main.go
