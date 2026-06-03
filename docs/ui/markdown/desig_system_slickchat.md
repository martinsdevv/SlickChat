# SlickChat Design System

**Produto:** SlickChat  
**Escopo:** Interface Web (React + Tailwind)  
**Aluno:** Gabriel Martins Torres <br>
**Matrícula:** 20241012001820 <br>
**Leitura rápida:** `v1` = fundamentos visuais | `v2` = componentes e estados em contexto real <br>

Este documento organiza as decisões de UI do SlickChat para manter consistência visual e facilitar evolução da interface.

## 1. Princípios

- **Privacidade primeiro:** decisões visuais devem reforçar confiança e discrição.
- **Simplicidade:** o usuário vê ações claras, não complexidade técnica.
- **Consistência:** padrões repetíveis têm prioridade sobre exceções.
- **Feedback:** toda ação deve gerar resposta visual imediata.
- **Acessibilidade:** foco visível, contraste e navegação por teclado são obrigatórios.

## 2. Fundamentos visuais (v1)

## 2.1 Cores

**Status:** parcialmente implementado (tokens existem e já são usados, mas ainda há `#hex` hardcoded em telas).

### Tokens base atuais

Fonte: `frontend/web/src/shared/styles/tokens.css`.

| Token | Valor |
| --- | --- |
| `--bg-0` | `#0A0A0A` |
| `--bg-1` | `#1A1A1A` |
| `--bg-2` | `#2D2D2D` |
| `--text-0` | `#FFFFFF` |
| `--text-1` | `#F0F0F0` |
| `--text-2` | `#A0A0A0` |
| `--text-3` | `#7F7F7F` |
| `--primary-500` | `#7A00FF` |
| `--primary-400` | `#A052FF` |
| `--primary-300` | `#B388FF` |
| `--primary-200` | `#C7A3FF` |
| `--success-500` | `#4CAF50` |
| `--warning-500` | `#FFC107` |
| `--danger-500` | `#F44336` |

### Mapeamento semântico de uso

| Papel | Token recomendado | Uso |
| --- | --- | --- |
| `surface/app` | `--bg-0` | fundo global |
| `surface/raised` | `--bg-1` | cards, painéis, formulários |
| `surface/emphasis` | `--bg-2` | destaque de seção |
| `text/primary` | `--text-0` | título e conteúdo principal |
| `text/secondary` | `--text-1` | subtítulo e destaque secundário |
| `text/muted` | `--text-2` | descrição e apoio |
| `text/subtle` | `--text-3` | metadata e placeholder |
| `action/primary` | `--primary-500` | CTA principal |
| `action/primary-hover` | `--primary-400` | hover/focus de CTA |
| `state/success` | `--success-500` | sucesso |
| `state/warning` | `--warning-500` | alerta |
| `state/danger` | `--danger-500` | erro e ação destrutiva |

## 2.2 Tipografia

**Status:** implementado.

- **Família:** `Inter, Roboto, Open Sans, system-ui, sans-serif`.

| Estilo | Tailwind | Peso | Line-height | Uso |
| --- | --- | --- | --- | --- |
| `display` | `text-4xl` a `text-5xl` | 600 | `leading-tight` | onboarding |
| `heading-1` | `text-2xl` a `text-3xl` | 600 | `leading-tight` | título de tela/sala |
| `heading-2` | `text-lg` a `text-xl` | 500/600 | `leading-snug` | títulos de seção |
| `body` | `text-base` | 400 | `leading-relaxed` | conteúdo principal |
| `support` | `text-sm` | 400/500 | `leading-normal` | ajuda e descrição |
| `meta` | `text-xs` | 400/500 | `leading-normal` | status e timestamp |

## 2.3 Espaçamento

**Status:** implementado via utilitários Tailwind (não há tokens `--space-*`).

| Classe Tailwind | Valor | Uso |
| --- | --- | --- |
| `p-1`, `m-1`, `gap-1` | 4px | ajuste fino |
| `p-2`, `m-2`, `gap-2` | 8px | espaçamento mínimo |
| `p-3`, `m-3`, `gap-3` | 12px | agrupamento compacto |
| `p-4`, `m-4`, `gap-4` | 16px | padrão |
| `p-5`, `m-5`, `gap-5` | 20px | separação moderada |
| `p-6`, `m-6`, `gap-6` | 24px | bloco de formulário |
| `p-8`, `m-8`, `gap-8` | 32px | separação estrutural |

Regras rápidas:
- controle interativo com altura mínima de 44px;
- formulários com padding entre 12px e 24px;
- chat com scroll isolado na lista e input fixo no rodapé.

## 2.4 Raio, sombra e camada

**Status:** diretriz (ainda não tokenizado no projeto).

| Categoria | Valor sugerido | Uso |
| --- | --- | --- |
| `radius-sm` | 8px | badges e controles pequenos |
| `radius-md` | 12px | inputs e botões padrão |
| `radius-lg` | 16px | cards e painéis |
| `radius-xl` | 20px+ | destaque |
| `shadow-1` | sombra sutil | elevação básica |
| `shadow-2` | sombra média | modal e foco |

Ordem de camada recomendada: base < dropdown < overlay < modal < toast.

## 3. Componentes principais (v2)

**Status:** modelagem alinhada ao que já aparece nas telas atuais.

## 3.1 Botão

- Variantes: `primary`, `secondary`, `ghost`, `danger`, `icon`.
- Uso no produto: CTA de autenticação, criação de sala, envio, ações rápidas.

## 3.2 Campo de entrada

- Tipos: texto, senha, número, busca, handle.
- Uso no produto: onboarding, criação de sala, adição de participante, composer.

## 3.3 Item de sala (lista lateral)

- Estrutura: nome, badges de tipo/zero logging, preview de última mensagem.
- Variações: padrão, ativo, hover, com não lidas.

## 3.4 Bolha de mensagem

- Estrutura: autor, conteúdo, metadata (hora/status), indicador de TTL quando houver.
- Variações: própria, terceiros, temporária, selecionada.

## 3.5 Composer de mensagem

- Estrutura: campo principal + ações auxiliares (anexo, áudio, enviar).
- Variações: padrão, zero logging, indisponível/offline.

## 3.6 Indicadores e painéis

- **Badge de conexão:** `connected`, `reconnecting/offline`.
- **Painel lateral da sala:** informações, membros e ações da sala.
- **Barra de seleção:** ações em massa para mensagens selecionadas.

## 4. Matriz de estados

Estados básicos exigidos: `default`, `hover`, `focus-visible`, `active`, `disabled`, `error` (quando aplicável).

| Componente | Estados básicos | Estados de domínio SlickChat |
| --- | --- | --- |
| Botão | default, hover, focus-visible, active, disabled, loading | confirmar/cancelar ação crítica |
| Campo de entrada | default, hover, focus, filled, disabled, error | validação de auth/sala |
| Item de sala | default, hover, focused, active | normal, temporary, zero-logging |
| Bolha de mensagem | default, selected | sending, sent, delivered, read, failed, ttl-warning, ttl-critical |
| Composer | default, focus, disabled | modo zero-logging, bloqueio por conexão |
| Badge de conexão | default | connecting, connected, reconnecting, offline |
| Barra de seleção | default, disabled | selection-off, selection-on, selection-on-with-items |

## 5. Responsividade e acessibilidade

### Responsividade

- **Mobile:** foco em uma área por vez (salas ou chat), painéis em overlay.
- **Tablet:** sidebar colapsável, painel auxiliar sob demanda.
- **Desktop:** sidebar + chat + painel opcional sem bloquear fluxo principal.

### Acessibilidade

- contraste AA para texto e controles;
- foco visível em elementos interativos;
- navegação por teclado;
- rótulos claros e erros com texto de apoio;
- evitar depender só de cor para comunicar estado.

## 6. Escopo por versão

- **v1:** fundamentos visuais (cores, tipografia, espaçamento e estados básicos).
- **v2:** componentes principais do SlickChat com variantes e estados em contexto real.

## 7. Governança

- Convenção de token semântico: `categoria/funcao/nivel`.
- Convenção de componente: `Componente/Variante/Estado`.
- Novo componente só entra quando o padrão se repetir em múltiplos contextos e variante não resolver bem.
