# SlickChat

SlickChat é uma plataforma de chat em tempo real baseada em arquitetura orientada a eventos (event-driven), com foco em anonimato, desacoplamento e escalabilidade.

---

# Arquitetura

O sistema segue uma combinação de:

* **Event-Driven Architecture (EDA)**
* **CQRS (Command Query Responsibility Segregation)**
* **Clean Architecture (parcial)**

---

## Fluxo principal

```
WebSocket (Gateway)
        ↓
      Kafka
        ↓
 ┌───────────────┬───────────────┐
 │               │               │
Fanout Worker   Persistence     (futuro: analytics, moderação)
(WebSocket)     Worker
        ↓               ↓
     Redis          Postgres
        ↓
   WebSocket Client
```

---

## Como pensar o sistema

* **Gateway não executa lógica** → apenas emite eventos
* **Workers reagem aos eventos**
* **Kafka é o backbone do sistema**
* **Redis é usado para presença e routing**
* **Postgres é a fonte de verdade**

---

# Estrutura de pastas

```
core/
├── domain/        # Entidades do domínio (Message, etc)
├── contracts/     # Interfaces (ex: MessageRepository)
├── events/        # Definição dos eventos do sistema
├── application/   # Casos de uso (SendMessage, etc)

infrastructure/
├── kafka/         # Producer + Consumer
├── postgres/      # Implementação do repository
├── redis/         # Integração com Redis
├── log/           # Logger (slog)

services/
├── api/           # Read model (HTTP)
├── gateway/       # WebSocket (entrada de comandos)
└── workers/
    ├── fanout/        # Entrega de mensagens (Redis → WS)
    └── persistence/   # Persistência no banco

deploy/
└── compose.yml    # Infra (Kafka, Postgres, Redis)
```

---

## Separação de responsabilidades

| Camada         | Responsabilidade                  |
| -------------- | --------------------------------- |
| core           | Regras de negócio puras           |
| infrastructure | Integração com serviços externos  |
| services       | Orquestração / entrada do sistema |

---

# Funcionalidades implementadas

## 🔹 Mensagens em tempo real

* WebSocket
* Kafka
* Fanout Worker

## 🔹 Persistência

* Worker dedicado
* Postgres
* TTL (modelo)
* destroy_after_read (modelo)

## 🔹 API (Read Model)

```
GET /messages?room_id={UUID}
```

* Retorna histórico
* Baseado diretamente no banco

---

# Como rodar o projeto

## 1. Subir infraestrutura

```bash
make infra-up
```

---

## 2. Rodar serviços

### API

```bash
make run-api
```

### Gateway

```bash
make run-gateway
```

### Fanout

```bash
make run-fanout
```

### Persistence

```bash
make run-persistence
```

---

# 🧪 Testando o sistema

## 1. Conectar via WebSocket

```bash
wscat -c "ws://localhost:8080/socket?user_id=<UUID>"
```

## 2. Adicionar usuário na sala

```bash
make redis-cli
SADD room_members:<ROOM_ID> <USER_ID>
```

## 3. Enviar mensagem

```json
{
  "type": "send_message",
  "payload": {
    "room_id": "<ROOM_ID>",
    "content": "hello world"
  }
}
```

---

## 4. Verificar persistência

```bash
make psql
SELECT * FROM messages;
```

---

## 5. Buscar histórico

```bash
curl "http://localhost:8081/messages?room_id=<ROOM_ID>"
```

---

# 🛠 Comandos úteis

```bash
make infra-down
make infra-logs
make infra-reset
make psql
make redis-cli
```

---

# Conceitos aplicados

* Event-driven architecture
* CQRS
* Desacoplamento de serviços
* Processamento assíncrono
* Pub/Sub com Redis

---
