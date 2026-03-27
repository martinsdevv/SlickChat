# SlickChat — Arquitetura do Sistema

## 1. Introdução

Este documento descreve a arquitetura do sistema SlickChat.

A arquitetura foi projetada para suportar comunicação em tempo real, processamento assíncrono e escalabilidade horizontal utilizando uma abordagem baseada em eventos.

---

# 2. Estilo Arquitetural

O SlickChat adota uma arquitetura **Event-Driven distribuída**.

Características principais:

* comunicação em tempo real via WebSocket
* processamento assíncrono baseado em eventos
* desacoplamento entre componentes
* escalabilidade horizontal

A comunicação entre componentes ocorre através de um **broker de mensageria (Kafka)**.

Além disso, o sistema utiliza **Redis** para dados temporários e operações de baixa latência.

---

# 3. Componentes Principais

O sistema é composto pelos seguintes containers principais.

```
Client Web
Realtime Gateway
API Service
Kafka Event Stream
Workers
Redis
Postgres Database
Media Storage
```

Cada componente possui uma responsabilidade específica dentro da arquitetura distribuída.

---

# 4. Realtime Gateway

O **Realtime Gateway** é responsável pela comunicação em tempo real com os clientes.

Funções:

* gerenciar conexões WebSocket
* autenticar usuários
* gerenciar sessões ativas
* atualizar presença de usuários
* receber eventos do cliente
* publicar eventos no broker de mensageria
* distribuir mensagens recebidas para clientes conectados

O gateway é **stateless**, permitindo múltiplas instâncias em execução.

Informações temporárias como presença e sessões podem ser armazenadas em **Redis**.

> **Nota de Implementação**: Para viabilizar a escalabilidade horizontal, o **Message Fanout Worker** não envia a mensagem diretamente ao socket. Ele publica a mensagem num tópico interno do **Redis (Pub/Sub)** ou via **gRPC stream**. Cada instância do **Realtime Gateway** subscreve-se a este canal e, ao receber a mensagem, verifica se o destinatário está conectado localmente para realizar a entrega final.

---

# 5. API Service

O **API Service** fornece endpoints HTTP para funcionalidades administrativas e operações que não exigem comunicação em tempo real.

Exemplos:

* criação de conta
* autenticação
* gerenciamento de salas
* moderação
* upload de anexos

Esse serviço também pode publicar eventos no Kafka quando necessário.

---

# 6. Broker de Mensageria

O SlickChat utiliza **Kafka** como broker de eventos.

Responsabilidades:

* transporte de eventos entre serviços
* desacoplamento entre componentes
* suporte a alta taxa de mensagens
* buffer de eventos para processamento assíncrono

Eventos comuns do sistema:

```
UserJoinedRoom
UserLeftRoom
MessageSent
MessageDeleted
RoomCreated
RoomExpired
UserMuted
UserBanned
ReportCreated
```

Kafka permite que múltiplos consumidores processem eventos de forma independente.

---

# 7. Redis

**Redis** é utilizado para armazenar dados temporários que exigem acesso rápido.

Principais responsabilidades:

* presença de usuários (online/offline)
* armazenamento de sessões ativas
* controle de rate limiting
* cache de dados frequentemente acessados

Essas operações exigem latência muito baixa e não precisam de persistência permanente.

---

# 8. Workers

Workers são responsáveis pelo processamento assíncrono de eventos consumidos do Kafka.

Tipos de workers:

### Message Fanout Worker

Responsável por distribuir mensagens para usuários conectados através do gateway.

Para viabilizar a escalabilidade horizontal, o **Message Fanout Worker** não envia a mensagem diretamente ao socket. Ele publica a mensagem num tópico interno do **Redis (Pub/Sub)** ou via **gRPC stream**. Cada instância do **Realtime Gateway** subscreve-se a este canal e, ao receber a mensagem, verifica se o destinatário está conectado localmente para realizar a entrega final.

### Persistence Worker

Responsável por persistir mensagens e eventos no banco de dados.

Esse worker não persiste mensagens quando a sala está em **Zero Logging Mode**.

### TTL Worker

Responsável por remover mensagens expiradas e salas temporárias.

### Moderation Worker

Responsável por processar ações de moderação.

Workers podem ser escalados horizontalmente conforme a carga do sistema.

---

# 9. Persistência de Dados

O sistema utiliza **Postgres** para armazenamento persistente.

Dados armazenados:

* usuários
* salas
* mensagens
* membros de salas
* ações de moderação
* denúncias

Mensagens temporárias podem ser removidas automaticamente através de processos de expiração.

---

# 10. Estratégia de Mensagens Efêmeras

O sistema suporta três modos de persistência.

### Persistência normal

Mensagens são armazenadas permanentemente no banco de dados.

### TTL

Mensagens são armazenadas temporariamente e removidas após expiração.

### Zero Logging Mode

Mensagens não são armazenadas no banco de dados e existem apenas no fluxo de eventos distribuídos pelo sistema.
No **Zero Logging Mode**, a inteligência de descarte reside no **Persistence Worker**. Ao consumir um evento marcado com a flag de zero logging, o worker interrompe o fluxo de persistência imediatamente, garantindo que os dados nunca toquem o storage do Postgres.

---

# 11. Armazenamento de Mídia

Arquivos anexados são armazenados em um sistema de armazenamento externo.

Exemplos possíveis:

* S3
* MinIO
* Cloud Storage

O banco de dados armazena apenas o **metadata do arquivo**.

---

# 12. Escalabilidade

A arquitetura foi projetada para permitir escalabilidade horizontal.

Componentes escaláveis:

* Realtime Gateway
* Workers
* API Service

Kafka permite distribuição de eventos entre múltiplas instâncias consumidoras.

Redis pode ser executado em cluster para suportar alto volume de operações.

---

# 13. Segurança

O sistema implementa diversas medidas de segurança.

### Autenticação

* username#xxxx
* senha
* recovery key
* identity key (modo paranoico)

### Proteção contra abuso

* rate limiting
* controle de flood
* moderação de salas

### Privacidade

* ausência de coleta de dados pessoais
* armazenamento mínimo de informações identificáveis

---

# 14. Observabilidade

O sistema deve registrar logs estruturados para monitoramento.

Informações importantes:

* eventos de mensagens
* conexões WebSocket
* erros do sistema
* ações de moderação

Logs devem evitar armazenamento de informações sensíveis.

Ferramentas de observabilidade podem incluir:

* métricas (Prometheus)
* visualização (Grafana)
* agregação de logs (ELK stack)
