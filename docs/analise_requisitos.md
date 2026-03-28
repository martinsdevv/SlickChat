# SlickChat — Documento de Requisitos

## 1. Introdução

Este documento descreve os requisitos funcionais e não funcionais do sistema **SlickChat**, uma plataforma de comunicação em tempo real focada em anonimato, privacidade e mensagens efêmeras.

O objetivo deste documento é definir de forma clara o comportamento esperado do sistema antes da fase de modelagem de domínio e arquitetura.

---

# 2. Escopo do Sistema

O SlickChat permite que usuários se comuniquem em tempo real utilizando identidades pseudônimas no formato `username#xxxx`, sem a necessidade de fornecer dados pessoais.

O sistema suporta:

* chats privados
* salas públicas
* salas privadas
* salas temporárias
* mensagens efêmeras
* anexos
* moderação de salas
* mecanismos de proteção contra abuso

---

# 3. Requisitos Funcionais

## 3.1 Identidade e Autenticação

RF01 — O sistema deve permitir a criação de contas utilizando apenas um `username`.

RF02 — O sistema deve gerar automaticamente um identificador numérico no formato `username#xxxx`.

RF03 — O sistema deve permitir autenticação utilizando `username#xxxx` e senha.

RF04 — O sistema deve gerar uma **recovery key** no momento do cadastro do usuário.

RF05 — O sistema deve permitir recuperação de conta utilizando `recovery key`.

RF06 — O sistema deve permitir criação de contas utilizando **modo paranoico**.

RF07 — No modo paranoico, o sistema deve autenticar o usuário exclusivamente através de uma `identity key`.

RF08 — O sistema deve permitir rotação do identificador numérico (`#xxxx`).

---

## 3.2 Chat Privado

RF09 — O sistema deve permitir iniciar conversas privadas entre usuários.

RF10 — O sistema deve permitir envio de mensagens em conversas privadas.

RF11 — O sistema deve permitir envio de anexos em conversas privadas.

RF12 — O sistema deve indicar status de entrega das mensagens.

---

## 3.3 Sistema de Salas

RF13 — O sistema deve permitir criação de salas públicas.

RF14 — O sistema deve permitir criação de salas privadas.

RF15 — O sistema deve permitir criação de salas temporárias.

RF16 — O sistema deve permitir entrada em salas públicas.

RF17 — O sistema deve permitir entrada em salas privadas mediante convite.

RF18 — O sistema deve permitir configuração de tempo de vida da sala.

RF19 — O sistema deve remover automaticamente salas expiradas.

---

## 3.4 Mensagens

O sistema suporta três estratégias de persistência de mensagens:

* **Persistência padrão** — mensagens são armazenadas permanentemente.
* **TTL (Time To Live)** — mensagens são armazenadas temporariamente e removidas após expiração.
* **Zero Logging Mode** — mensagens não são armazenadas permanentemente e existem apenas no fluxo de eventos em tempo real.

RF20 — O sistema deve permitir envio de mensagens em tempo real.

RF21 — O sistema deve permitir configuração de **TTL (Time To Live)** para mensagens.

RF22 — O sistema deve remover automaticamente mensagens expiradas.

RF23 — O sistema deve permitir envio de mensagens auto destrutivas.

RF24 — O sistema deve remover mensagens auto destrutivas após leitura.

---

## 3.5 Anexos

RF25 — O sistema deve permitir envio de imagens.

RF26 — O sistema deve permitir envio de vídeos.

RF27 — O sistema deve permitir envio de arquivos.

RF28 — O sistema deve permitir envio de mensagens de áudio.

RF29 — O sistema deve validar tipo e tamanho de arquivos enviados como anexos.

---

## 3.6 Presença

RF30 — O sistema deve indicar quando usuários estão online.

RF31 — O sistema deve indicar quando usuários estão offline.

RF32 — O sistema deve permitir que usuários ativem modo invisível, ocultando seu status de presença.

---

## 3.7 Moderação

RF33 — O sistema deve permitir definir administradores de sala.

RF34 — O sistema deve permitir administradores remover usuários da sala.

RF35 — O sistema deve permitir administradores silenciar usuários.

RF36 — O sistema deve permitir administradores excluir mensagens.

RF37 — O sistema deve permitir administradores encerrar salas.

RF38 — O sistema deve permitir que administradores banam usuários de uma sala.

---

## 3.8 Segurança e Abuso

RF39 — O sistema deve implementar mecanismos de **rate limiting**.

RF40 — O sistema deve implementar mecanismos de **controle de flood**.

RF41 — O sistema deve permitir denúncia de mensagens.

RF42 — O sistema deve permitir denúncia de usuários.

---

## 3.9 Modos de Privacidade

RF43 — O sistema deve permitir ativação do **modo paranoia** em salas.

RF44 — O sistema não deve persistir mensagens em salas com modo paranoia ativo.

RF45 — O sistema deve permitir ativação do **Zero Logging Mode** em salas.

RF46 — No Zero Logging Mode, mensagens não devem ser persistidas em banco de dados e devem existir apenas no fluxo de eventos em tempo real.

---

## 3.10 Sessões

RF47 — O sistema deve criar uma sessão autenticada após login bem sucedido.

RF48 — O sistema deve permitir múltiplas sessões simultâneas por usuário.

RF49 — O sistema deve expirar sessões automaticamente após período de inatividade.

---

# 4. Requisitos Não Funcionais

## 4.1 Privacidade

RNF01 — O sistema não deve exigir dados pessoais para criação de conta.

RNF02 — O sistema deve minimizar o armazenamento de dados identificáveis.

RNF03 — O sistema deve permitir autenticação sem email ou telefone.

---

## 4.2 Comunicação em Tempo Real

RNF04 — O sistema deve suportar comunicação em tempo real entre clientes e servidor.

RNF05 — O sistema deve entregar mensagens com baixa latência.

---

## 4.3 Arquitetura

RNF06 — O sistema deve utilizar arquitetura baseada em eventos.

RNF07 — O sistema deve utilizar mensageria para processamento assíncrono.

RNF08 — O sistema deve suportar escalabilidade horizontal.

---

## 4.4 Segurança

RNF09 — O sistema deve armazenar senhas utilizando hashing seguro.

RNF10 — O sistema deve proteger endpoints contra abuso.

---

## 4.5 Disponibilidade

RNF11 — O sistema deve suportar múltiplos usuários conectados simultaneamente.

RNF12 — O sistema deve suportar reconexão automática de clientes.

---

## 4.6 Persistência

RNF13 — O sistema deve permitir expiração automática de mensagens.

RNF14 — O sistema deve permitir remoção automática de salas expiradas.

---

# 5. Resumo

Total de requisitos definidos neste documento:

* **49 requisitos funcionais**
* **14 requisitos não funcionais**
