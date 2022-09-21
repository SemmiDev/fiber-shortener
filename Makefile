run:
	go run main.go
sqlc:
	sqlc generate
up:
	docker-compose -f docker-compose.dev.yml up -d
stop:
	docker-compose -f docker-compose.dev.yml stop
down:
	docker-compose -f docker-compose.dev.yml down

.PHONY: sqlc up stop down