export function createWsClient(url: string): WebSocket {
  return new WebSocket(url);
}
