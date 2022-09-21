run:
	go run main.go
sqlc:
	sqlc generate
up:
	docker-compose up -d
stop:
	docker-compose stop
down:
	docker-compose down
up-dev:
	docker-compose -f docker-compose.dev.yml up -d
stop-dev:
	docker-compose -f docker-compose.dev.yml stop
down-dev:
	docker-compose -f docker-compose.dev.yml down

.PHONY: sqlc up stop up-dev stop-dev down down-dev