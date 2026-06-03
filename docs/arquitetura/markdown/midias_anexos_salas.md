# Mídias — anexos de mensagem, avatar e banner de sala

Este documento fixa as regras de ciclo de vida e implementação, alinhado a `definicoes_ui.md`, `modelo_dominio.md` e `arquitetura_sistema.md`.

## 1. Princípio central

**Nenhum objeto no MinIO sobrevive ao “destroy” do contexto que o autorizou.**

| Contexto | Quando apagar o arquivo |
|----------|------------------------|
| Mensagem com TTL | `message.expired` / remoção física no worker |
| Mensagem apagada | `message.deleted` |
| Sala zero logging / paranoid (mensagem não persiste) | Mesmo evento de destroy da mensagem (sem linha em `attachments` obrigatória) |
| Troca de avatar/banner | Após confirmar o novo upload, apagar a chave anterior |
| Sala encerrada (futuro) | Apagar `avatar_object_key`, `banner_object_key` e anexos órfãos da sala |

Metadados em Postgres (`attachments`, chaves em `rooms`) **só existem** quando o contexto é persistível. Caso contrário, a chave do objeto vive apenas no fluxo da mensagem (Redis / evento) até o destroy.

## 2. Anexos de mensagem

### 2.1 Fluxo (com placeholder otimista)

1. Cliente escolhe arquivo → UI mostra **placeholder otimista** (preview local + estado `uploading`).
2. `POST /media/upload-request` — API valida membro da sala, tipo, tamanho e `room.CanPersistAttachments()` (ou política efêmera).
3. Cliente `PUT` na pre-signed URL (MinIO).
4. `POST /media/upload-complete` — registra metadado **somente se** a mensagem for persistível; senão retorna `object_key` efêmero para o WS.
5. `send_message` com `message_type` (`IMAGE`, …) e `content` / referência ao objeto.
6. Em falha após PUT: cliente chama cancelamento ou worker de limpeza remove objeto órfão (TTL curto na chave).

### 2.2 Persistência vs efêmero

```text
CanPersistAttachments() == CanPersistMessages()
  (= false se zero_logging ou paranoid_mode)
```

- **Persistível:** linha em `attachments` + objeto em `messages/{room_id}/{message_id}/{attachment_id}`.
- **Efêmero:** sem linha em `attachments`; `object_key` associado à mensagem no Redis até `message.deleted` / `message.expired`; **Persistence Worker** não grava mensagem; **Fanout / serviço de mídia** executa `DeleteObject` no destroy.

### 2.3 Delete no destroy

- **Persistence Worker:** ao processar `message.deleted.v1` e `message.expired.v1`, após `DELETE` da mensagem, listar anexos por `message_id`, apagar objetos no MinIO e remover linhas em `attachments`.
- **Zero logging / não persistido:** handler do evento de destroy (fanout ou gateway) lê chaves do Redis (`message:{id}` → `attachment_object_keys`) e apaga no MinIO.

## 3. Avatar e banner de sala

Branding da sala — **persistente** enquanto a sala existir (independente de zero logging nas mensagens).

| Campo | Uso |
|-------|-----|
| `avatar_object_key` | Ícone/lista (~256², crop quadrado) |
| `banner_object_key` | Faixa no painel da sala / cabeçalho (~1200×400) |

Chaves sugeridas:

- `rooms/{room_id}/avatar/{upload_id}`
- `rooms/{room_id}/banner/{upload_id}`

Somente **ADMIN** da sala pode solicitar upload. Ao completar, substituir a chave na linha `rooms` e **apagar** a chave antiga no MinIO.

URLs de leitura: pre-signed GET de curta duração ou rota autenticada da API — nunca bucket público.

## 4. UI

- **Mensagem:** placeholder otimista até `upload-complete` + ACK do envio; depois preview (imagem inline, etc.).
- **Sala:** manter placeholder com iniciais quando não houver `avatar_url`; banner com gradiente neutro ou imagem quando `banner_url` existir (painel Info + cabeçalho do chat).

## 5. Fases de implementação

1. MinIO no `deploy/compose.yml`, migration `008_media.sql`, domínio + contratos + purge no worker (mensagens persistidas).
2. Endpoints `/media/upload-request` e `/media/upload-complete`; extensão WS `message_type`.
3. Frontend: anexo com placeholder; avatar/banner no painel (upload admin).
4. Purge efêmero no fanout para zero logging; testes §5.4 de `cobertura_testes.md`.
