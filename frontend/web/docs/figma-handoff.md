# Handoff Figma -> Frontend (SlickChat)

## Como enviar uma tela para implementação
Envie sempre:
1. URL do arquivo com `node-id` (frame/componente exato)
2. Nome da tela/componente
3. Estado (default, hover, active, disabled, erro, etc.)
4. Breakpoints (mobile/tablet/desktop), se houver variação

Formato preferido:
`https://www.figma.com/design/<fileKey>/<nome>?node-id=<nodeId>`

## Fluxo recomendado
1. Implementar layout base da página
2. Implementar componentes da tela
3. Ajustar tokens e estados visuais
4. Revisar responsividade e acessibilidade
