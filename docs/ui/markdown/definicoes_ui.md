# SlickChat — Definição de Experiência do Usuário (UX) e Interface (UI)

## 1. Introdução

Este documento detalha a experiência do usuário (UX) e a modelagem da interface do usuário (UI) para o sistema SlickChat. Ele estabelece diretrizes e princípios para garantir que a interface seja consistente, intuitiva e alinhada com os valores fundamentais do produto: anonimato, privacidade e efemeridade. O objetivo é traduzir os conceitos técnicos subjacentes em uma interação fluida e compreensível para o usuário final.

---

# 2. Posicionamento do Produto

## 2.1 Modelo Mental do Usuário

O SlickChat deve ser percebido como:

> "Um Telegram anônimo com mensagens efêmeras configuráveis por sala."

Características percebidas:

*   Não exige dados pessoais
*   Identidade pseudônima (username#xxxx)
*   Comunicação em tempo real
*   Controle de privacidade por contexto (sala)

---

## 2.2 Princípios de UX

1.  **Simplicidade Primeiro**

    *   O usuário não deve entender Kafka, TTL ou arquitetura.
    *   Apenas "enviar mensagem".

2.  **Privacidade Invisível**

    *   Segurança forte sem fricção.

3.  **Efemeridade Contextual**

    *   O comportamento da mensagem depende da sala.

4.  **Progressive Disclosure**

    *   Funcionalidades avançadas só aparecem quando necessário.

---

# 3. Modelo de Identidade

## 3.1 Exibição

Formato padrão:

```
username#1234
```

Regras:

*   Sempre exibir completo em contextos formais.
*   Pode ocultar parcialmente em UI compacta.

Exemplo:

*   Lista: `shadow#****`
*   Chat aberto: `shadow#1827`

---

## 3.2 Ações Relacionadas à Identidade

*   Copiar ID.
*   Rotacionar discriminator (#xxxx).
*   Visualizar perfil mínimo.

⚠️ Não existe:

*   Foto obrigatória.
*   Bio obrigatória.
*   Dados pessoais.

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

## 4.2 UX dos Modos

Os modos NÃO devem ser complexos.

### Estratégia:

*   Mostrar como **"tipo da sala"**.
*   Não como configuração técnica.

Exemplo:

*   🟢 Sala normal
*   ⏳ Sala temporária
*   🔒 Sala privada efêmera

---

# 5. Onboarding

## 5.1 Criação de Conta

Fluxo:

1.  Escolher username.
2.  Definir senha OU modo paranoico.
3.  Exibir:

    *   `username#xxxx`
    *   `recovery key`

⚠️ Ponto crítico:

*   Recovery key exibida UMA vez.
*   UI deve forçar o usuário a salvar.

---

## 5.2 Primeiro Acesso

Estado inicial:

*   Sem chats.
*   Sugestão de:

    *   iniciar conversa.
    *   entrar em sala.

---

# 6. Estrutura da Interface

## 6.1 Layout Principal

O layout da aplicação seguirá um modelo responsivo e adaptável, inspirado em aplicações de chat modernas, para garantir uma experiência consistente em diferentes tamanhos de tela.

### Sidebar (Navegação Esquerda)

Responsável pela navegação principal e exibição de listas de contextos de comunicação.

*   **Conteúdo:** Lista de conversas privadas, lista de salas (públicas, privadas, temporárias).
*   **Ações:** Botões claros para "Nova Conversa" e "Nova Sala".
*   **Responsividade:** Deve ser recolhível em telas menores (ex: mobile) e sempre visível em telas maiores (ex: desktop/tablet).

### Área Central (Conteúdo Principal)

Exibe o conteúdo interativo da conversa ou sala selecionada.

*   **Conteúdo:** Área de exibição de mensagens, incluindo texto, anexos e indicadores de status.
*   **Interação:** Campo de input de texto para envio de mensagens, botões para anexar arquivos, gravar áudio, etc.
*   **Adaptação:** Ocupa a maior parte do espaço disponível, ajustando-se dinamicamente conforme a sidebar e o painel lateral.

### Painel Lateral (Direita - Opcional)

Fornece informações contextuais e funcionalidades adicionais, podendo ser expandido ou recolhido conforme a necessidade.

*   **Conteúdo:** Informações detalhadas da sala/conversa, configurações específicas (ex: TTL da sala), lista de participantes, opções de moderação.
*   **Comportamento:** Ativado sob demanda, podendo ser um overlay em telas menores para evitar sobrecarga visual.

## 6.2 Componentes de Interface Comuns

Para garantir consistência e agilidade no desenvolvimento, serão utilizados componentes de interface padronizados. A prioridade é a reutilização e a clareza.

*   **Botões:** Padrões para ações primárias, secundárias, destrutivas e ícones.
*   **Campos de Input:** Para texto, senhas, pesquisa, com validação visual e feedback.
*   **Cards/Itens de Lista:** Para exibir conversas, mensagens, usuários, com estados de hover, selecionado e inativo.
*   **Modais/Diálogos:** Para confirmações, formulários complexos e informações críticas (ex: exibição da Recovery Key).
*   **Notificações/Toasts:** Para feedback rápido e não-intrusivo de ações do sistema.
*   **Avatares/Identidades:** Representação padronizada do `username#xxxx`.
*   **Loaders/Spinners:** Indicadores visuais de carregamento de conteúdo ou processamento de ações.

---

# 7. Fluxos Principais

## 7.1 Chat Privado

Fluxo:

1.  Buscar usuário (`username#xxxx`).
2.  Iniciar conversa.
3.  Enviar mensagem.

---

## 7.2 Salas

### Criar Sala

Inputs:

*   Nome
*   Tipo de persistência
*   TTL (opcional)

---

### Entrar em Sala

*   Pública: direto.
*   Privada: convite.

---

## 7.3 Envio de Mensagem

Ações:

*   Texto
*   Anexo
*   Áudio

Feedback:

*   Enviada
*   Entregue
*   Lida

---

# 8. Representação de Efemeridade

Esta seção detalha como a efemeridade das mensagens e salas será comunicada e gerenciada na interface do usuário.

## 8.1 Mensagens com TTL (Time To Live)

Mensagens com TTL possuem um tempo de vida definido. A UI deve comunicar claramente o estado e a expiração.

*   **Contador Visível:** Exibir um contador regressivo (`countdown`) próximo à mensagem, mostrando o tempo restante para expiração (ex: "expira em 10s", "expira em 1m", "expira em 5min").
*   **Animação/Feedback Visual:** Quando uma mensagem está prestes a expirar (ex: nos últimos 10 segundos), o contador pode mudar de cor, piscar ou ter uma animação sutil para chamar a atenção. Após a expiração, a mensagem deve desaparecer suavemente (fade-out) da tela.
*   **Sincronização:** A expiração deve ser consistente entre todos os clientes que veem a mensagem. O contador regressivo deve refletir o tempo real de expiração, mesmo se o cliente se reconectar ou abrir a sala mais tarde.
*   **Edge Case: Mensagem Visível Expirando:** Se um usuário está visualizando uma mensagem e ela expira, a mensagem deve desaparecer imediatamente da sua tela, com um feedback visual claro.
*   **Edge Case: Usuário Abre Sala com Mensagens Expirando:** Mensagens que expiraram antes de serem visualizadas não devem ser exibidas. Mensagens ativas devem aparecer com o contador atualizado.
*   **Edge Case: Mensagem Auto-Destrutiva:** Uma mensagem configurada para "destruir após leitura" deve desaparecer da interface do leitor imediatamente após ser visualizada, sem deixar rastros. Um feedback sutil (ex: um pequeno ícone de "lida e destruída") pode ser exibido para o remetente por um breve período.

## 8.2 Zero Logging Mode (Visão Geral da UX)

O Zero Logging Mode impede o armazenamento permanente de mensagens. A UI deve comunicar esta condição de forma inconfundível. Esta seção serve como uma introdução, com detalhes mais profundos na Seção 11.

*   **Indicador Visual Forte na Sala:** Exibir um indicador visual proeminente no cabeçalho da sala (ex: um cadeado aberto com uma nuvem riscada ou um banner com texto) que diz: "Esta sala não armazena mensagens" ou "Modo Efêmero Ativo: Mensagens não são salvas".
*   **Comunicação no Input:** No campo de digitação de mensagens, o placeholder pode mudar para "Enviar mensagem (não será salva)".

---

# 9. Modo Paranoico

## 9.1 Definição

*   Ativado apenas na criação da conta.

---

## 9.2 UX

*   Aviso claro: irreversível.
*   Sem recuperação de conta.

---

# 10. Comportamento em Tempo Real

Esta seção descreve o comportamento esperado da interface em cenários de comunicação em tempo real, incluindo gerenciamento de conexão e feedback de latência.

## 10.1 Estados de Conexão

A interface deve comunicar claramente o estado da conexão do usuário.

*   **Conectando (Connecting):** Indicador visual (ex: spinner sutil no canto superior) enquanto o sistema tenta estabelecer a conexão inicial.
*   **Conectado (Connected):** Estado normal de operação, geralmente sem indicador explícito, mas com feedback rápido para ações.
*   **Reconectando (Reconnecting):** Indicador visual claro (ex: banner na parte superior da tela ou ícone animado) quando a conexão é perdida e o sistema tenta reestabelecê-la automaticamente. Mensagem como "Conexão perdida. Tentando reconectar..." pode ser exibida.
*   **Offline:** Indicador de que o sistema está sem conexão. Nenhuma mensagem pode ser enviada/recebida. Mensagem como "Você está offline. Verifique sua conexão."

## 10.2 Reconexão Automática

Em caso de queda de conexão, o sistema deve tentar reconectar automaticamente.

*   **Transparência:** O usuário deve ser informado que uma reconexão está em andamento.
*   **Persistência da Sessão:** Após a reconexão, a sessão do usuário deve ser mantida, e ele deve retornar à sala ou conversa anterior.
*   **Sincronização:** Mensagens enviadas ou recebidas durante o período offline devem ser sincronizadas assim que a conexão for reestabelecida.

## 10.3 Ordem de Mensagens e Latência

Garantir que as mensagens sejam exibidas na ordem correta e que o usuário receba feedback sobre o status de envio.

*   **Mensagens "Enviando":** Ao enviar uma mensagem, ela deve aparecer imediatamente na interface do remetente com um status de "Enviando" (ex: texto em itálico, ícone de relógio, cor levemente diferente).
*   **Feedback de Latência:** O status "Enviando" deve permanecer até que o servidor confirme o recebimento (ACK). Caso a confirmação demore, um ícone de "retentando" ou "falha" pode ser exibido.
*   **Ordem Consistente:** Mensagens devem sempre ser exibidas em ordem cronológica de recebimento, mesmo que pacotes cheguem fora de ordem. O cliente deve ser capaz de reordenar ou aguardar.

## 10.4 Presença de Usuário

*   **Online:** Indicado por um círculo verde ou ícone similar ao lado do `username#xxxx`.
*   **Offline:** Indicado por um círculo cinza ou ausência de indicador ativo.
*   **Invisível:** O usuário aparece como offline para os demais, mesmo estando online. Não deve haver um indicador explícito de "invisível", pois isso anularia a funcionalidade.

---

# 11. Zero Logging Mode (UX Detalhado)

Detalha a experiência do usuário em salas com Zero Logging, onde nenhuma mensagem é persistida.

*   **Comportamento ao Recarregar a Página:** Se um usuário recarregar a página ou fechar o aplicativo e reabri-lo em uma sala com Zero Logging, o histórico de mensagens *não* será carregado. A tela de chat estará vazia, refletindo a natureza não persistente do modo.
*   **Comportamento ao Reconectar:** Similarmente, após uma reconexão (ex: devido a uma queda temporária de internet), apenas as novas mensagens (aquelas enviadas e recebidas a partir do momento da reconexão) aparecerão. O histórico pré-desconexão será perdido.
*   **Comunicação Clara sobre a Ausência de Histórico:**
    *   **Banner Contínuo:** Um banner não-intrusivo no topo da sala ou um indicador permanente próximo ao nome da sala deve lembrar constantemente o usuário que "As mensagens desta sala não são salvas."
    *   **Pop-up de Confirmação:** Ao criar ou entrar em uma sala com Zero Logging, um pop-up de confirmação com texto claro como "Você está entrando em uma sala com Zero Logging. Nenhuma mensagem será salva e seu histórico será perdido ao sair ou recarregar a página. Deseja continuar?" deve ser exibido.
    *   **Tooltips e Ajuda:** Tooltips informativos em ícones ou configurações relacionadas ao modo Zero Logging.

---

# 12. Notificações Web

O sistema deve fornecer notificações para alertar o usuário sobre novas mensagens ou eventos importantes, adaptando-se ao estado de atividade da aba.

## 12.1 Notificações In-App (Dentro da Aplicação)

*   **Indicadores Visuais:**
    *   **Contador de Mensagens Não Lidas:** Um número visível no ícone da conversa/sala na sidebar para indicar novas mensagens.
    *   **Destaque da Conversa:** A entrada da conversa/sala na sidebar deve ser visualmente destacada (ex: negrito, cor diferente) quando há mensagens não lidas.
    *   **Badge no Título da Página:** O título da aba do navegador deve mostrar um número (`(X) SlickChat`) para indicar novas notificações quando a aba não está ativa.
*   **Toasts/Snackbars:** Para feedback rápido de ações (ex: "Mensagem enviada", "Usuário desconectado"), exibidos brevemente na parte inferior ou superior da tela.

## 12.2 Notificações do Navegador (Web Push Notifications)

O sistema deve solicitar permissão para enviar notificações do navegador, usadas para alertar o usuário mesmo quando a aba do SlickChat não está em foco.

*   **Conteúdo:** As notificações devem ser concisas e preservar a privacidade (ex: "Nova mensagem em [Nome da Sala]" ou "Mensagem de [username#xxxx]").
*   **Ação:** Clicar na notificação deve abrir ou focar a aba do SlickChat na conversa/sala correspondente.

## 12.3 Regras de Notificação

Para evitar sobrecarga de notificações, regras claras devem ser aplicadas.

*   **Aba Ativa e Focada:**
    *   **Na Sala/Conversa Atual:** Nenhuma notificação do navegador ou in-app (exceto talvez um som sutil) deve ser disparada se o usuário está ativamente na conversa onde a mensagem foi recebida.
    *   **Em Outra Sala/Conversa Ativa:** Notificações in-app (contadores na sidebar) e badge no título da página devem ser ativadas. Notificações do navegador *podem* ser enviadas, mas com frequência limitada para evitar spam.
*   **Aba em Background (Não Focada):**
    *   **Qualquer Nova Mensagem:** Notificações in-app (contadores, destaques), badge no título da página e notificações do navegador devem ser disparadas para *todas* as novas mensagens.
*   **Modo Silencioso/Não Perturbe:** O usuário deve ter a opção de desativar todas as notificações sonoras ou visuais por um período.
*   **Prioridade:** Mensagens diretas podem ter prioridade maior de notificação do que mensagens em salas públicas.

---

# 13. Casos de Uso Críticos (Edge Cases)

Esta seção aborda como a UI deve reagir a cenários menos comuns, mas críticos, que impactam a experiência do usuário.

*   **Mensagem Enviada, mas Conexão Cai Antes da Confirmação:**
    *   A mensagem deve permanecer na UI do remetente com um status "Enviando" ou "Falha no Envio", com um botão de "Tentar Novamente" ou um ícone de retry.
    *   Um alerta temporário deve informar "Não foi possível enviar a mensagem. Tentando novamente..."
    *   Se a reconexão for bem-sucedida, o sistema deve tentar re-enviar a mensagem.
*   **Mensagem Duplicada:** O sistema deve ter mecanismos para detectar e exibir apenas uma instância de uma mensagem, mesmo que ela seja recebida múltiplas vezes devido a retries de rede. No caso improvável de uma duplicidade, a UI deve apresentar apenas uma cópia, sem feedback de erro ao usuário, para manter a fluidez.
*   **Usuário Offline Recebendo Mensagens:**
    *   Mensagens enviadas para um usuário offline devem ser marcadas como "enviadas" (pelo remetente), mas não como "entregues" ou "lidas".
    *   Ao reconectar, as mensagens pendentes devem ser entregues e aparecer na UI do destinatário em ordem cronológica.
*   **Sala Deletada Enquanto Usuário Está Dentro:**
    *   A UI do usuário deve ser atualizada imediatamente, exibindo uma notificação clara como "Esta sala foi encerrada." ou "Você foi removido da sala."
    *   O usuário deve ser redirecionado para a lista de salas ou para sua tela inicial.
    *   O campo de input da sala deletada deve ser desativado.
*   **TTL Expirando Durante a Leitura (Mensagem Auto-Destrutiva):**
    *   Para mensagens com TTL que também são auto-destrutivas, a prioridade é a destruição após a leitura. Se o TTL expirar *antes* da leitura, a mensagem desaparece pelo TTL. Se for lida *antes* do TTL, desaparece pela regra de auto-destruição.
    *   O feedback visual deve ser o desaparecimento imediato da mensagem após a leitura, sem o contador regressivo de TTL se a leitura for o gatilho principal.

---

# 14. Performance Percebida

Estratégias para melhorar a percepção de velocidade e responsividade da interface.

*   **Skeleton Loading:** Em vez de exibir telas vazias ou spinners genéricos, usar "esqueletos" de conteúdo (placeholders em formato de texto/imagem) ao carregar listas de conversas, mensagens ou detalhes de perfil. Isso dá a impressão de que o conteúdo está "a caminho".
*   **Paginação / Lazy Loading de Mensagens:**
    *   **Scroll Infinito:** Em chats e salas, as mensagens devem ser carregadas em blocos à medida que o usuário rola para cima (histórico) ou para baixo (novas mensagens).
    *   **Indicadores de Carregamento:** Um spinner sutil deve ser exibido no topo da área de mensagens enquanto o histórico é carregado.
*   **Respostas Otimistas:** Ao enviar uma mensagem, exibi-la imediatamente na interface do remetente (com status "Enviando") antes de receber a confirmação do servidor. Isso reduz a percepção de latência.
*   **Microinterações:** Animações sutis e feedback visual rápido para cliques em botões ou interações pequenas para indicar que a ação foi registrada pelo sistema.

---

# 15. Moderação Básica (UX)

Apesar do foco em anonimato, a moderação é crucial. A UI deve facilitar essas ações.

*   **Bloquear Usuário (Chat Direto):**
    *   Opção clara de "Bloquear Usuário" no perfil do usuário ou menu contextual da conversa.
    *   Confirmação ao bloquear: "Você deseja bloquear [username#xxxx]? Você não receberá mais mensagens dele e ele não poderá entrar em suas salas privadas."
    *   Feedback visual de que o usuário está bloqueado.
*   **Silenciar Sala:**
    *   Opção "Silenciar Notificações" nas configurações da sala.
    *   Um ícone de sino riscado pode aparecer na entrada da sala na sidebar.
*   **Rate Limit Visual:**
    *   Se o usuário atingir um limite de rate limit (ex: mensagens muito rápidas), o campo de input de texto pode ficar temporariamente desativado ou exibir uma mensagem como "Você está enviando mensagens muito rápido. Aguarde X segundos."
    *   Um indicador visual (ex: barra de progresso) pode mostrar quando o usuário pode enviar novamente.

---

# 16. Erros e Feedback

## 16.1 Tipos

O sistema deve prever e comunicar diversos tipos de erros ao usuário.

*   **Falha de conexão:** Erros relacionados à interrupção da conexão com o servidor.
*   **Rate limit:** Indicação de que o usuário excedeu o limite de ações permitidas.
*   **Permissão negada:** Ações que o usuário não tem autorização para realizar.
*   **Validação de formulário:** Mensagens específicas para campos inválidos.
*   **Erros de servidor:** Problemas internos que impedem a conclusão de uma operação.

## 16.2 Estratégia de Comunicação

O feedback deve ser imediato, claro e útil.

*   **Mensagens claras e concisas:** Evitar jargões técnicos. Ex: "Você não tem permissão para fazer isso." em vez de "HTTP 403 Forbidden".
*   **Orientação para solução:** Se possível, informar ao usuário como resolver o problema. Ex: "Sua conexão caiu. Tentando reconectar..."
*   **Localização do Feedback:**
    *   **Contextual:** Mensagens de erro de validação diretamente nos campos de formulário.
    *   **Global (Toast/Snackbar):** Para erros não críticos ou de sistema, exibidos temporariamente no canto da tela.
    *   **Bloqueador (Modal):** Para erros críticos que exigem a interação do usuário ou impedem o prosseguimento da tarefa.
*   **Consistência Visual:** Utilizar cores e ícones padronizados para feedback (ex: vermelho para erro, amarelo para aviso, verde para sucesso).

---

# 17. Decisões de Design Críticas

## 17.1 O que NÃO Expor

*   Termos técnicos de infraestrutura (ex: Kafka, Redis, Eventos). A complexidade técnica deve ser abstraída e traduzida em funcionalidades de UX simples.

---

## 17.2 O que Destacar

*   **Privacidade:** Reforçar visualmente que o usuário está em um ambiente seguro e anônimo.
*   **Efemeridade:** Deixar claro o tempo de vida das mensagens e salas.
*   **Simplicidade:** Manter a interface limpa e intuitiva, com o mínimo de distrações.

---

# 18. Diretrizes de Design Visual e Usabilidade

Esta seção estabelece as diretrizes visuais e de usabilidade para o SlickChat, garantindo uma experiência coesa, moderna e acessível.

## 18.1 Responsividade

O sistema deve ser totalmente responsivo, adaptando-se a diferentes tamanhos de tela (desktop, tablet, mobile) para garantir usabilidade e consistência em qualquer dispositivo.

*   **Grid System:** Utilização de um sistema de grid flexível (ex: classes `col-md-x`, `col-lg-x` do Bootstrap) para organização de conteúdo.
*   **Breakpoints:** Priorizar os breakpoints padrão de frameworks como Bootstrap (sm, md, lg, xl) para garantir a adaptação do layout.
*   **Elementos Adaptáveis:** Componentes de navegação (sidebar), painéis e inputs devem ajustar seu comportamento ou serem recolhidos/expandidos em telas menores.
*   **Touch-Friendly:** Elementos interativos devem ser dimensionados para facilitar a interação por toque em dispositivos móveis.

## 18.2 Tipografia

A escolha da tipografia visa clareza, legibilidade e um toque moderno.

*   **Font-Family:** Será utilizada uma família de fontes sem serifa (sans-serif) de fácil leitura (ex: `Roboto`, `Inter`, `Open Sans`).
*   **Hierarquia Visual:** Tamanhos e pesos de fonte (ex: `h1` a `h6`, `text-muted` do Bootstrap) bem definidos para cabeçalhos, corpo de texto e informações auxiliares.
*   **Legibilidade:** Contraste adequado entre texto e fundo, garantindo acessibilidade.

## 18.3 Paleta de Cores

A paleta de cores primária será baseada em um tema escuro para reforçar a ideia de privacidade e foco, com cores de destaque para ações e estados.

*   **Tema Padrão:** Escuro (dark mode) por padrão.
    *   **Fundo:** Tons de cinza escuro ou preto (`#121212`, `#1e1e1e`).
    *   **Texto Principal:** Branco ou cinza claro (`#ffffff`, `#e0e0e0`).
    *   **Texto Secundário:** Tons de cinza para informações menos importantes (`#aaaaaa`, `#888888`).
*   **Cores de Destaque:** Cores vibrantes, mas suaves, para elementos interativos e indicadores de status.
    *   **Primária (Ações):** Azul ou Verde-água (ex: `#23A2D9` - inspirado na cor dos containers C4) para botões, links, ícones ativos.
    *   **Sucesso:** Verde (`#4CAF50`).
    *   **Aviso:** Amarelo (`#FFC107`).
    *   **Erro:** Vermelho (`#F44336`).

## 18.4 Ícones

Os ícones devem ser simples, minimalistas e reconhecíveis universalmente.

*   **Estilo:** Ícones lineares (outline) ou sólidos, de fácil compreensão.
*   **Biblioteca:** Utilização de uma biblioteca de ícones (ex: Font Awesome, Material Icons, ou ícones SVG personalizados) para consistência.
*   **Uso:** Complementar texto, indicar ações, ilustrar estados.

## 18.5 Espaçamento e Alinhamento

Um design limpo é fundamental para a usabilidade.

*   **Grid de 8px:** Utilização de um grid de espaçamento de 8px (ou múltiplo) para margens, paddings e tamanhos de componentes, promovendo harmonia visual.
*   **Alinhamento:** Elementos de interface alinhados consistentemente para criar ordem e facilitar a varredura visual.
*   **Clareza:** Espaço em branco generoso para reduzir a sobrecarga cognitiva e destacar o conteúdo principal.

## 18.6 Elementos Interativos e Feedback

Todos os elementos interativos devem fornecer feedback visual claro ao usuário.

*   **Estados:** Hover, focus, active, disabled para botões, links, campos de input.
*   **Animações:** Transições sutis e microinterações para indicar mudanças de estado e melhorar a fluidez da experiência.
*   **Confirmações:** Ações potencialmente destrutivas devem requerer confirmação explícita do usuário.

## 18.7 Acessibilidade

O design deve ser acessível a um público amplo.

*   **Contraste:** Cores de texto e fundo devem ter contraste suficiente (WCAG AA).
*   **Navegação por Teclado:** Todos os elementos interativos devem ser navegáveis via teclado.
*   **ARIA Labels:** Utilização de atributos ARIA para descrever elementos da UI para leitores de tela.
*   **Redução de Movimento:** Considerar opções para usuários com sensibilidade a movimentos.

## 18.8 Padrões de Componentes

Para a implementação, serão adotados padrões de componentes que facilitam a reutilização e manutenção. O uso de um framework UI (como Bootstrap ou um design system customizado) é recomendado para acelerar o desenvolvimento e garantir a consistência.

*   **Botões:** `btn btn-primary`, `btn btn-secondary`, `btn btn-danger`, `btn btn-link`.
*   **Alertas:** `alert alert-success`, `alert alert-warning`, `alert alert-danger`, `alert alert-info`.
*   **Formulários:** `form-control`, `form-label`, `is-invalid`, `valid-feedback`, `invalid-feedback`.
*   **Layout:** `container`, `row`, `col-*-*`, `d-flex`, `justify-content-*`, `align-items-*`.
*   **Navegação:** `nav`, `nav-item`, `nav-link`, `dropdown`.

