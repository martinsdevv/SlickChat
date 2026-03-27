# SlickChat — Diagramas de Design

Este documento contém os principais diagramas de design do sistema **SlickChat** baseados no modelo de domínio definido em modelo_dominio.

Os diagramas são fornecidos em **Mermaid** para fácil visualização e também podem ser importados em diversas ferramentas.

---

# 1. Diagrama de Classes

```mermaid
classDiagram

class User {
  UUID id
  string username
  string discriminator
  string password_hash
  bool paranoid_mode
  datetime created_at
}

class IdentityKey {
  UUID id
  UUID user_id
  string key_hash
  datetime created_at
}

class Session {
  UUID id
  UUID user_id
  datetime created_at
  datetime expires_at
  string ip_hash
}

class Room {
  UUID id
  string name
  string type
  UUID owner_id
  int ttl
  bool paranoid_mode
  bool zero_logging
  datetime created_at
  datetime expires_at
}

class RoomMembership {
  UUID id
  UUID room_id
  UUID user_id
  string role
  datetime joined_at
}

class Message {
  UUID id
  UUID room_id
  UUID sender_id
  string content
  string message_type
  int ttl
  bool destroy_after_read
  datetime created_at
  datetime expires_at
}

class Attachment {
  UUID id
  UUID message_id
  string storage_url
  string type
  int size
  datetime created_at
}

class ModerationAction {
  UUID id
  UUID room_id
  UUID moderator_id
  UUID target_user_id
  string action
  datetime created_at
}

class Report {
  UUID id
  UUID reporter_id
  UUID target_user_id
  UUID message_id
  string reason
  datetime created_at
}

User "1" --> "0..1" IdentityKey
User "1" --> "0..*" Session
User "1" --> "0..*" RoomMembership

Room "1" --> "0..*" RoomMembership
Room "1" --> "0..*" Message
Room "1" --> "0..*" ModerationAction

Message "1" --> "0..*" Attachment

Report "*" --> "1" User : reporter
Report "*" --> "1" User : target
```

---

# 2. C4 Model — System Context

```mermaid
flowchart TB

User((User))
Admin((Room Admin))

SlickChat["SlickChat System"]

MediaStorage[(Media Storage)]

User -->|Send & receive messages| SlickChat
Admin -->|Moderate rooms| SlickChat

SlickChat -->|Store attachments| MediaStorage
```

Descrição:

* Usuários interagem com o sistema SlickChat
* Administradores de sala realizam moderação
* O sistema utiliza armazenamento para mídia e banco de dados para persistência

---

# 3. C4 Model — Container Diagram

```mermaid
flowchart TD

Client[Web Client]

Gateway[WebSocket]
API[API Service]
Kafka[(Kafka Event Stream)]
Redis[(Redis)]

FanoutWorker[Message Fanout Worker]
PersistenceWorker[Persistence Worker]
TTLWorker[TTL Worker]
ModerationWorker[Moderation Worker]

Postgres[(Postgres Database)]
Media[(Media Storage)]

Client --> Gateway
Client --> API

Gateway --> Kafka
Gateway --> Redis

Kafka --> FanoutWorker
Kafka --> PersistenceWorker
Kafka --> TTLWorker
Kafka --> ModerationWorker

PersistenceWorker --> Postgres
FanoutWorker --> Gateway

API --> Postgres
API --> Media
```

Descrição:

* Clientes se conectam via **WebSocket Gateway** para comunicação em tempo real

* Eventos de chat são publicados em **Kafka**

* Workers consomem eventos para:

  * distribuir mensagens
  * persistir dados
  * aplicar TTL
  * executar moderação

* **Redis** é utilizado para presença, sessões e rate limiting

* Dados persistentes são armazenados em **Postgres**

* Arquivos e anexos são armazenados em **Media Storage**

---

# 4. Fluxo de Mensagem

```mermaid
sequenceDiagram

participant U as User
participant GW as Realtime Gateway
participant K as Kafka
participant W as Workers
participant DB as Database

U->>GW: Send Message
GW->>K: Publish MessageSent

K->>W: Deliver Event
W->>GW: Fanout to Clients
W->>DB: Persist Message
```

---

# 5. Fluxo de Expiração de Mensagens

```mermaid
sequenceDiagram

participant Worker
participant DB

Worker->>DB: Scan expired messages
Worker->>DB: DELETE expired messages
```

Mensagens com TTL são removidas fisicamente do banco de dados.
