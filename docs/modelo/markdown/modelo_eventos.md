# SlickChat — Modelo de Eventos

## 1. Introdução

Este documento descreve o modelo de eventos utilizado no sistema SlickChat.

O sistema utiliza uma arquitetura **event-driven**, na qual ações realizadas pelos usuários geram eventos que são publicados em um broker de mensageria (Kafka).

Esses eventos são consumidos por diferentes componentes do sistema para realizar processamento assíncrono.

---

# 2. Estrutura Base de Evento

Todos os eventos seguem uma estrutura padrão.

| Campo         | Tipo               | Descrição                                      |
|---------------|--------------------|------------------------------------------------|
| event_id      | UUID               | Identificador único do evento.                 |
| event_type    | string             | Tipo do evento (ex: `MessageSent`).            |
| event_version | int                | Versão do esquema do evento.                   |
| timestamp     | datetime (ISO8601) | Momento em que o evento foi gerado.            |
| partition_key | string             | Chave de particionamento (ex: `room_id`).      |
| payload       | object             | Dados específicos do evento.                   |

**Nota**: A `partition_key` deve ser o `room_id` para garantir a ordenação correta das mensagens dentro de uma sala no Kafka.

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

| Campo        | Tipo               | Descrição                     |
|--------------|--------------------|-------------------------------|
| user_id      | UUID               | Identificador do usuário.     |
| username     | string             | Nome de usuário escolhido.    |
| discriminator| string             | Número `#xxxx` gerado.        |
| created_at   | datetime (ISO8601) | Data de criação.              |

---

## UserAuthenticated

Gerado quando um usuário realiza login.

| Campo      | Tipo               | Descrição                    |
|------------|--------------------|------------------------------|
| user_id    | UUID               | Identificador do usuário.    |
| session_id | UUID               | Identificador da sessão.     |
| timestamp  | datetime (ISO8601) | Momento da autenticação.     |

---

## UserIdentityRotated

Gerado quando um usuário rotaciona seu identificador `#xxxx`.

| Campo            | Tipo               | Descrição                          |
|------------------|--------------------|------------------------------------|
| user_id          | UUID               | Identificador do usuário.          |
| old_discriminator| string             | Discriminador anterior (`#xxxx`).  |
| new_discriminator| string             | Novo discriminador (`#xxxx`).      |
| timestamp        | datetime (ISO8601) | Momento da rotação.                |

---

# 5. Presence Events

Eventos relacionados ao estado de presença dos usuários, utilizando Redis para armazenamento de baixa latência.

## UserConnected

Gerado quando o WebSocket Gateway estabelece uma nova conexão ativa com o cliente.

| Campo         | Tipo               | Descrição                                  |
|---------------|--------------------|--------------------------------------------|
| user_id       | UUID               | Identificador do usuário.                  |
| connection_id | UUID               | Identificador único da conexão.            |
| timestamp     | datetime (ISO8601) | Momento da conexão.                        |

---

## UserDisconnected

Gerado quando uma conexão WebSocket é encerrada.

| Campo         | Tipo               | Descrição                                           |
|---------------|--------------------|-----------------------------------------------------|
| user_id       | UUID               | Identificador do usuário.                           |
| connection_id | UUID               | Identificador da conexão encerrada.                 |
| reason        | string             | Motivo: `client_closed`, `ping_timeout`, `kicked`. |
| timestamp     | datetime (ISO8601) | Momento da desconexão.                              |

---

## UserPresenceChanged

Gerado quando o estado global de presença do usuário muda (ex: de offline para online ao abrir a primeira conexão).

| Campo     | Tipo               | Descrição                                      |
|-----------|--------------------|------------------------------------------------|
| user_id   | UUID               | Identificador do usuário.                      |
| status    | string             | `online`, `offline`, `invisible`.              |
| timestamp | datetime (ISO8601) | Momento da alteração de estado.                |

---

# 6. Session Events

Eventos relacionados a sessões de autenticação.

## UserSessionCreated

Gerado quando uma nova sessão é criada.

| Campo      | Tipo               | Descrição                    |
|------------|--------------------|------------------------------|
| session_id | UUID               | Identificador da sessão.     |
| user_id    | UUID               | Identificador do usuário.    |
| created_at | datetime (ISO8601) | Data de criação da sessão.   |
| expires_at | datetime (ISO8601) | Data de expiração da sessão. |

---

## UserSessionExpired

Gerado quando uma sessão expira.

| Campo      | Tipo               | Descrição                    |
|------------|--------------------|------------------------------|
| session_id | UUID               | Identificador da sessão.     |
| user_id    | UUID               | Identificador do usuário.    |
| expired_at | datetime (ISO8601) | Momento da expiração.        |

---

# 7. Room Events

Eventos relacionados a salas.

## RoomCreated

Gerado quando uma sala é criada.

| Campo      | Tipo               | Descrição                                  |
|------------|--------------------|--------------------------------------------|
| room_id    | UUID               | Identificador da sala.                     |
| owner_id   | UUID               | Identificador do proprietário.             |
| type       | string             | `PUBLIC`, `PRIVATE`, `DIRECT`, `TEMPORARY`.|
| created_at | datetime (ISO8601) | Data de criação.                           |

---

## RoomExpired

Gerado quando uma sala temporária expira.

| Campo      | Tipo               | Descrição                    |
|------------|--------------------|------------------------------|
| room_id    | UUID               | Identificador da sala.       |
| expired_at | datetime (ISO8601) | Momento da expiração.        |

---

## UserJoinedRoom

Gerado quando um usuário entra em uma sala.

| Campo      | Tipo               | Descrição                 |
|------------|--------------------|---------------------------|
| room_id    | UUID               | Identificador da sala.    |
| user_id    | UUID               | Identificador do usuário. |
| joined_at  | datetime (ISO8601) | Momento de entrada.       |

---

## UserLeftRoom

Gerado quando um usuário sai de uma sala.

| Campo      | Tipo               | Descrição                 |
|------------|--------------------|---------------------------|
| room_id    | UUID               | Identificador da sala.    |
| user_id    | UUID               | Identificador do usuário. |
| left_at    | datetime (ISO8601) | Momento de saída.         |

---

# 8. Message Events

Eventos relacionados a mensagens.

## MessageSent

Gerado quando uma mensagem é enviada.

| Campo            | Tipo               | Descrição                                          |
|------------------|--------------------|----------------------------------------------------|
| message_id       | UUID               | Identificador da mensagem.                         |
| room_id          | UUID               | Identificador da sala.                             |
| sender_id        | UUID               | Identificador do remetente.                        |
| message_type     | string             | `TEXT`, `IMAGE`, `VIDEO`, `FILE`, `AUDIO`, `SYSTEM`.|
| content          | string             | Conteúdo da mensagem (texto ou URL de mídia).      |
| is_zero_logging  | boolean            | Indica se a mensagem não deve ser persistida.      |
| ttl              | int (segundos)     | Tempo de vida da mensagem (0 = permanente).        |
| expires_at       | datetime (ISO8601) | Data de expiração (se TTL > 0).                    |
| timestamp        | datetime (ISO8601) | Momento do envio.                                   |

---

## MessageDelivered

Gerado quando o destinatário recebe a mensagem no cliente (acknowledgment de entrega).

| Campo        | Tipo               | Descrição                    |
|--------------|--------------------|------------------------------|
| message_id   | UUID               | Identificador da mensagem.   |
| room_id      | UUID               | Identificador da sala.       |
| user_id      | UUID               | Identificador do destinatário.|
| delivered_at | datetime (ISO8601) | Momento da entrega.          |

---

## MessageRead

Gerado quando o destinatário marca a mensagem como lida.

| Campo    | Tipo               | Descrição                    |
|----------|--------------------|------------------------------|
| message_id | UUID               | Identificador da mensagem.   |
| room_id  | UUID               | Identificador da sala.       |
| user_id  | UUID               | Identificador do destinatário.|
| read_at  | datetime (ISO8601) | Momento da leitura.          |

---

## MessageDeleted

Gerado quando uma mensagem é removida.

| Campo      | Tipo               | Descrição                    |
|------------|--------------------|------------------------------|
| message_id | UUID               | Identificador da mensagem.   |
| room_id    | UUID               | Identificador da sala.       |
| deleted_at | datetime (ISO8601) | Momento da remoção.          |

---

## MessageExpired

Gerado quando uma mensagem com TTL expira.

| Campo      | Tipo               | Descrição                    |
|------------|--------------------|------------------------------|
| message_id | UUID               | Identificador da mensagem.   |
| room_id    | UUID               | Identificador da sala.       |
| expired_at | datetime (ISO8601) | Momento da expiração.        |

---

## AttachmentUploaded

Gerado quando um anexo é enviado em uma mensagem.

| Campo         | Tipo               | Descrição                           |
|---------------|--------------------|-------------------------------------|
| attachment_id | UUID               | Identificador do anexo.             |
| message_id    | UUID               | Identificador da mensagem associada.|
| type          | string             | `IMAGE`, `VIDEO`, `FILE`, `AUDIO`.  |
| size          | int (bytes)        | Tamanho do arquivo.                 |
| storage_url   | string             | URL de acesso ao arquivo.           |

---

# 9. Moderation Events

Eventos relacionados a moderação.

## UserMuted

Gerado quando um usuário é silenciado.

| Campo          | Tipo               | Descrição                    |
|----------------|--------------------|------------------------------|
| room_id        | UUID               | Identificador da sala.       |
| moderator_id   | UUID               | Identificador do moderador.  |
| target_user_id | UUID               | Identificador do usuário alvo.|
| duration       | int (segundos)     | Duração do silenciamento.    |

---

## UserKicked

Gerado quando um usuário é removido da sala.

| Campo          | Tipo               | Descrição                    |
|----------------|--------------------|------------------------------|
| room_id        | UUID               | Identificador da sala.       |
| moderator_id   | UUID               | Identificador do moderador.  |
| target_user_id | UUID               | Identificador do usuário alvo.|

---

## UserBanned

Gerado quando um usuário é banido.

| Campo          | Tipo               | Descrição                    |
|----------------|--------------------|------------------------------|
| room_id        | UUID               | Identificador da sala.       |
| moderator_id   | UUID               | Identificador do moderador.  |
| target_user_id | UUID               | Identificador do usuário alvo.|

---

## ReportCreated

Gerado quando um usuário cria uma denúncia.

| Campo          | Tipo               | Descrição                    |
|----------------|--------------------|------------------------------|
| report_id      | UUID               | Identificador da denúncia.   |
| reporter_id    | UUID               | Identificador do denunciante.|
| target_user_id | UUID               | Identificador do denunciado. |
| message_id     | UUID               | Identificador da mensagem (se houver).|
| reason         | string             | Motivo da denúncia.          |
| created_at     | datetime (ISO8601) | Momento da denúncia.         |

---

# 10. Fluxo de Eventos

## 10.1 Fluxo Padrão (Persistente)
1. **User → Gateway**: Via WebSocket.
2. **Gateway → Kafka**: Publica `MessageSent`.
3. **Kafka → Fanout Worker**: Entrega em tempo real.
4. **Kafka → Persistence Worker**: Salva no Postgres.

## 10.2 Fluxo Zero Logging (Stream Only)
1. **Gateway → Kafka**: Publica `MessageSent` com `is_zero_logging: true`.
2. **Fanout Worker**: Entrega em tempo real normalmente.
3. **Persistence Worker**: Identifica a flag `true` e descarta o evento sem salvar no banco.

---

# 11. Versionamento de Eventos

Eventos possuem um campo `event_version` (int) para evolução do sistema sem quebrar consumidores existentes.

---

# 12. Benefícios da Arquitetura de Eventos

A arquitetura baseada em eventos oferece:

* desacoplamento entre serviços
* escalabilidade horizontal
* processamento assíncrono
* maior resiliência do sistema