# SlickChat — Modelo de Domínio

## 1. Introdução

Este documento descreve o **modelo de domínio do sistema SlickChat**.

O objetivo do modelo de domínio é identificar as entidades principais do sistema e os relacionamentos entre elas, servindo como base para:

* diagrama de classes
* modelagem do banco de dados
* definição da arquitetura do sistema

---

# 2. Visão Geral do Domínio

O domínio do SlickChat é composto pelas seguintes entidades principais:

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

Essas entidades representam os principais conceitos do sistema de chat.

O domínio é dividido em duas camadas de persistência:

1. **Camada Persistente (Postgres)**: Entidades de identidade, configuração de salas e histórico de mensagens (quando permitido).
2. **Camada Efêmera (Redis)**: Estado de presença em tempo real, sessões ativas e contadores de rate limit.

**Nota sobre Presença**: O status (`online`, `offline`, `invisible`) é uma entidade puramente volátil. Não existe uma tabela de "presença" no banco; o Gateway lê e escreve diretamente no Redis para garantir latência sub-milissegundo.

---

# 3. User

Representa um usuário do sistema.

Cada usuário possui uma identidade exibida no formato:

```
username#xxxx
```

Atributos:

```
id (UUID)
username
discriminator (#xxxx)
password_hash
paranoid_mode
created_at
```

Observações:

* O identificador real do sistema é o `UUID`.
* `username#xxxx` é apenas uma identidade de exibição.
* O campo `discriminator` pode ser **rotacionado** pelo usuário, gerando um novo identificador público.

---

# 4. IdentityKey

A entidade **IdentityKey** é utilizada para autenticação no **modo paranoico**.

Nesse modo, o usuário não utiliza senha.

Atributos:

```
id
user_id
key_hash
created_at
```

Características:

* a chave é gerada apenas uma vez
* nunca é armazenada em texto puro
* se perdida, a conta não pode ser recuperada

---

# 5. Session

Representa uma sessão ativa de autenticação. No sistema, a `Session` existe como um registro de auditoria no Postgres e como uma chave ativa no Redis.

Atributos:

```
id
user_id
token_hash
created_at
expires_at
ip_hash
```

Observações:

* um usuário pode possuir múltiplas sessões simultâneas
* o endereço IP pode ser armazenado apenas como **hash**, preservando anonimato
* Ao expirar no Redis, a sessão é considerada encerrada, mesmo que o registro no Postgres ainda exista.
* O WebSocket Gateway valida o token contra o Redis em cada nova conexão (evento `UserConnected`).

---

# 6. Room

A entidade **Room** representa qualquer espaço de comunicação no sistema. O comportamento de persistência da sala é ditado pelas flags de privacidade.

Todos os tipos de conversa são modelados como **Room**, incluindo chats privados.

Tipos possíveis:

```
PUBLIC
PRIVATE
DIRECT
TEMPORARY
```

Atributos:

```
id
name
type
owner_id
ttl
paranoid_mode
zero_logging
created_at
expires_at
```

Características:

* salas podem ser públicas ou privadas
* salas temporárias possuem tempo de vida configurável e é removida fisicamente do banco pelo **TTL Worker** após o `expires_at`
* salas podem operar em **modo paranoia**
* em salas `Zero Logging`, as mensagens existem apenas como eventos no Kafka e no fluxo de memória do Fanout Worker.

---

# 7. RoomMembership

A entidade **RoomMembership** relaciona usuários às salas.

Atributos:

```
id
room_id
user_id
role
joined_at
```

Papéis possíveis:

```
ADMIN
MODERATOR
MEMBER
```

Responsabilidades:

* controlar quem participa da sala
* controlar permissões de moderação

---

# 8. Message

A entidade **Message** representa mensagens enviadas em uma sala.

Atributos:

```
id
room_id
sender_id
content
message_type
ttl
created_at
expires_at
destroy_after_read
```

Tipos de mensagem:

```
TEXT
IMAGE
VIDEO
FILE
AUDIO
SYSTEM
```

Observações:

* mensagens podem possuir **TTL configurável**
* mensagens podem ser **auto destrutivas após leitura** (`destroy_after_read`)
* quando `zero_logging` está ativo na sala, mensagens não são persistidas
* quando uma mensagem é deletada (por TTL ou auto-destruição), o registro é removido fisicamente (`HARD DELETE`) para garantir o requisito de privacidade e anonimato do sistema.

---

# 9. Attachment

A entidade **Attachment** representa arquivos anexados a mensagens.

Atributos:

```
id
message_id
storage_url
type
size
created_at
```

Tipos possíveis:

```
IMAGE
VIDEO
FILE
AUDIO
```

Observações:

* arquivos são armazenados em um sistema de armazenamento externo
* o banco de dados armazena apenas **metadados e URL de acesso**

---

# 10. ModerationAction

Representa ações de moderação executadas em uma sala.

Atributos:

```
id
room_id
moderator_id
target_user_id
action
created_at
```

Tipos de ação:

```
KICK
MUTE
DELETE_MESSAGE
BAN
```

---

# 11. Report

Representa denúncias realizadas por usuários.

Atributos:

```
id
reporter_id
target_user_id
message_id
reason
created_at
```

O sistema pode utilizar essas informações para análise de abuso e moderação.

---

# 12. Estrutura Conceitual do Domínio

Relações principais entre entidades:

```
User
 ├─ IdentityKey
 ├─ Session
 └─ RoomMembership

Room
 ├─ RoomMembership
 ├─ Message
 └─ ModerationAction

Message
 └─ Attachment

Report
```

---

# 13. Considerações de Arquitetura

O modelo de domínio foi projetado para suportar uma arquitetura baseada em eventos.

Mensagens e ações do sistema são representadas como **eventos distribuídos através de um broker de mensageria (Kafka)**.

Essa abordagem permite:

* comunicação em tempo real
* processamento assíncrono
* escalabilidade horizontal
* desacoplamento entre serviços
