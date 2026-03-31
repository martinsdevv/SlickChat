# SlickChat — Arquitetura de Segurança

## 1. Introdução

Este documento descreve a arquitetura de segurança do sistema SlickChat, projetada com foco em **privacidade, anonimato e proteção contra abuso**.

---

# 2. Princípios de Segurança

* anonimato por design
* minimização de coleta de dados
* proteção contra abuso
* proteção contra acesso não autorizado
* proteção contra exposição de dados sensíveis

---

# 3. Identidade e Autenticação

O sistema utiliza um modelo de identidade pseudônima no formato `username#xxxx`.

### Autenticação por senha

```
username#xxxx
password
```


A senha é armazenada utilizando hashing seguro.

### Recovery Key

Gerada no cadastro. Permite recuperação da conta caso o usuário esqueça a senha.

Características:
* exibida apenas uma vez
* não armazenada em texto puro
* deve ser guardada pelo usuário

### Modo Paranoico

Autenticação exclusivamente via **identity key**. Não há senha.

```
username#xxxx
identity_key
```


Se a chave for perdida, a conta é irrecuperável.

---

# 4. Proteção de Credenciais

O sistema adota as seguintes práticas:

* hashing seguro de senha (bcrypt/Argon2)
* hashing de identity keys
* hashing de recovery keys
* uso de UUID para identificação interna
* tokens de sessão armazenados apenas em hash no Redis e Postgres

Credenciais nunca são armazenadas em texto puro.

---

# 5. Proteção contra Abuso

### Rate Limiting

Limita o número de requisições por usuário, implementado com **Redis** para contadores globais.

Exemplos:
* máximo de mensagens por segundo
* máximo de conexões simultâneas
* máximo de tentativas de login

### Flood Control

Evita envio massivo de mensagens em curto intervalo de tempo.

### Moderação de Salas

Administradores podem:
* remover usuários
* silenciar usuários
* excluir mensagens
* banir usuários

---

# 6. Privacidade de Dados

O sistema **não coleta**:
* email
* telefone
* nome real
* localização

### Armazenamento de IP

O endereço IP pode ser armazenado apenas como **hash**, quando necessário para:
* rate limiting
* prevenção de abuso

---

# 7. Mensagens Efêmeras

O sistema suporta três modos de persistência:

| Modo                | Comportamento                                                                 |
|---------------------|-------------------------------------------------------------------------------|
| **Persistência normal** | Mensagens são armazenadas permanentemente.                                  |
| **TTL**             | Mensagens são armazenadas temporariamente e removidas após expiração.          |
| **Zero Logging Mode** | Mensagens não são persistidas; existem apenas no fluxo de eventos.            |

No Zero Logging Mode, o **Persistence Worker** descarta o evento imediatamente.

---

# 8. Segurança de Arquivos

Arquivos anexados são armazenados no **MinIO** (S3‑compatible).

Boas práticas:
* **pre‑signed URLs** para upload (validade curta)
* validação de tipo e tamanho
* controle de acesso por usuário
* metadados no Postgres sem conteúdo do arquivo

---

# 9. Segurança de Comunicação

Toda comunicação externa utiliza conexões seguras:
* **HTTPS** para API
* **WSS** (WebSocket Secure) para gateway

O Edge Router (Traefik) gerencia os certificados TLS.

---

# 10. Observabilidade e Logs

Logs devem:
* evitar dados sensíveis
* evitar conteúdo de mensagens
* registrar apenas eventos técnicos

Ferramenta utilizada: Prometheus

---

# 11. Resumo

A arquitetura de segurança garante:
* preservação do anonimato
* proteção de credenciais
* prevenção de abuso
* comunicação segura