infra-up:
	docker compose -f ./deploy/compose.yml up -d

infra-down:
	docker compose -f ./deploy/compose.yml down

infra-logs:
	docker compose -f ./deploy/compose.yml logs -f

run-api:
	go run ./services/api/cmd

run-gateway:
	go run ./services/gateway/cmd

run-fanout:
	go run ./services/workers/fanout/cmd
