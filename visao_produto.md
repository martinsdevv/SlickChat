# SlickChat — Definição do Produto e Escopo Funcional

## 1. Visão do Produto

**Nome do sistema:** SlickChat

SlickChat é uma plataforma de comunicação em tempo real projetada com foco em anonimato, privacidade e mensagens efêmeras. O sistema permite que usuários interajam por meio de chats privados ou salas de conversa sem fornecer dados pessoais, utilizando apenas identidades pseudônimas no formato `username#xxxx`.

O objetivo principal é oferecer comunicação rápida e segura preservando o anonimato do usuário por design.

---

# 2. Princípios do Sistema

O SlickChat adota os seguintes princípios fundamentais:

## 2.1 Privacidade por design

O sistema não coleta:

* email
* telefone
* nome real
* data de nascimento

Dados pessoais não são necessários para criação de conta.

---

## 2.2 Identidade pseudônima

Cada usuário possui um identificador único:

```
username#xxxx
```

Exemplo:

```
nova#1827
shadow#9021
pixel#4410
```

O número é gerado automaticamente para evitar colisões de nome.

---

## 2.3 Controle de persistência

Usuários e salas podem controlar por quanto tempo mensagens existem no sistema.

O SlickChat suporta três modos de persistência:

* **Persistência padrão** — mensagens são armazenadas permanentemente.
* **TTL (Time To Live)** — mensagens expiram automaticamente após um período definido.
* **Zero Logging Mode** — mensagens não são armazenadas e existem apenas no fluxo em tempo real.

---

## 2.4 Comunicação em tempo real

Todas as mensagens são transmitidas instantaneamente utilizando comunicação persistente entre cliente e servidor através de **WebSockets**.

---

## 2.5 Anonimato como requisito arquitetural

O sistema foi projetado para minimizar coleta e armazenamento de dados identificáveis. Informações de autenticação são limitadas ao mínimo necessário para permitir acesso seguro à conta.

---

# 3. Sistema de Identidade e Autenticação

## 3.1 Identidade do usuário

Cada usuário possui um identificador único no formato:

```
username#xxxx
```

Onde:

* `username` é escolhido pelo usuário
* `#xxxx` é um identificador numérico gerado automaticamente

Isso permite que múltiplos usuários utilizem o mesmo username sem colisões.

Exemplo:

```
shadow#1827
shadow#4410
shadow#9021
```

---

## 3.2 Autenticação padrão

O modelo padrão de autenticação utiliza:

```
username#xxxx
password
recovery_key
```

### Cadastro

Durante o cadastro:

* o usuário escolhe um `username`
* define uma `senha`
* o sistema gera automaticamente:

```
username#xxxx
recovery_key
```

A recovery key é exibida **uma única vez** e deve ser armazenada pelo usuário.

---

### Login

Login padrão:

```
username#xxxx
password
```

---

### Recuperação de conta

Caso o usuário esqueça a senha, ele pode recuperar a conta usando:

```
username#xxxx
recovery_key
```

O sistema não utiliza email ou telefone para recuperação.

---

## 3.3 Modo paranoico de conta

O sistema oferece um modo opcional chamado **Paranoid Account Mode**.

Nesse modo:

* nenhuma senha é definida
* autenticação ocorre **exclusivamente por chave secreta**

Formato:

```
username#xxxx
identity_key
```

Exemplo:

```
shadow#8271
identity_key: SC-91HF-K2P8-AD7L-92QP
```

Características:

* a chave é gerada uma única vez
* **não pode ser recuperada**
* se o usuário perder a chave, a conta é perdida permanentemente

---

# 4. Funcionalidades do Sistema

## 4.1 Chat privado

Usuários podem iniciar conversas privadas com outros usuários.

Funcionalidades do chat privado:

* envio de mensagens
* envio de anexos
* envio de áudio
* mensagens temporárias
* status de entrega

---

## 4.2 Sistema de salas

O sistema possui três tipos principais de salas:

### Sala pública

* qualquer usuário pode entrar

### Sala privada

* entrada via convite

### Sala temporária

* possui tempo de vida configurável
* é destruída automaticamente após expiração

---

## 4.3 Mensagens com TTL

Mensagens podem possuir tempo de vida configurável.

Exemplos de TTL:

```
10 segundos
1 minuto
5 minutos
1 hora
1 dia
permanente
```

Após o tempo definido, a mensagem é removida automaticamente.

---

## 4.4 Mensagens auto destrutivas

Usuários podem enviar mensagens que desaparecem automaticamente após serem visualizadas.

Exemplo:

```
Mensagem expira 10 segundos após leitura
```

---

## 4.5 Anexos

O sistema permite envio de arquivos como:

* imagens
* vídeos
* arquivos
* áudio

---

## 4.6 Presença

Sistema de presença em tempo real:

* online
* offline
* invisível

---

## 4.7 Identidade rotativa

Usuário pode regenerar seu identificador numérico.

Exemplo:

```
shadow#9021 → shadow#4418
```

---

## 4.8 Modo paranoia de sala

Salas podem ativar um modo de privacidade máxima.

Características:

* histórico não é salvo
* mensagens não persistem

---

## 4.9 Zero Logging Mode

Quando ativado:

* mensagens não são armazenadas permanentemente
* sistema funciona apenas em streaming

---

## 4.10 Salas auto destrutivas

Salas podem definir um tempo máximo de existência.

Exemplo:

```
Sala expira em 30 minutos
```

Após expiração:

* sala é removida
* histórico é apagado

---

## 4.11 Moderação

Salas possuem administradores responsáveis pela moderação.

Permissões do administrador:

* remover usuários
* silenciar usuários
* excluir mensagens
* encerrar sala

---

## 4.12 Sistema de denúncia

Usuários podem denunciar mensagens ou usuários.

---

## 4.13 Controle anti-spam

O sistema implementa mecanismos de proteção contra abuso.

Inclui:

* rate limit
* flood control

---

# 5. Funcionalidades do MVP

O MVP (Minimum Viable Product) do SlickChat inclui:

* criação de identidade `username#xxxx`
* autenticação com senha
* recuperação com recovery key
* modo paranoico de conta (login por chave)
* chat privado
* salas públicas
* salas privadas
* salas temporárias
* mensagens em tempo real
* mensagens com TTL
* mensagens auto destrutivas
* envio de anexos
* envio de áudio
* presença online
* identidade rotativa
* modo paranoia
* zero logging mode
* moderação de salas
* sistema básico anti spam

---

# 6. Objetivo Técnico do Projeto

O projeto tem como objetivo demonstrar a implementação de um sistema de comunicação em tempo real utilizando arquitetura moderna baseada em eventos e mensageria.

A solução deverá incluir:

* comunicação em tempo real
* processamento assíncrono
* arquitetura escalável
* controle de persistência de dados

O sistema deverá ser capaz de suportar múltiplos usuários conectados simultaneamente e distribuir mensagens em tempo real.

---

# 7. Diferenciais do Projeto

O SlickChat apresenta os seguintes diferenciais técnicos:

* anonimato por design
* autenticação sem email ou telefone
* recuperação offline via recovery key
* modo paranoico com autenticação por chave única
* mensagens efêmeras
* salas auto destrutivas
* controle granular de persistência
* identidade rotativa
* arquitetura baseada em eventos
* comunicação em tempo real
