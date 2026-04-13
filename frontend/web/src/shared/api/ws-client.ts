import { env } from "../config/env";
import type { WsInboundPayload, WsOutEvent } from "./types";

export type WsEnvelope<T = unknown> = {
  type: WsOutEvent;
  payload: T;
};

export type WsConnectionHandlers = {
  onOpen: () => void;
  onClose: () => void;
  onError: () => void;
  onMessage: (event: WsEnvelope) => void;
};

export type WsClient = {
  socket: WebSocket;
  send: <T extends keyof WsInboundPayload>(
    type: T,
    payload: WsInboundPayload[T],
  ) => void;
  close: () => void;
};

export function createWsClient(
  ticket: string,
  handlers: WsConnectionHandlers,
): WsClient {
  const normalizedWsBase = env.wsBaseUrl.startsWith("ws")
    ? env.wsBaseUrl
    : `${window.location.protocol === "https:" ? "wss:" : "ws:"}//${
        window.location.host
      }${env.wsBaseUrl}`;
  const url = new URL(`${normalizedWsBase}/socket`);
  url.searchParams.set("ticket", ticket);

  const socket = new WebSocket(url.toString());

  socket.addEventListener("open", handlers.onOpen);
  socket.addEventListener("close", handlers.onClose);
  socket.addEventListener("error", handlers.onError);
  socket.addEventListener("message", (event) => {
    try {
      const parsed = JSON.parse(event.data) as WsEnvelope;
      handlers.onMessage(parsed);
    } catch {
      handlers.onMessage({
        type: "error",
        payload: { code: "invalid_gateway_payload" },
      });
    }
  });

  return {
    socket,
    send: (type, payload) => {
      socket.send(JSON.stringify({ type, payload }));
    },
    close: () => {
      socket.close();
    },
  };
}
