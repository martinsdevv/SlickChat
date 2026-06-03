import { create } from "zustand";
import { apiRequest } from "../../shared/api/http-client";
import type { MessageHistoryItem, WSTicketResponse } from "../../shared/api/types";
import { createWsClient, type WsClient, type WsEnvelope } from "../../shared/api/ws-client";
import {
  clampMessageContent,
  MAX_MESSAGE_CONTENT_LENGTH,
} from "../../shared/constants/messages";
import { useRoomsStore } from "../rooms/store";
import { useSessionStore } from "../session/store";

type DeliveryStatus = "sending" | "sent" | "delivered" | "read" | "failed";
type PresenceStatus = "online" | "offline" | "unknown";
export type ConnectionStatus = "idle" | "connecting" | "connected" | "reconnecting" | "offline";

export type ChatMessage = {
  id: string;
  roomId: string;
  authorId: string;
  authorHandle: string;
  content: string;
  messageType: "TEXT" | "IMAGE";
  imageObjectKey?: string;
  imagePreviewUrl?: string;
  status: DeliveryStatus;
  createdAt: string;
  isOwn: boolean;
  expiresAt: string | null;
  ttlSeconds: number | null;
  isZeroLogging: boolean;
  isTemporary: boolean;
};

type PendingAttachment = {
  file: File;
  previewUrl: string;
};

type ChatState = {
  messagesByRoom: Record<string, ChatMessage[]>;
  pendingAckIds: string[];
  draftByRoom: Record<string, string>;
  pendingAttachmentByRoom: Record<string, PendingAttachment | undefined>;
  connectionStatus: ConnectionStatus;
  connectionError: string | null;
  presenceByUserId: Record<string, PresenceStatus>;
  setDraft: (roomId: string, value: string) => void;
  setPendingAttachment: (roomId: string, file: File) => void;
  clearPendingAttachment: (roomId: string) => void;
  connect: (token: string) => Promise<void>;
  disconnect: () => void;
  loadRoomHistory: (roomId: string, token: string) => Promise<boolean>;
  sendMessage: (roomId: string) => void;
  sendComposer: (roomId: string) => void;
  deleteMessages: (roomId: string, messageIds: string[]) => void;
  markMessageRead: (roomId: string, messageId: string) => void;
  markMessageDelivered: (roomId: string, messageId: string) => void;
};

let wsClient: WsClient | null = null;
let reconnectTimer: number | null = null;
let reconnectAttempts = 0;
const presenceTimeouts = new Map<string, number>();
const PRESENCE_ONLINE_TTL_MS = 45_000;

function clearPresenceTimeouts() {
  presenceTimeouts.forEach((timeoutId) => {
    window.clearTimeout(timeoutId);
  });
  presenceTimeouts.clear();
}

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
  const existingIndex = roomMessages.findIndex((item) => item.id === messageFromServer.id);
  if (existingIndex !== -1) {
    const existing = roomMessages[existingIndex];
    if (
      messageFromServer.messageType === "IMAGE" &&
      existing.messageType === "IMAGE"
    ) {
      const merged = mergeImageMessageFromServer(existing, messageFromServer);
      if (
        merged.content !== existing.content ||
        merged.imageObjectKey !== existing.imageObjectKey
      ) {
        const nextRoomMessages = roomMessages.map((item, index) =>
          index === existingIndex ? merged : item,
        );
        return {
          messagesByRoom: {
            ...messagesByRoom,
            [messageFromServer.roomId]: nextRoomMessages,
          },
          reconciledLocalId: null as string | null,
        };
      }
    }
    return { messagesByRoom, reconciledLocalId: null as string | null };
  }

  const serverTimestamp = Date.parse(messageFromServer.createdAt);
  const optimisticIndex = roomMessages.findIndex((item) => {
    if (!item.isOwn) {
      return false;
    }
    if (item.id === messageFromServer.id) {
      return true;
    }
    if (!item.id.startsWith("local-")) {
      return false;
    }
    if (item.messageType !== messageFromServer.messageType) {
      return false;
    }
    if (item.messageType === "IMAGE") {
      return item.imageObjectKey === messageFromServer.imageObjectKey;
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
  const optimistic = roomMessages[optimisticIndex];
  const nextRoomMessages = roomMessages.map((item, index) =>
    index === optimisticIndex
      ? messageFromServer.messageType === "IMAGE"
        ? mergeImageMessageFromServer(optimistic, messageFromServer)
        : {
            ...messageFromServer,
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

function parseImageMessageFields(
  messageType: "TEXT" | "IMAGE",
  content: string,
  attachmentObjectKey: string,
  explicitCaption?: string,
): { caption: string; imageObjectKey?: string } | null {
  if (messageType !== "IMAGE") {
    return { caption: explicitCaption?.trim() || content };
  }

  const key =
    attachmentObjectKey ||
    (content.startsWith("messages/") ? content : "");
  if (!key) {
    return null;
  }

  let caption = explicitCaption?.trim() ?? "";
  if (!caption) {
    if (attachmentObjectKey.length > 0 && content && !content.startsWith("messages/")) {
      caption = content;
    } else if (!content.startsWith("messages/")) {
      caption = content;
    }
  }

  return { caption, imageObjectKey: key };
}

function mergeImageMessageFromServer(
  optimistic: ChatMessage,
  fromServer: ChatMessage,
): ChatMessage {
  const serverCaption = fromServer.content.trim();
  const optimisticCaption = optimistic.content.trim();
  const caption =
    serverCaption || optimisticCaption || fromServer.content || optimistic.content;

  return {
    ...fromServer,
    content: caption,
    imageObjectKey: fromServer.imageObjectKey ?? optimistic.imageObjectKey,
    imagePreviewUrl: fromServer.imagePreviewUrl ?? optimistic.imagePreviewUrl,
    status:
      optimistic.status === "delivered" || optimistic.status === "read"
        ? optimistic.status
        : fromServer.status,
  };
}

function createMessageFromGateway(
  payload: Record<string, unknown>,
  currentUserId: string,
): ChatMessage | null {
  const roomId = String(payload.room_id ?? "");
  const messageId = String(payload.message_id ?? "");
  const senderId = String(payload.sender_id ?? "");
  const messageTypeRaw = String(payload.message_type ?? "TEXT").toUpperCase();
  const messageType = messageTypeRaw === "IMAGE" ? "IMAGE" : "TEXT";
  const rawContent = String(payload.content ?? "");
  const attachmentObjectKey = String(payload.attachment_object_key ?? "");
  const explicitCaption = String(payload.caption ?? "");
  const parsed = parseImageMessageFields(
    messageType,
    rawContent,
    attachmentObjectKey,
    explicitCaption,
  );
  if (!roomId || !messageId || !parsed) {
    return null;
  }
  if (messageType === "TEXT" && !parsed.caption) {
    return null;
  }

  const sentAtRaw = payload.sent_at;
  const sentAt = typeof sentAtRaw === "string" ? sentAtRaw : new Date().toISOString();
  const expiresAtRaw = payload.expires_at;
  const ttlRaw = payload.ttl;
  const ttl = typeof ttlRaw === "number" ? ttlRaw : null;
  const expiresAt =
    typeof expiresAtRaw === "string"
      ? expiresAtRaw
      : ttl && ttl > 0
        ? new Date(Date.parse(sentAt) + ttl * 1000).toISOString()
        : null;
  const isTemporary = Boolean(ttl && ttl > 0);

  return {
    id: messageId,
    roomId,
    authorId: senderId,
    authorHandle: senderId === currentUserId ? "você" : `user#${senderId.slice(0, 4)}`,
    content: clampMessageContent(parsed.caption),
    messageType,
    imageObjectKey: parsed.imageObjectKey,
    status: senderId === currentUserId ? "sent" : "delivered",
    createdAt: sentAt,
    isOwn: senderId === currentUserId,
    expiresAt,
    ttlSeconds: ttl,
    isZeroLogging: Boolean(payload.is_zero_logging),
    isTemporary,
  };
}

function revokePendingAttachment(pending?: PendingAttachment) {
  if (pending?.previewUrl.startsWith("blob:")) {
    URL.revokeObjectURL(pending.previewUrl);
  }
}

export const useChatStore = create<ChatState>((set, get) => ({
  messagesByRoom: {},
  pendingAckIds: [],
  draftByRoom: {},
  pendingAttachmentByRoom: {},
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
  setPendingAttachment: (roomId, file) => {
    const previewUrl = URL.createObjectURL(file);
    set((state) => {
      revokePendingAttachment(state.pendingAttachmentByRoom[roomId]);
      return {
        pendingAttachmentByRoom: {
          ...state.pendingAttachmentByRoom,
          [roomId]: { file, previewUrl },
        },
      };
    });
  },
  clearPendingAttachment: (roomId) => {
    set((state) => {
      revokePendingAttachment(state.pendingAttachmentByRoom[roomId]);
      const next = { ...state.pendingAttachmentByRoom };
      delete next[roomId];
      return { pendingAttachmentByRoom: next };
    });
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
        const currentUserId = useSessionStore.getState().user?.userId;
        set((state) => ({
          connectionStatus: "connected",
          connectionError: null,
          presenceByUserId: currentUserId
            ? {
                ...state.presenceByUserId,
                [currentUserId]: "online",
              }
            : state.presenceByUserId,
        }));
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
            const existingPresenceTimeout = presenceTimeouts.get(message.authorId);
            if (existingPresenceTimeout) {
              window.clearTimeout(existingPresenceTimeout);
            }
            const timeoutId = window.setTimeout(() => {
              set((currentState) => ({
                presenceByUserId: {
                  ...currentState.presenceByUserId,
                  [message.authorId]: "offline",
                },
              }));
              presenceTimeouts.delete(message.authorId);
            }, PRESENCE_ONLINE_TTL_MS);
            presenceTimeouts.set(message.authorId, timeoutId);

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
            const activeRoomId = useRoomsStore.getState().activeRoomId;
            if (message.roomId !== activeRoomId) {
              useRoomsStore.getState().bumpRoomUnread(message.roomId);
            }
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
    clearPresenceTimeouts();
    wsClient?.close();
    wsClient = null;
    set({ connectionStatus: "idle" });
  },
  loadRoomHistory: async (roomId, token) => {
    if (!roomId || !token) {
      return false;
    }

    const maxAttempts = 4;
    let lastError: unknown = null;
    let historyPayload: MessageHistoryItem[] | unknown = [];
    for (let attempt = 1; attempt <= maxAttempts; attempt += 1) {
      try {
        historyPayload = await apiRequest<MessageHistoryItem[] | unknown>("/messages", {
          token,
          query: { room_id: roomId },
        });
        lastError = null;
        break;
      } catch (error) {
        lastError = error;
        if (attempt < maxAttempts) {
          await new Promise((resolve) => window.setTimeout(resolve, attempt * 500));
        }
      }
    }

    if (lastError) {
      set({
        connectionError:
          lastError instanceof Error ? lastError.message : "history_load_failed",
      });
      return false;
    }

    let parsedPayload: unknown = historyPayload;
    if (typeof historyPayload === "string") {
      try {
        parsedPayload = JSON.parse(historyPayload);
      } catch {
        parsedPayload = [];
      }
    }
    const history = Array.isArray(parsedPayload) ? parsedPayload : [];

    const currentUserId = useSessionStore.getState().user?.userId ?? "";
    const currentHandle = useSessionStore.getState().user?.handle ?? "";
    const room = useRoomsStore.getState().rooms.find((r) => r.room_id === roomId);
    const mapped: ChatMessage[] = history.flatMap((item) => {
      const senderId = typeof item.sender_id === "string" ? item.sender_id : "";
      const isOwn = senderId !== "" && senderId === currentUserId;
      const expiresAt = item.expires_at ?? null;
      const hasExpiration = expiresAt !== null;
      const messageType = item.type === "IMAGE" ? "IMAGE" : "TEXT";
      const parsed = parseImageMessageFields(
        messageType,
        item.content,
        item.attachment_object_key ?? "",
        item.caption,
      );
      if (messageType === "IMAGE" && !parsed) {
        return [];
      }

      return [
        {
          id: item.id,
          roomId,
          authorId: senderId || "unknown",
          authorHandle: isOwn
            ? currentHandle
            : senderId
              ? `user#${senderId.slice(0, 4)}`
              : "desconhecido",
          content: clampMessageContent(parsed?.caption ?? item.content),
          messageType,
          imageObjectKey: parsed?.imageObjectKey,
          status: "read" as const,
          createdAt: item.created_at,
          isOwn,
          expiresAt,
          ttlSeconds: room?.ttl ?? null,
          isZeroLogging: room?.zero_logging ?? false,
          isTemporary: hasExpiration || (room?.type === "TEMPORARY"),
        },
      ];
    });
    const sortedMapped = [...mapped].sort((a, b) => a.createdAt.localeCompare(b.createdAt));

    set((state) => ({
      messagesByRoom: {
        ...state.messagesByRoom,
        [roomId]: sortedMapped,
      },
      presenceByUserId: {
        ...state.presenceByUserId,
        [currentUserId]: "online",
      },
    }));
    return true;
  },
  sendMessage: (roomId) => {
    const draft = get().draftByRoom[roomId]?.trim();
    const sessionUser = useSessionStore.getState().user;
    if (!draft || !sessionUser) {
      return;
    }
    if (draft.length > MAX_MESSAGE_CONTENT_LENGTH) {
      set({ connectionError: "content_too_long" });
      return;
    }

    if (!wsClient || wsClient.socket.readyState !== WebSocket.OPEN) {
      set({ connectionError: "offline_send_blocked" });
      return;
    }

    const messageId = `local-${crypto.randomUUID()}`;
    const room = useRoomsStore.getState().rooms.find((item) => item.room_id === roomId) ?? null;
    const ttlSeconds = room && room.ttl > 0 ? room.ttl : null;
    const expiresAt =
      ttlSeconds !== null ? new Date(Date.now() + ttlSeconds * 1000).toISOString() : null;
    const optimisticMessage: ChatMessage = {
      id: messageId,
      roomId,
      authorId: sessionUser.userId,
      authorHandle: sessionUser.handle,
      content: draft,
      messageType: "TEXT",
      status: "sending",
      createdAt: new Date().toISOString(),
      isOwn: true,
      expiresAt,
      ttlSeconds,
      isZeroLogging: Boolean(room?.zero_logging),
      isTemporary: Boolean(ttlSeconds && ttlSeconds > 0),
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
  sendComposer: async (roomId) => {
    const pending = get().pendingAttachmentByRoom[roomId];
    if (pending) {
      const caption = get().draftByRoom[roomId]?.trim() ?? "";
      if (caption.length > MAX_MESSAGE_CONTENT_LENGTH) {
        set({ connectionError: "content_too_long" });
        return;
      }
      const sessionUser = useSessionStore.getState().user;
      const authToken = useSessionStore.getState().token;
      if (!sessionUser || !authToken) {
        return;
      }

      if (!wsClient || wsClient.socket.readyState !== WebSocket.OPEN) {
        set({ connectionError: "offline_send_blocked" });
        return;
      }

      const messageId = crypto.randomUUID();
      const { file } = pending;
      const messagePreviewUrl = URL.createObjectURL(file);
      const room = useRoomsStore.getState().rooms.find((item) => item.room_id === roomId) ?? null;
      const ttlSeconds = room && room.ttl > 0 ? room.ttl : null;
      const expiresAt =
        ttlSeconds !== null ? new Date(Date.now() + ttlSeconds * 1000).toISOString() : null;

      const optimisticMessage: ChatMessage = {
        id: messageId,
        roomId,
        authorId: sessionUser.userId,
        authorHandle: sessionUser.handle,
        content: caption,
        messageType: "IMAGE",
        imagePreviewUrl: messagePreviewUrl,
        status: "sending",
        createdAt: new Date().toISOString(),
        isOwn: true,
        expiresAt,
        ttlSeconds,
        isZeroLogging: Boolean(room?.zero_logging),
        isTemporary: Boolean(ttlSeconds && ttlSeconds > 0),
      };

      set((state) => {
        revokePendingAttachment(state.pendingAttachmentByRoom[roomId]);
        const nextPending = { ...state.pendingAttachmentByRoom };
        delete nextPending[roomId];
        return {
          messagesByRoom: appendMessage(state.messagesByRoom, roomId, optimisticMessage),
          pendingAckIds: [...state.pendingAckIds, messageId],
          pendingAttachmentByRoom: nextPending,
          draftByRoom: { ...state.draftByRoom, [roomId]: "" },
        };
      });

      try {
        const { uploadMessageImage } = await import("../media/upload-message-image");
        const uploaded = await uploadMessageImage(authToken, roomId, messageId, file);

        set((state) => ({
          messagesByRoom: {
            ...state.messagesByRoom,
            [roomId]: (state.messagesByRoom[roomId] ?? []).map((message) =>
              message.id === messageId
                ? { ...message, imageObjectKey: uploaded.object_key }
                : message,
            ),
          },
        }));

        wsClient.send("send_message", {
          room_id: roomId,
          content: caption,
          message_id: messageId,
          message_type: "IMAGE",
          object_key: uploaded.object_key,
        });
      } catch (error) {
        URL.revokeObjectURL(messagePreviewUrl);
        set((state) => ({
          messagesByRoom: {
            ...state.messagesByRoom,
            [roomId]: (state.messagesByRoom[roomId] ?? []).filter((message) => message.id !== messageId),
          },
          pendingAckIds: state.pendingAckIds.filter((id) => id !== messageId),
          connectionError:
            error instanceof Error ? error.message.replaceAll("\n", " ").trim() : "upload_failed",
        }));
      }
      return;
    }

    get().sendMessage(roomId);
  },
  deleteMessages: (roomId, messageIds) => {
    if (!roomId || messageIds.length === 0) {
      return;
    }
    const uniqueMessageIds = [...new Set(messageIds)];

    // Optimistic UI: remove messages immediately from local state.
    set((state) => ({
      messagesByRoom: {
        ...state.messagesByRoom,
        [roomId]: (state.messagesByRoom[roomId] ?? []).filter(
          (message) => !uniqueMessageIds.includes(message.id),
        ),
      },
    }));

    if (!wsClient || wsClient.socket.readyState !== WebSocket.OPEN) {
      set({ connectionError: "offline_delete_blocked" });
      return;
    }

    uniqueMessageIds.forEach((messageId) => {
      wsClient?.send("delete_message", {
        room_id: roomId,
        message_id: messageId,
      });
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
