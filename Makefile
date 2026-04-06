.PHONY: infra-up infra-down infra-logs infra-reset \
        run-api run-gateway run-fanout

COMPOSE_FILE=./deploy/compose.yml

infra-up:
	docker compose -f $(COMPOSE_FILE) up -d

infra-down:
	docker compose -f $(COMPOSE_FILE) down

infra-logs:
	docker compose -f $(COMPOSE_FILE) logs -f

infra-reset:
	docker compose -f $(COMPOSE_FILE) down -v

# 🚀 serviços
run-api:
	go run ./services/api/cmd

run-gateway:
	go run ./services/gateway/cmd

run-fanout:
	go run ./services/workers/fanout/cmd

psql:
	docker exec -it slickchat-postgres psql -U postgres -d slickchat

redis-cli:
	docker exec -it slickchat-redis redis-cli
