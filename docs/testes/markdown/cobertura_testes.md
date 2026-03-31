# SlickChat — Plano de Testes

## 1. Introdução

Este documento define a estratégia de testes para o sistema SlickChat, abrangendo validação de requisitos funcionais, não funcionais e garantia de qualidade. Os testes são essenciais para assegurar anonimato, privacidade, comunicação em tempo real e resiliência da arquitetura baseada em eventos.

---

## 2. Objetivos

- Validar conformidade com os requisitos funcionais (RF) e não funcionais (RNF).
- Garantir a integridade dos fluxos de mensagens e eventos.
- Assegurar baixa latência e alta disponibilidade.
- Prevenir regressões durante a evolução do sistema.
- Verificar mecanismos de segurança e privacidade.

---

## 3. Tipos de Testes

### 3.1 Testes Unitários

| Aspecto | Descrição |
|---------|-----------|
| **Objetivo** | Validar componentes isolados (entidades, serviços, utilitários). |
| **Ferramentas** | Go `testing` + `testify` + `gomock`. |
| **Cobertura** | Mínimo 70% nas camadas de domínio e lógica de negócio. |
| **Exemplos** | - Validação de regras de negócio (ex: TTL, permissões).<br>- Geração de discriminadores.<br>- Hashing de credenciais.<br>- Parsing de eventos. |

### 3.2 Testes de Integração

| Aspecto | Descrição |
|---------|-----------|
| **Objetivo** | Validar interação entre componentes (API ↔ Redis, Worker ↔ Kafka, Gateway ↔ Redis). |
| **Ambiente** | Docker Compose com dependências reais (Postgres, Redis, Kafka, MinIO). |
| **Ferramentas** | Testcontainers (Go). |
| **Exemplos** | - Publicação e consumo de eventos no Kafka.<br>- Persistência de mensagens no Postgres.<br>- Atualização de presença no Redis.<br>- Geração de pre‑signed URLs no MinIO. |

### 3.3 Testes End‑to‑End (E2E)

| Aspecto | Descrição |
|---------|-----------|
| **Objetivo** | Simular fluxos completos de usuário (frontend + API + Gateway). |
| **Ambiente** | Ambiente de staging com todos os serviços. |
| **Ferramentas** | Playwright (frontend) + scripts Go para APIs. |
| **Exemplos** | - Cadastro e login.<br>- Envio e recebimento de mensagens em tempo real.<br>- Moderação (kick, mute, ban).<br>- Expiração de mensagem TTL.<br>- Zero Logging Mode. |

### 3.4 Testes de Performance e Carga

| Aspecto | Descrição |
|---------|-----------|
| **Objetivo** | Avaliar latência, throughput e escalabilidade sob carga. |
| **Ferramentas** | k6 (WebSocket/HTTP), Vegeta (HTTP). |
| **Métricas críticas** | - Latência de entrega de mensagem < 200ms (p95).<br>- Suporte a 10.000 conexões WebSocket simultâneas.<br>- Processamento de 10.000 mensagens/segundo no Kafka.<br>- Uso de CPU/memória dentro dos limites. |

### 3.5 Testes de Segurança

| Aspecto | Descrição |
|---------|-----------|
| **Objetivo** | Verificar mecanismos de autenticação, autorização, rate limit e privacidade. |
| **Ferramentas** | Scripts personalizados, OWASP ZAP (opcional). |
| **Exemplos** | - Tentativa de acesso a salas sem permissão.<br>- Rate limit global (Redis) – exceder limite e verificar bloqueio.<br>- Upload de arquivos maliciosos (tamanho, tipo).<br>- Injeção de conteúdo em mensagens.<br>- Validação de recovery key e identity key. |

### 3.6 Testes de Resiliência

| Aspecto | Descrição |
|---------|-----------|
| **Objetivo** | Validar comportamento do sistema diante de falhas. |
| **Exemplos** | - Reinício de um Worker – mensagens não processadas devem ser reprocessadas.<br>- Falha do Kafka – mecanismos de retry e circuit breaker.<br>- Reconexão WebSocket – cliente recupera estado da sessão.<br>- Falha do Redis – degradação controlada (ex: fallback para cache local). |

---

## 4. Cobertura por Componente

| Componente            | Unitários | Integração | E2E | Performance | Segurança | Resiliência |
|-----------------------|-----------|------------|-----|-------------|-----------|-------------|
| API Service           | ✓         | ✓          | ✓   | ✓           | ✓         | ✓           |
| Realtime Gateway      | ✓         | ✓          | ✓   | ✓           | ✓         | ✓           |
| Workers (Fanout, Persistence, TTL, Moderation) | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Kafka (eventos)       | –         | ✓          | –   | ✓           | –         | ✓           |
| Redis (presença, sessões, rate limit) | – | ✓ | – | ✓ | – | ✓ |
| Postgres              | –         | ✓          | –   | ✓           | –         | ✓ |
| MinIO                 | –         | ✓          | –   | –           | ✓         | – |
| Frontend (React)      | –         | –          | ✓   | –           | –         | – |

---

## 5. Estratégia de Testes Específicos por Funcionalidade

### 5.1 Identidade e Autenticação

- Testar criação de conta com username, geração de discriminator e recovery key.
- Login com senha, identity key (modo paranoico).
- Recuperação com recovery key.
- Rotação de discriminador.

### 5.2 Chat Privado e Salas

- Envio de mensagem em sala pública/privada.
- Entrada em sala privada com convite.
- Expiração de sala temporária (TTL Worker).
- Moderação: kick, mute, ban, exclusão de mensagem.
- Denúncia de mensagem/usuario.

### 5.3 Mensagens Efêmeras

- Mensagem com TTL – verificar remoção após expiração.
- Mensagem auto‑destrutiva – remoção após leitura.
- Zero Logging Mode – mensagem não persiste no Postgres, mas é entregue.

### 5.4 Anexos

- Upload de imagem/vídeo/áudio via pre‑signed URL.
- Validação de tamanho e tipo.
- Acesso ao anexo após upload.

### 5.5 Presença

- Atualização de status online/offline/invisible.
- Reconexão WebSocket e recuperação de estado.

### 5.6 Segurança e Abuso

- Rate limit por usuário/IP (ex: 10 mensagens/segundo) – verificar bloqueio.
- Flood control – mensagens em curto intervalo são limitadas.
- Moderação – usuário silenciado não consegue enviar.

---

## 6. Ambiente de Testes

- **Desenvolvimento:** Docker Compose local com todos os serviços.
- **Staging:** Cluster Kubernetes ou VM com ambiente completo, usado para E2E e performance.
- **CI/CD:** GitHub Actions (ou GitLab CI) executando testes unitários, integração e lint a cada push; testes E2E e performance em merge requests para branches principais.

---

## 7. Ferramentas e Tecnologias

| Tipo               | Ferramenta            | Uso                                                |
|--------------------|-----------------------|----------------------------------------------------|
| Unitários          | Go `testing`, `testify` | Testes em Go.                                     |
| Mocks              | `gomock`              | Mock de interfaces para isolamento.                |
| Integração         | Testcontainers        | Gerenciamento de contêineres (Postgres, Kafka, etc). |
| E2E                | Playwright, Go HTTP client | Testes de fluxo completo.                      |
| Performance        | k6                    | Simulação de carga WebSocket e HTTP.               |
| Segurança          | OWASP ZAP (opcional)  | Análise de vulnerabilidades.                       |
| CI                 | GitHub Actions        | Pipeline automatizado.                             |

---

## 8. Critérios de Aceitação

- Todos os testes unitários e de integração passam.
- Cobertura de código ≥70% nas camadas críticas.
- Testes E2E críticos (login, envio de mensagem, moderação) executados com sucesso.
- Teste de carga com 10k usuários simultâneos atende SLA (p95 < 200ms).
- Nenhuma vulnerabilidade crítica identificada nos testes de segurança.

---

## 9. Execução e Monitoramento

- Relatórios de cobertura gerados a cada execução.
- Logs estruturados para diagnóstico de falhas em testes.
- Métricas de performance coletadas e armazenadas para análise histórica.

---