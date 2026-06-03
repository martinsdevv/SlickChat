# Demo pública na sua máquina (Cloudflare Tunnel)

Guia para ter um **link HTTPS compartilhável** (ex.: `https://algo.trycloudflare.com`) rodando o SlickChat no seu PC — ideal para apresentar amanhã.

## Pré-requisitos

| Ferramenta | Uso |
|------------|-----|
| Docker + Compose | Postgres, Redis, Kafka, MinIO |
| Go 1.22+ | API, gateway, workers |
| Node 20+ | Build do frontend |
| `cloudflared` | Túnel HTTPS gratuito (sem abrir porta no roteador) |

```bash
# Arch Linux
sudo pacman -S cloudflared docker docker-compose

# Frontend
cd frontend/web && npm ci
```

## Passo a passo (primeira vez)

### 1. Infraestrutura

```bash
make infra-up
make demo-migrate
```

### 2. Build do frontend + backend

```bash
make demo-build      # gera frontend/web/dist
make demo-backend    # API :8081, gateway :8080, workers
make demo-proxy      # Caddy em http://localhost:3000
```

Teste local: abra **http://localhost:3000**, crie conta, salas e mensagens.

### 3. Link público (Cloudflare)

Em **outro terminal**:

```bash
make demo-tunnel
```

O `cloudflared` imprime algo como:

```text
https://random-words.trycloudflare.com
```

Esse é o link para enviar no grupo / usar na apresentação.

### 4. Fotos e anexos pelo túnel (opcional)

Uploads vão ao MinIO. Para funcionar **fora da sua rede**:

```bash
make demo-set-minio URL=https://random-words.trycloudflare.com/storage
make demo-restart-api
```

Por padrão os uploads vão pela API (`PUT /api/media/upload-put`), não para `localhost:9000` — assim funciona no **celular** e no túnel. O navegador remoto não consegue acessar o MinIO da sua máquina.

Confirme no log da API ao subir: `upload_via_api=true`.

**Importante:** `make demo-restart-api` **não altera** o `.env` — só relê o arquivo. Evite `cp deploy/.env.example deploy/.env`.

Leitura de imagens passa por `/api/media/object` e funciona sem configurar MinIO público.

## Comandos úteis

| Comando | O que faz |
|---------|-----------|
| `make demo-up` | infra + migrate + build + backend + proxy |
| `make demo-tunnel` | expõe `:3000` na internet (Cloudflare) |
| `make demo-set-minio URL=...` | grava `MINIO_PUBLIC_URL` no `.env` |
| `make demo-restart-api` | reinicia API lendo `deploy/.env` (não edita o arquivo) |
| `make demo-down` | para backend, proxy e infra |
| `make demo-logs` | logs dos serviços Go |
| `tail -f deploy/logs/api.log` | debug API |

## Checklist antes do pitch

- [ ] Badge **Conectado** verde na sala
- [ ] Duas contas (duas abas anônimas ou celular + notebook)
- [ ] Salas `publica`, `zero log`, `temporaria` com mensagens (ver `docs/pitch/preparacao_demo.md`)
- [ ] Túnel rodando no terminal (não fechar até acabar a demo)
- [ ] PC ligado na tomada / sem suspender

## Parar tudo

```bash
make demo-down
# Ctrl+C no terminal do cloudflared
```

## Alternativa: domínio fixo no Cloudflare

Se você já tem domínio na Cloudflare, pode criar um **Named Tunnel** no dashboard (Zero Trust → Networks → Tunnels) apontando para `http://localhost:3000`. O quick tunnel acima é mais rápido para um dia.

## Limitações

- URL do quick tunnel **muda** cada vez que reinicia o `cloudflared` (a menos que use named tunnel).
- Máquina precisa ficar ligada durante a demo.
- Primeira conexão pode levar alguns segundos (cold start).
