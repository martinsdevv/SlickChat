import { create } from "zustand";
import { apiRequest } from "../../shared/api/http-client";
import type { MessageHistoryItem, WSTicketResponse } from "../../shared/api/types";
import { createWsClient, type WsClient, type WsEnvelope } from "../../shared/api/ws-client";
import { useSessionStore } from "../session/store";

type DeliveryStatus = "sending" | "sent" | "delivered" | "read" | "failed";
type PresenceStatus = "online" | "offline" | "unknown";
type ConnectionStatus = "idle" | "connecting" | "connected" | "reconnecting" | "offline";

export type ChatMessage = {
  id: string;
  roomId: string;
  authorId: string;
  authorHandle: string;
  content: string;
  status: DeliveryStatus;
  createdAt: string;
  isOwn: boolean;
  expiresAt: string | null;
  ttlSeconds: number | null;
  isZeroLogging: boolean;
  isTemporary: boolean;
};

type ChatState = {
  messagesByRoom: Record<string, ChatMessage[]>;
  pendingAckIds: string[];
  draftByRoom: Record<string, string>;
  connectionStatus: ConnectionStatus;
  connectionError: string | null;
  presenceByUserId: Record<string, PresenceStatus>;
  setDraft: (roomId: string, value: string) => void;
  connect: (token: string) => Promise<void>;
  disconnect: () => void;
  loadRoomHistory: (roomId: string, token: string) => Promise<void>;
  sendMessage: (roomId: string) => void;
  markMessageRead: (roomId: string, messageId: string) => void;
  markMessageDelivered: (roomId: string, messageId: string) => void;
};

let wsClient: WsClient | null = null;
let reconnectTimer: number | null = null;
let reconnectAttempts = 0;

function appendMessage(
  messagesByRoom: Record<string, ChatMessage[]>,
  roomId: string,
  message: ChatMessage,
) {
  const current = messagesByRoom[roomId] ?? [];
  if (current.some((item) => item.id === message.id)) {
    return messagesByRoom;
  }

  return {
    ...messagesByRoom,
    [roomId]: [...current, message].sort((a, b) =>
      a.createdAt.localeCompare(b.createdAt),
    ),
  };
}

function upsertOwnMessageFromServer(
  messagesByRoom: Record<string, ChatMessage[]>,
  messageFromServer: ChatMessage,
) {
  const roomMessages = messagesByRoom[messageFromServer.roomId] ?? [];
  if (roomMessages.some((item) => item.id === messageFromServer.id)) {
    return { messagesByRoom, reconciledLocalId: null as string | null };
  }

  const serverTimestamp = Date.parse(messageFromServer.createdAt);
  const optimisticIndex = roomMessages.findIndex((item) => {
    if (!item.isOwn || !item.id.startsWith("local-")) {
      return false;
    }
    if (item.content !== messageFromServer.content) {
      return false;
    }
    const localTimestamp = Date.parse(item.createdAt);
    return Math.abs(localTimestamp - serverTimestamp) <= 15000;
  });

  if (optimisticIndex === -1) {
    return {
      messagesByRoom: appendMessage(
        messagesByRoom,
        messageFromServer.roomId,
        messageFromServer,
      ),
      reconciledLocalId: null as string | null,
    };
  }

  const reconciledLocalId = roomMessages[optimisticIndex].id;
  const nextRoomMessages = roomMessages.map((item, index) =>
    index === optimisticIndex
      ? {
          ...messageFromServer,
          // Preserve advanced status if it already progressed locally.
          status:
            item.status === "delivered" || item.status === "read"
              ? item.status
              : messageFromServer.status,
        }
      : item,
  );

  return {
    messagesByRoom: {
      ...messagesByRoom,
      [messageFromServer.roomId]: nextRoomMessages.sort((a, b) =>
        a.createdAt.localeCompare(b.createdAt),
      ),
    },
    reconciledLocalId,
  };
}

function deriveConnectionError(event: WsEnvelope): string | null {
  if (event.type !== "error") {
    return null;
  }

  const payload = event.payload as { code?: string };
  return payload.code ?? "unknown_gateway_error";
}

function scheduleReconnect(token: string, connect: (token: string) => Promise<void>) {
  if (reconnectTimer !== null) {
    window.clearTimeout(reconnectTimer);
  }

  reconnectAttempts += 1;
  const waitMs = Math.min(1000 * 2 ** reconnectAttempts, 10_000);
  reconnectTimer = window.setTimeout(() => {
    void connect(token);
  }, waitMs);
}

function createMessageFromGateway(
  payload: Record<string, unknown>,
  currentUserId: string,
): ChatMessage | null {
  const roomId = String(payload.room_id ?? "");
  const messageId = String(payload.message_id ?? "");
  const senderId = String(payload.sender_id ?? "");
  const content = String(payload.content ?? "");
  if (!roomId || !messageId || !content) {
    return null;
  }

  const sentAtRaw = payload.sent_at;
  const sentAt = typeof sentAtRaw === "string" ? sentAtRaw : new Date().toISOString();
  const expiresAtRaw = payload.expires_at;
  const expiresAt = typeof expiresAtRaw === "string" ? expiresAtRaw : null;
  const ttlRaw = payload.ttl;
  const ttl = typeof ttlRaw === "number" ? ttlRaw : null;
  const isTemporary = Boolean(ttl && ttl > 0);

  return {
    id: messageId,
    roomId,
    authorId: senderId,
    authorHandle: senderId === currentUserId ? "você" : `user#${senderId.slice(0, 4)}`,
    content,
    status: senderId === currentUserId ? "sent" : "delivered",
    createdAt: sentAt,
    isOwn: senderId === currentUserId,
    expiresAt,
    ttlSeconds: ttl,
    isZeroLogging: Boolean(payload.is_zero_logging),
    isTemporary,
  };
}

export const useChatStore = create<ChatState>((set, get) => ({
  messagesByRoom: {},
  pendingAckIds: [],
  draftByRoom: {},
  connectionStatus: "idle",
  connectionError: null,
  presenceByUserId: {},
  setDraft: (roomId, value) => {
    set((state) => ({
      draftByRoom: {
        ...state.draftByRoom,
        [roomId]: value,
      },
    }));
  },
  connect: async (token) => {
    if (wsClient?.socket.readyState === WebSocket.OPEN) {
      return;
    }

    set((state) => ({
      connectionStatus:
        state.connectionStatus === "connected" ? "reconnecting" : "connecting",
      connectionError: null,
    }));

    const ticket = await apiRequest<WSTicketResponse>("/ws-ticket", {
      method: "POST",
      token,
    });

    wsClient = createWsClient(ticket.ticket, {
      onOpen: () => {
        reconnectAttempts = 0;
        set({ connectionStatus: "connected", connectionError: null });
      },
      onClose: () => {
        const currentToken = useSessionStore.getState().token;
        set({ connectionStatus: "offline" });
        if (currentToken) {
          scheduleReconnect(currentToken, get().connect);
        }
      },
      onError: () => {
        set({ connectionStatus: "offline", connectionError: "websocket_error" });
      },
      onMessage: (event) => {
        const currentUser = useSessionStore.getState().user;
        if (!currentUser) {
          return;
        }

        if (event.type === "message.received") {
          const message = createMessageFromGateway(
            event.payload as Record<string, unknown>,
            currentUser.userId,
          );
          if (!message) {
            return;
          }

          set((state) => {
            if (message.isOwn) {
              const { messagesByRoom, reconciledLocalId } = upsertOwnMessageFromServer(
                state.messagesByRoom,
                message,
              );

              return {
                messagesByRoom,
                pendingAckIds: reconciledLocalId
                  ? state.pendingAckIds.filter((id) => id !== reconciledLocalId)
                  : state.pendingAckIds,
                presenceByUserId: {
                  ...state.presenceByUserId,
                  [message.authorId]: "online",
                },
              };
            }

            return {
              messagesByRoom: appendMessage(state.messagesByRoom, message.roomId, message),
              presenceByUserId: {
                ...state.presenceByUserId,
                [message.authorId]: "online",
              },
            };
          });

          if (!message.isOwn) {
            get().markMessageDelivered(message.roomId, message.id);
          }
          return;
        }

        if (event.type === "message_ack") {
          set((state) => {
            const ackedId = state.pendingAckIds[0];
            if (!ackedId) {
              return state;
            }

            const nextMessagesByRoom = Object.fromEntries(
              Object.entries(state.messagesByRoom).map(([roomId, items]) => [
                roomId,
                items.map((message) =>
                  message.id === ackedId ? { ...message, status: "sent" as const } : message,
                ),
              ]),
            );

            return {
              messagesByRoom: nextMessagesByRoom,
              pendingAckIds: state.pendingAckIds.slice(1),
            };
          });
          return;
        }

        if (event.type === "message.delivered" || event.type === "message.read") {
          const payload = event.payload as { message_id?: string; room_id?: string };
          if (!payload.message_id || !payload.room_id) {
            return;
          }
          const roomId = payload.room_id;
          const messageId = payload.message_id;
          set((state) => ({
            messagesByRoom: {
              ...state.messagesByRoom,
              [roomId]: (state.messagesByRoom[roomId] ?? []).map((message) =>
                message.id === messageId
                  ? {
                      ...message,
                      status: event.type === "message.read" ? "read" : "delivered",
                    }
                  : message,
              ),
            },
          }));
          return;
        }

        if (event.type === "message.deleted" || event.type === "message.expired") {
          const payload = event.payload as { message_id?: string; room_id?: string };
          if (!payload.message_id || !payload.room_id) {
            return;
          }
          const roomId = payload.room_id;
          const messageId = payload.message_id;
          set((state) => ({
            messagesByRoom: {
              ...state.messagesByRoom,
              [roomId]: (state.messagesByRoom[roomId] ?? []).filter(
                (message) => message.id !== messageId,
              ),
            },
          }));
          return;
        }

        if (event.type === "session_expired") {
          useSessionStore.getState().logout().catch(() => undefined);
          set({ connectionStatus: "offline", connectionError: "session_expired" });
          return;
        }

        const nextError = deriveConnectionError(event);
        if (nextError) {
          set({ connectionError: nextError });
        }
      },
    });
  },
  disconnect: () => {
    if (reconnectTimer !== null) {
      window.clearTimeout(reconnectTimer);
      reconnectTimer = null;
    }
    reconnectAttempts = 0;
    wsClient?.close();
    wsClient = null;
    set({ connectionStatus: "idle" });
  },
  loadRoomHistory: async (roomId, token) => {
    if (!roomId || !token) {
      return;
    }

    let historyPayload: MessageHistoryItem[] | unknown;
    try {
      historyPayload = await apiRequest<MessageHistoryItem[] | unknown>("/messages", {
        token,
        query: { room_id: roomId },
      });
    } catch (error) {
      set({
        connectionError:
          error instanceof Error ? error.message : "history_load_failed",
      });
      return;
    }

    const history = Array.isArray(historyPayload) ? historyPayload : [];

    const currentUserId = useSessionStore.getState().user?.userId ?? "";
    const mapped: ChatMessage[] = history.map((item) => ({
      id: item.id,
      roomId,
      authorId: "system",
      authorHandle: "sistema",
      content: item.content,
      status: "read",
      createdAt: item.created_at,
      isOwn: false,
      expiresAt: null,
      ttlSeconds: null,
      isZeroLogging: false,
      isTemporary: false,
    }));

    set((state) => ({
      messagesByRoom: {
        ...state.messagesByRoom,
        [roomId]: mapped.length > 0 ? mapped : state.messagesByRoom[roomId] ?? [],
      },
      presenceByUserId: {
        ...state.presenceByUserId,
        [currentUserId]: "online",
      },
    }));
  },
  sendMessage: (roomId) => {
    const draft = get().draftByRoom[roomId]?.trim();
    const sessionUser = useSessionStore.getState().user;
    if (!draft || !sessionUser) {
      return;
    }

    if (!wsClient || wsClient.socket.readyState !== WebSocket.OPEN) {
      set({ connectionError: "offline_send_blocked" });
      return;
    }

    const messageId = `local-${crypto.randomUUID()}`;
    const optimisticMessage: ChatMessage = {
      id: messageId,
      roomId,
      authorId: sessionUser.userId,
      authorHandle: "você",
      content: draft,
      status: "sending",
      createdAt: new Date().toISOString(),
      isOwn: true,
      expiresAt: null,
      ttlSeconds: null,
      isZeroLogging: false,
      isTemporary: false,
    };

    set((state) => ({
      messagesByRoom: appendMessage(state.messagesByRoom, roomId, optimisticMessage),
      pendingAckIds: [...state.pendingAckIds, messageId],
      draftByRoom: {
        ...state.draftByRoom,
        [roomId]: "",
      },
    }));

    wsClient.send("send_message", {
      room_id: roomId,
      content: draft,
    });
  },
  markMessageRead: (roomId, messageId) => {
    set((state) => ({
      messagesByRoom: {
        ...state.messagesByRoom,
        [roomId]: (state.messagesByRoom[roomId] ?? []).map((message) =>
          message.id === messageId ? { ...message, status: "read" } : message,
        ),
      },
    }));

    if (!wsClient || wsClient.socket.readyState !== WebSocket.OPEN) {
      return;
    }
    wsClient.send("message_read", {
      room_id: roomId,
      message_id: messageId,
    });
  },
  markMessageDelivered: (roomId, messageId) => {
    set((state) => ({
      messagesByRoom: {
        ...state.messagesByRoom,
        [roomId]: (state.messagesByRoom[roomId] ?? []).map((message) =>
          message.id === messageId && message.status === "sent"
            ? { ...message, status: "delivered" }
            : message,
        ),
      },
    }));

    if (!wsClient || wsClient.socket.readyState !== WebSocket.OPEN) {
      return;
    }
    wsClient.send("message_delivered", {
      room_id: roomId,
      message_id: messageId,
    });
  },
}));
