# SlickChat — UX & Interface Definition

## 1. Introdução

Este documento define a experiência do usuário (UX) e a modelagem de interface (UI) do SlickChat.

O objetivo é traduzir os conceitos técnicos (event-driven, anonimato, efemeridade) em uma experiência simples, intuitiva e consistente.

---

# 2. Posicionamento do Produto

## 2.1 Modelo mental do usuário

O SlickChat deve ser percebido como:

> "Um Telegram anônimo com mensagens efêmeras configuráveis por sala."

Características percebidas:

* Não exige dados pessoais
* Identidade pseudônima (username#xxxx)
* Comunicação em tempo real
* Controle de privacidade por contexto (sala)

---

## 2.2 Princípios de UX

1. **Simplicidade primeiro**

   * O usuário não deve entender Kafka, TTL ou arquitetura
   * Apenas "enviar mensagem"

2. **Privacidade invisível**

   * Segurança forte sem fricção

3. **Efemeridade contextual**

   * O comportamento da mensagem depende da sala

4. **Progressive disclosure**

   * Funcionalidades avançadas só aparecem quando necessário

---

# 3. Modelo de Identidade

## 3.1 Exibição

Formato padrão:

```
username#1234
```

Regras:

* Sempre exibir completo em contextos formais
* Pode ocultar parcialmente em UI compacta

Exemplo:

* Lista: `shadow#****`
* Chat aberto: `shadow#1827`

---

## 3.2 Ações relacionadas à identidade

* Copiar ID
* Rotacionar discriminator (#xxxx)
* Visualizar perfil mínimo

⚠️ Não existe:

* foto obrigatória
* bio obrigatória
* dados pessoais

---

# 4. Modos do Sistema

## 4.1 Definição

Os modos são configurados **por sala**.

### Modos disponíveis:

| Modo         | Comportamento                         |
| ------------ | ------------------------------------- |
| Persistente  | Mensagens salvas normalmente          |
| TTL          | Mensagens expiram após tempo definido |
| Zero Logging | Mensagens não são armazenadas         |

---

## 4.2 UX dos modos

Os modos NÃO devem ser complexos.

### Estratégia:

* Mostrar como **"tipo da sala"**
* Não como configuração técnica

Exemplo:

* 🟢 Sala normal
* ⏳ Sala temporária
* 🔒 Sala privada efêmera

---

# 5. Onboarding

## 5.1 Criação de conta

Fluxo:

1. Escolher username
2. Definir senha OU modo paranoico
3. Exibir:

   * username#xxxx
   * recovery key

⚠️ Ponto crítico:

* Recovery key exibida UMA vez
* UI deve forçar o usuário a salvar

---

## 5.2 Primeiro acesso

Estado inicial:

* Sem chats
* Sugestão de:

  * iniciar conversa
  * entrar em sala

---

# 6. Estrutura da Interface

## 6.1 Layout principal

Modelo baseado em apps de chat modernos:

### Sidebar (esquerda)

* Lista de conversas
* Lista de salas
* Botão "Nova conversa"
* Botão "Nova sala"

---

### Área central

* Mensagens
* Input de texto
* Anexos

---

### Painel lateral (direita - opcional)

* Informações da sala
* Configurações
* Participantes

---

# 7. Fluxos Principais

## 7.1 Chat privado

Fluxo:

1. Buscar usuário (username#xxxx)
2. Iniciar conversa
3. Enviar mensagem

---

## 7.2 Salas

### Criar sala

Inputs:

* Nome
* Tipo de persistência
* TTL (opcional)

---

### Entrar em sala

* Pública: direto
* Privada: convite

---

## 7.3 Envio de mensagem

Ações:

* Texto
* Anexo
* Áudio

Feedback:

* enviada
* entregue
* lida

---

# 8. Representação de Efemeridade

## 8.1 TTL

* Mostrar countdown
* Exemplo: "expira em 10s"

---

## 8.2 Zero Logging

* Indicador visual forte

Exemplo:

> "Esta sala não armazena mensagens"

---

# 9. Modo Paranoico

## 9.1 Definição

* Ativado apenas na criação da conta

---

## 9.2 UX

* Aviso claro: irreversível
* Sem recuperação de conta

---

# 10. Estados do Sistema

## 10.1 Presença

* online
* offline
* invisível

---

## 10.2 Mensagens

* enviando
* enviada
* entregue
* lida
* expirada

---

# 11. Erros e Feedback

## 11.1 Tipos

* Falha de conexão
* Rate limit
* Permissão negada

---

## 11.2 Estratégia

* Mensagens claras
* Sem termos técnicos

---

# 12. Decisões de Design Críticas

## 12.1 O que NÃO expor

* Kafka
* Redis
* Eventos

---

## 12.2 O que destacar

* Privacidade
* Efemeridade
* Simplicidade

---

# 13. Direção Visual

* Tema escuro por padrão
* Interface limpa
* Ícones simples

---