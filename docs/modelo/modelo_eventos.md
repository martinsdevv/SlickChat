# SlickChat — Modelo de Eventos

## 1. Introdução

Este documento descreve o modelo de eventos utilizado no sistema SlickChat.

O sistema utiliza uma arquitetura **event-driven**, na qual ações realizadas pelos usuários geram eventos que são publicados em um broker de mensageria (Kafka).

Esses eventos são consumidos por diferentes componentes do sistema para realizar processamento assíncrono.

---

# 2. Estrutura Base de Evento

Todos os eventos seguem uma estrutura padrão.

```
event_id
event_type
event_version
timestamp
payload
```

Exemplo:

```json
{
  "event_id": "uuid",
  "event_type": "string",
  "event_version": 1,
  "timestamp": "ISO8601",
  "partition_key": "room_id", 
  "payload": {}
}
```

**Nota**: A partition_key deve ser o room_id para garantir a ordenação correta das mensagens dentro de uma sala no Kafka.

---

# 3. Categorias de Eventos

Os eventos do sistema são divididos em cinco categorias principais.

```
User Events: Gestão de conta.
Presence Events: Status online/offline via Redis.
Room Events: Ciclo de vida de salas.
Message Events: Comunicação e efemeridade.
Moderation Events: Segurança e denúncias.
```

---

# 4. User Events

Eventos relacionados a usuários.

## UserCreated

Gerado quando uma nova conta é criada.

Payload:

```
user_id
username
discriminator
created_at
```

---

## UserAuthenticated

Gerado quando um usuário realiza login.

Payload:

```
user_id
session_id
timestamp
```

---

## UserIdentityRotated

Gerado quando um usuário rotaciona seu identificador `#xxxx`.

Payload:

```
user_id
old_discriminator
new_discriminator
timestamp
```

---

# 5. Presence Events

Estes eventos geram o ciclo de vida das ligações e o estado visível do utilizador, utilizando o Redis para armazenamento de estado de baixa latência.

## UserConnected

Gerado quando o WebSocket Gateway estabelece uma nova ligação ativa com o cliente.

Payload:

```
user_id
connection_id
timestamp
```

## UserDisconnected
Gerado quando uma ligação WebSocket é encerrada, seja por iniciativa do cliente ou falha de rede.

Payload:

```
user_id: UUID do utilizador.
connection_id: ID da ligação que foi encerrada.
reason: Motivo do encerramento (ex: `client_closed`, `ping_timeout`, `kicked`).
timestamp: Momento da desconexão.
```

## UserPresenceChanged
Gerado quando o estado global de presença do utilizador muda (ex: passa de offline para online ao abrir a primeira conexão).

Payload:

```
user_id: UUID do utilizador.
status: Novo estado visível (online, offline, invisible).
timestamp: Momento da alteração de estado.
```
# 6. Session Events

Eventos relacionados a sessões de autenticação.

## UserSessionCreated

Gerado quando uma nova sessão é criada.

Payload:

```
session_id
user_id
created_at
expires_at
```

---

## UserSessionExpired

Gerado quando uma sessão expira.

Payload:

```
session_id
user_id
expired_at
```

---

# 7. Room Events

Eventos relacionados a salas.

## RoomCreated

Gerado quando uma sala é criada.

Payload:

```
room_id
owner_id
type
created_at
```

---

## RoomExpired

Gerado quando uma sala temporária expira.

Payload:

```
room_id
expired_at
```

---

## UserJoinedRoom

Gerado quando um usuário entra em uma sala.

Payload:

```
room_id
user_id
joined_at
```

---

## UserLeftRoom

Gerado quando um usuário sai de uma sala.

Payload:

```
room_id
user_id
left_at
```

---

# 8. Message Events

Eventos relacionados a mensagens.

## MessageSent

Gerado quando uma mensagem é enviada.

Payload:

```
message_id
room_id
sender_id
message_type
content
is_zero_logging
ttl
expires_at
timestamp
```

---

## MessageDeleted

Gerado quando uma mensagem é removida.

Payload:

```
message_id
room_id
deleted_at
```

---

## MessageExpired

Gerado quando uma mensagem com TTL expira.

Payload:

```
message_id
room_id
expired_at
```

---

## AttachmentUploaded

Gerado quando um anexo é enviado em uma mensagem.

Payload:

```
attachment_id
message_id
type
size
storage_url
```

---

# 9. Moderation Events

Eventos relacionados a moderação.

## UserMuted

Gerado quando um usuário é silenciado.

Payload:

```
room_id
moderator_id
target_user_id
duration
```

---

## UserKicked

Gerado quando um usuário é removido da sala.

Payload:

```
room_id
moderator_id
target_user_id
```

---

## UserBanned

Gerado quando um usuário é banido.

Payload:

```
room_id
moderator_id
target_user_id
```

---

## ReportCreated

Gerado quando um usuário cria uma denúncia.

Payload:

```
report_id
reporter_id
target_user_id
message_id
reason
created_at
```

---

# 10. Fluxo de Eventos

## 10.1 Fluxo Padrão (Persistente)
1. **User → Gateway**: Via WebSocket.
2. **Gateway → Kafka**: Publica `MessageSent`.
3. **Kafka → Fanout Worker**: Entrega em tempo real.
4. **Kafka → Persistence Worker**: Salva no Postgres.

## 10.2 Fluxo Zero Logging (Stream Only)
1. **Gateway → Kafka**: Publica com `is_zero_logging: true`.
2. **Fanout Worker**: Entrega em tempo real normalmente.
3. **Persistence Worker**: Identifica a flag `true` e descarta o evento sem salvar no banco.

---

# 11. Versionamento de Eventos

Eventos possuem um campo `event_version`.

Isso permite evolução do sistema sem quebrar consumidores existentes.

---

# 12. Benefícios da Arquitetura de Eventos

A arquitetura baseada em eventos oferece:

* desacoplamento entre serviços
* escalabilidade horizontal
* processamento assíncrono
* maior resiliência do sistema
