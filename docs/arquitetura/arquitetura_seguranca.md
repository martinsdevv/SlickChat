# SlickChat — Arquitetura de Segurança

## 1. Introdução

Este documento descreve a arquitetura de segurança do sistema SlickChat.

O SlickChat foi projetado com foco em **privacidade, anonimato e proteção contra abuso**, adotando princípios de segurança desde o design inicial do sistema.

---

# 2. Princípios de Segurança

A arquitetura de segurança do SlickChat segue os seguintes princípios:

* anonimato por design
* minimização de coleta de dados
* proteção contra abuso
* proteção contra acesso não autorizado
* proteção contra exposição de dados sensíveis

---

# 3. Identidade e Autenticação

O sistema utiliza um modelo de identidade pseudônima.

Formato:

```
username#xxxx
```

O sistema suporta três mecanismos de autenticação.

---

## Autenticação por senha

```
username#xxxx
password
```

A senha é armazenada utilizando hashing seguro.

---

## Recovery Key

Durante o cadastro, o sistema gera uma **recovery key**.

Essa chave permite recuperação da conta caso o usuário esqueça a senha.

Características:

* exibida apenas uma vez
* não armazenada em texto puro
* deve ser guardada pelo usuário

---

## Modo Paranoico

O SlickChat oferece um modo opcional chamado **Paranoid Mode**.

Nesse modo:

* não existe senha
* autenticação ocorre apenas via **identity key**

Formato:

```
username#xxxx
identity_key
```

Se a chave for perdida, a conta não pode ser recuperada.

---

# 4. Proteção de Credenciais

O sistema adota as seguintes práticas:

* hashing seguro de senha
* hashing de identity keys
* hashing de recovery keys
* uso de UUID para identificação interna

Credenciais nunca são armazenadas em texto puro.

---

# 5. Proteção contra Abuso

O sistema implementa mecanismos para prevenir abuso e spam.

## Rate Limiting

Limita o número de requisições por usuário.

Exemplo:

```
máximo de mensagens por segundo
máximo de conexões simultâneas
```

Rate limiting pode ser implementado utilizando **Redis** para controle rápido de contadores.

---

## Flood Control

Evita envio massivo de mensagens em curto intervalo de tempo.

---

## Moderação de Salas

Administradores de sala possuem ferramentas para:

* remover usuários
* silenciar usuários
* excluir mensagens
* banir usuários

---

# 6. Privacidade de Dados

O SlickChat minimiza coleta de dados pessoais.

O sistema **não coleta**:

* email
* telefone
* nome real
* localização

---

## Armazenamento de IP

O endereço IP pode ser armazenado apenas como **hash**, quando necessário para:

* rate limiting
* prevenção de abuso

---

# 7. Mensagens Efêmeras

O sistema suporta mensagens temporárias.

Tipos:

### Persistência normal

Mensagens são armazenadas permanentemente.

---

### TTL

Mensagens são armazenadas temporariamente e removidas após expiração.

---

### Zero Logging Mode

Mensagens não são armazenadas no banco de dados.

Elas existem apenas no fluxo de eventos em tempo real distribuído pelo sistema.

---

# 8. Segurança de Arquivos

Arquivos anexados são armazenados em um sistema externo de armazenamento.

Boas práticas incluem:

* URLs assinadas
* controle de acesso
* validação de tipo de arquivo
* limite de tamanho

O fluxo de upload adota **Pre-signed URLs**:
1. O cliente solicita um upload à **API Service**.
2. A API valida permissões e solicita ao **MinIO** uma URL temporária assinada.
3. O cliente faz o `PUT` do arquivo diretamente para o **MinIO**.
4. O MinIO notifica a API (ou o cliente avisa) para confirmar a persistência do metadado no Postgres.

---

# 9. Segurança de Comunicação

A comunicação entre cliente e servidor deve ocorrer através de conexões seguras.

Exemplo:

```
HTTPS
WSS (WebSocket Secure)
```

---

# 10. Observabilidade e Logs

Logs são utilizados para monitoramento do sistema.

Os logs devem:

* evitar dados sensíveis
* evitar conteúdo de mensagens
* registrar apenas eventos técnicos

Ferramentas de observabilidade podem incluir:

* métricas (Prometheus)
* dashboards (Grafana)
* agregação de logs (ELK Stack)

---

# 11. Resumo

A arquitetura de segurança do SlickChat foi projetada para:

* preservar anonimato dos usuários
* proteger credenciais
* prevenir abuso do sistema
* garantir comunicação segura
