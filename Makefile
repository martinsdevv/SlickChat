.PHONY: infra-up infra-down infra-logs infra-reset \
        run-api run-gateway run-fanout run-persistence run-ttl dev-up dev-down

COMPOSE_FILE=./deploy/compose.yml

infra-up:
	docker compose -f $(COMPOSE_FILE) up -d

infra-down:
	docker compose -f $(COMPOSE_FILE) down

infra-logs:
	docker compose -f $(COMPOSE_FILE) logs -f

infra-reset:
	docker compose -f $(COMPOSE_FILE) down -v

run-api:
	go run ./services/api/cmd

run-gateway:
	go run ./services/gateway/cmd

run-fanout:
	go run ./services/workers/fanout/cmd

run-persistence:
	go run ./services/workers/persistence/cmd

run-ttl:
	go run ./services/workers/ttl/cmd

psql:
	docker exec -it slickchat-postgres psql -U postgres -d slickchat

redis-cli:
	docker exec -it slickchat-redis redis-cli

dev-up: infra-up
	kitty @ launch --type=window --cwd $(PWD) --title "persistence" make run-persistence
	kitty @ launch --type=window --cwd $(PWD) --title "fanout" make run-fanout
	kitty @ launch --type=window --cwd $(PWD) --title "ttl" make run-ttl
	kitty @ launch --type=window --cwd $(PWD) --title "api" make run-api
	kitty @ launch --type=window --cwd $(PWD) --title "gateway" make run-gateway

dev-down:
	-@kitty @ close-window --match title:persistence
	-@kitty @ close-window --match title:fanout
	-@kitty @ close-window --match title:ttl
	-@kitty @ close-window --match title:api
	-@kitty @ close-window --match title:gateway
	$(MAKE) infra-down
