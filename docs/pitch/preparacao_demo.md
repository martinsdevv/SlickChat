# Preparação da demo (pitch ~5 min)

Execute **na véspera** ou **1 h antes** da apresentação.

## 1. Subir o ambiente

```bash
make infra-up
make run-api          # terminal 1
make run-gateway      # terminal 2
make run-fanout       # terminal 3
make run-persistence  # terminal 4
make run-ttl          # terminal 5 (opcional, para sala TEMPORARY)
cd frontend/web && npm run dev
```

## 2. Duas contas (obrigatório)

| Aba | Usuário sugerido | Papel na demo |
|-----|------------------|---------------|
| 1 | Conta principal (ex: `jinxxx#7297`) | Apresentador |
| 2 | Segunda conta anônima | Responde mensagens ao vivo |

Na conta **admin**, use **Info da sala → Adicionar participante** com o handle da segunda conta.

## 3. Três salas com conteúdo

Crie (ou use) estas salas e **envie 2–3 mensagens em cada** antes do pitch:

| Sala | Tipo | O que mostrar |
|------|------|----------------|
| `zero log` | PUBLIC + Zero logging | Banner amarelo; F5 apaga histórico |
| `temporaria` | TEMPORARY + TTL 60s | Countdown / blur na mensagem |
| `publica` | PUBLIC | Conversa “normal” com histórico |

**Copie o UUID** da sala pública e teste **Entrar com ID da sala** na segunda aba (sem ser admin).

## 4. Roteiro rápido (2 min de demo)

1. Abrir sala **zero log** → histórico visível → enviar mensagem → segunda aba responde.
2. **F5** na sala zero log → mensagens antigas sumiram (zero logging).
3. Sala **temporaria** → mensagem com TTL desaparecendo.
4. Mencionar Figma + roadmap (anexos, feed de salas) em **uma frase**.

## 5. Se algo falhar

- Badge **Conectado** deve estar verde; se não, reinicie gateway + fanout.
- Sem mensagens na lista lateral → envie pelo menos uma por sala.
- Segunda pessoa não entra → admin adiciona por `usuario#1234` ou **Entrar com ID**.
