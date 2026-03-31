# SlickChat — Modelo de Domínio

## 1. Introdução

Este documento descreve o **modelo de domínio do sistema SlickChat**, identificando as entidades principais e seus relacionamentos.

---

# 2. Visão Geral do Domínio

O domínio é composto pelas seguintes entidades principais:

```
User
IdentityKey
Session
Room
RoomMembership
Message
Attachment
ModerationAction
Report
```


O domínio é dividido em duas camadas de persistência:

1. **Camada Persistente (Postgres)**: Entidades de identidade, configuração de salas e histórico de mensagens (quando permitido).
2. **Camada Efêmera (Redis)**: Estado de presença em tempo real, sessões ativas e contadores de rate limit.

---

# 3. User

Representa um usuário do sistema.

| Atributo        | Tipo               | Descrição                                      |
|-----------------|--------------------|------------------------------------------------|
| id              | UUID               | Identificador interno.                         |
| username        | string             | Nome de usuário escolhido (sem discriminator). |
| discriminator   | string             | Número `#xxxx` gerado.                         |
| password_hash   | string             | Hash da senha (se não estiver em modo paranoico).|
| paranoid_mode   | boolean            | Indica se autenticação é por identity key.     |
| created_at      | datetime (ISO8601) | Data de criação.                               |

---

# 4. IdentityKey

Utilizada para autenticação no **modo paranoico**.

| Atributo   | Tipo               | Descrição                          |
|------------|--------------------|------------------------------------|
| id         | UUID               | Identificador da chave.            |
| user_id    | UUID               | Referência ao usuário.             |
| key_hash   | string             | Hash da identity key.              |
| created_at | datetime (ISO8601) | Data de criação.                   |

---

# 5. Session

Representa uma sessão ativa de autenticação.

| Atributo   | Tipo               | Descrição                          |
|------------|--------------------|------------------------------------|
| id         | UUID               | Identificador da sessão.           |
| user_id    | UUID               | Referência ao usuário.             |
| token_hash | string             | Hash do token de sessão.           |
| created_at | datetime (ISO8601) | Data de criação.                   |
| expires_at | datetime (ISO8601) | Data de expiração.                 |
| ip_hash    | string             | Hash do endereço IP (opcional).    |

---

# 6. Room

Representa qualquer espaço de comunicação (sala ou chat direto).

| Atributo      | Tipo               | Descrição                                          |
|---------------|--------------------|----------------------------------------------------|
| id            | UUID               | Identificador da sala.                             |
| name          | string             | Nome da sala (se aplicável).                       |
| type          | string             | `PUBLIC`, `PRIVATE`, `DIRECT`, `TEMPORARY`.        |
| owner_id      | UUID               | Identificador do proprietário.                     |
| ttl           | int (segundos)     | Tempo de vida da sala (0 = permanente).            |
| paranoid_mode | boolean            | Se true, mensagens não são persistidas.            |
| zero_logging  | boolean            | Se true, mensagens não são armazenadas.            |
| created_at    | datetime (ISO8601) | Data de criação.                                   |
| expires_at    | datetime (ISO8601) | Data de expiração (se TTL > 0).                    |

---

# 7. RoomMembership

Relaciona usuários às salas e define papéis.

| Atributo   | Tipo               | Descrição                          |
|------------|--------------------|------------------------------------|
| id         | UUID               | Identificador da associação.       |
| room_id    | UUID               | Referência à sala.                 |
| user_id    | UUID               | Referência ao usuário.             |
| role       | string             | `ADMIN`, `MODERATOR`, `MEMBER`.    |
| joined_at  | datetime (ISO8601) | Data de entrada.                   |

---

# 8. Message

Representa mensagens enviadas em uma sala.

| Atributo            | Tipo               | Descrição                                          |
|---------------------|--------------------|----------------------------------------------------|
| id                  | UUID               | Identificador da mensagem.                         |
| room_id             | UUID               | Referência à sala.                                 |
| sender_id           | UUID               | Referência ao remetente.                           |
| content             | string             | Texto da mensagem ou URL de mídia.                 |
| message_type        | string             | `TEXT`, `IMAGE`, `VIDEO`, `FILE`, `AUDIO`, `SYSTEM`.|
| ttl                 | int (segundos)     | Tempo de vida (0 = permanente).                    |
| created_at          | datetime (ISO8601) | Data de envio.                                     |
| expires_at          | datetime (ISO8601) | Data de expiração (se TTL > 0).                    |
| destroy_after_read  | boolean            | Se true, mensagem é removida após leitura.         |

---

# 9. Attachment

Representa arquivos anexados a mensagens.

| Atributo    | Tipo               | Descrição                          |
|-------------|--------------------|------------------------------------|
| id          | UUID               | Identificador do anexo.            |
| message_id  | UUID               | Referência à mensagem.             |
| storage_url | string             | URL de acesso ao arquivo.          |
| type        | string             | `IMAGE`, `VIDEO`, `FILE`, `AUDIO`. |
| size        | int (bytes)        | Tamanho do arquivo.                |
| created_at  | datetime (ISO8601) | Data de upload.                    |

---

# 10. ModerationAction

Registra ações de moderação em uma sala.

| Atributo       | Tipo               | Descrição                          |
|----------------|--------------------|------------------------------------|
| id             | UUID               | Identificador da ação.             |
| room_id        | UUID               | Referência à sala.                 |
| moderator_id   | UUID               | Referência ao moderador.           |
| target_user_id | UUID               | Referência ao usuário alvo.        |
| action         | string             | `KICK`, `MUTE`, `DELETE_MESSAGE`, `BAN`. |
| created_at     | datetime (ISO8601) | Data da ação.                      |

---

# 11. Report

Registra denúncias feitas por usuários.

| Atributo       | Tipo               | Descrição                          |
|----------------|--------------------|------------------------------------|
| id             | UUID               | Identificador da denúncia.         |
| reporter_id    | UUID               | Referência ao denunciante.         |
| target_user_id | UUID               | Referência ao denunciado.          |
| message_id     | UUID               | Referência à mensagem (opcional).  |
| reason         | string             | Motivo da denúncia.                |
| created_at     | datetime (ISO8601) | Data da denúncia.                  |

---

# 12. Estrutura Conceitual do Domínio

Relações principais entre entidades:

```
User
├─ IdentityKey (1:0..1)
├─ Session (1:0..)
└─ RoomMembership (1:0..)

Room
├─ RoomMembership (1:0..)
├─ Message (1:0..)
└─ ModerationAction (1:0..*)

Message
└─ Attachment (1:0..*)

Report
├─ User (reporter)
└─ User (target)
```


---

# 13. Considerações de Arquitetura

O modelo de domínio foi projetado para suportar uma arquitetura baseada em eventos. Mensagens e ações são representadas como eventos distribuídos via Kafka, permitindo:

* comunicação em tempo real
* processamento assíncrono
* escalabilidade horizontal
* desacoplamento entre serviços