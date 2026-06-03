.PHONY: infra-up infra-down infra-logs infra-reset \
        run-api run-gateway run-fanout run-persistence run-ttl dev-up dev-down \
        demo-migrate demo-build demo-backend demo-proxy demo-restart-api \
        demo-set-minio demo-up demo-tunnel demo-down demo-logs

COMPOSE_FILE=./deploy/compose.yml
DEMO_COMPOSE=./deploy/compose.demo.yml

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

demo-migrate:
	bash deploy/scripts/migrate.sh

demo-build:
	bash deploy/scripts/build-frontend.sh

demo-backend:
	bash deploy/scripts/start-backend.sh

demo-proxy:
	bash deploy/scripts/demo-proxy.sh

demo-restart-api:
	bash deploy/scripts/restart-api.sh

demo-stop-api:
	bash deploy/scripts/kill-port.sh 8081
	@pkill -f "services/api/cmd" 2>/dev/null || true

# Ex.: make demo-set-minio URL=https://abc.trycloudflare.com/storage
demo-set-minio:
	@if [ -z "$(URL)" ]; then echo "Uso: make demo-set-minio URL=https://host.trycloudflare.com/storage"; exit 1; fi
	bash deploy/scripts/set-minio-public-url.sh "$(URL)"

demo-up: infra-up demo-migrate demo-build demo-backend demo-proxy
	@echo ""
	@echo "Demo local: http://localhost:3000"
	@echo "Link público: make demo-tunnel (outro terminal)"
	@echo "Guia: deploy/DEMO_DEPLOY.md"

demo-tunnel:
	bash deploy/scripts/cloudflare-tunnel.sh

demo-down:
	bash deploy/scripts/stop-backend.sh
	docker compose -f $(DEMO_COMPOSE) down 2>/dev/null || true
	$(MAKE) infra-down

demo-logs:
	@tail -n 30 deploy/logs/*.log 2>/dev/null || echo "Sem logs em deploy/logs/"
