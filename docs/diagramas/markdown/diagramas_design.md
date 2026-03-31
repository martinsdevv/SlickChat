# SlickChat — Diagramas de Design

Este documento contém os principais diagramas de design do sistema **SlickChat** baseados no modelo de domínio definido em modelo_dominio.

---

# 1. Diagrama de Classes

![Diagrama de Classes](../png/diagrama_classes.png)
---

# 2. C4 Model — System Context

![Diagrama de Contexto](../png/diagrama_c4_context.png)
---

# 3. C4 Model — Container Diagram

![Diagrama de Container](../png/diagrama_c4_container.png)

Descrição:

* Clientes se conectam via **WebSocket Gateway** para comunicação em tempo real

* Eventos de chat são publicados em **Kafka**

* Workers consomem eventos para:

  * distribuir mensagens
  * persistir dados
  * aplicar TTL
  * executar moderação

* **Redis** é utilizado para presença, sessões e rate limiting

* Dados persistentes são armazenados em **Postgres**

* Arquivos e anexos são armazenados em **Media Storage**

---

# 4. Fluxo de Mensagem

![](../png/diagrama_fluxo_mensagem.png)

---

# 5. Fluxo de Expiração de Mensagens

![](../png/diagrama_fluxo_message_expired.png)

Mensagens com TTL são removidas fisicamente do banco de dados.