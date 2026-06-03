/** Deve coincidir com `services/gateway/internal/ws/delivery.go` */
export const MAX_MESSAGE_CONTENT_LENGTH = 2000;

export function clampMessageContent(content: string): string {
  if (content.length <= MAX_MESSAGE_CONTENT_LENGTH) {
    return content;
  }
  return `${content.slice(0, MAX_MESSAGE_CONTENT_LENGTH)}…`;
}
