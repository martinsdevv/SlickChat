import { useEffect, useId, useMemo, useRef, useState, type CSSProperties } from "react";
import {
  useChatStore,
  type ChatMessage,
  type ConnectionStatus,
} from "../../../features/chat/store";
import { filterRooms, useRoomsStore } from "../../../features/rooms/store";
import { useSessionStore } from "../../../features/session/store";
import { useUIStore } from "../../../features/ui/store";
import { ApiError } from "../../../shared/api/types";
import {
  formatDateTime,
  formatRoomType,
  formatRole,
  MetaRow,
  ParticipantRow,
  RoomAvatar,
  RoomTypeBadge,
  UserSessionBar,
} from "./room-ui";
import { MAX_MESSAGE_CONTENT_LENGTH } from "../../../shared/constants/messages";
import { cn } from "../../../shared/lib/cn";
import { MessageImageContent } from "./message-image-content";
import { MessageText } from "./message-text";
import { RoomMediaEditor } from "./room-media-editor";

const CONNECTION_LABELS: Record<ConnectionStatus, string> = {
  idle: "Preparando",
  connecting: "Conectando…",
  connected: "Conectado",
  reconnecting: "Reconectando…",
  offline: "Offline",
};

function formatCountdown(seconds: number) {
  if (seconds <= 0) {
    return "desaparece agora";
  }
  return `desaparece em ${seconds}s`;
}

function getTemporaryVisualState(message: ChatMessage, nowMs: number) {
  if (!message.expiresAt) {
    return {
      timeLeftSeconds: null,
      className: "",
      style: {},
    };
  }

  const expirationMs = new Date(message.expiresAt).getTime();
  const timeLeftSeconds = Math.max(Math.ceil((expirationMs - nowMs) / 1000), 0);

  if (timeLeftSeconds <= 5) {
    const blur = Number(((5 - timeLeftSeconds) * 0.9).toFixed(2));
    return {
      timeLeftSeconds,
      className: "opacity-80",
      style: {
        filter: `blur(${blur}px)`,
        opacity: Math.max(timeLeftSeconds / 5, 0.2),
      },
    };
  }

  if (timeLeftSeconds <= 10) {
    return {
      timeLeftSeconds,
      className: "animate-pulse",
      style: {
        opacity: 0.9,
      },
    };
  }

  return {
    timeLeftSeconds,
    className: "",
    style: {},
  };
}

export function ChatShell() {
  const profileAvatarInputId = useId();
  const messageFileInputId = useId();
  const { user, token, logout } = useSessionStore();
  const {
    rooms,
    activeRoomId,
    membersByRoom,
    roomFilter,
    isLoadingRooms,
    isCreatingRoom,
    setRoomFilter,
    setActiveRoom,
    unreadByRoom,
    loadRooms,
    createRoom,
    joinRoom,
    addMember,
    loadMembers,
    patchRoomMedia,
  } = useRoomsStore();
  const {
    messagesByRoom,
    draftByRoom,
    connectionStatus,
    connectionError,
    presenceByUserId,
    setDraft,
    connect,
    disconnect,
    loadRoomHistory,
    sendComposer,
    setPendingAttachment,
    clearPendingAttachment,
    pendingAttachmentByRoom,
    deleteMessages,
    markMessageRead,
  } = useChatStore();
  const {
    isRightPanelOpen,
    mobileView,
    toggleRightPanel,
    setMobileView,
  } = useUIStore();
  const [clock, setClock] = useState(() => Date.now());
  const [memberHandleInput, setMemberHandleInput] = useState("");
  const [addMemberFeedback, setAddMemberFeedback] = useState<string | null>(null);
  const [isCreateFormOpen, setIsCreateFormOpen] = useState(false);
  const [newRoomName, setNewRoomName] = useState("");
  const [newRoomDescription, setNewRoomDescription] = useState("");
  const [newRoomType, setNewRoomType] = useState<"PUBLIC" | "PRIVATE" | "TEMPORARY">("PUBLIC");
  const [newRoomTTL, setNewRoomTTL] = useState(60);
  const [newRoomZeroLogging, setNewRoomZeroLogging] = useState(false);
  const [createRoomFeedback, setCreateRoomFeedback] = useState<string | null>(null);
  const [isSelectionMode, setIsSelectionMode] = useState(false);
  const [selectedMessageIds, setSelectedMessageIds] = useState<string[]>([]);
  const [joinRoomIdInput, setJoinRoomIdInput] = useState("");
  const [isJoiningRoom, setIsJoiningRoom] = useState(false);
  const [joinRoomFeedback, setJoinRoomFeedback] = useState<string | null>(null);
  const [roomIdCopyFeedback, setRoomIdCopyFeedback] = useState<string | null>(null);
  const messagesEndRef = useRef<HTMLDivElement | null>(null);

  const activeRoom = rooms.find((room) => room.room_id === activeRoomId) ?? null;
  const roomMessages = useMemo(
    () => (activeRoomId ? messagesByRoom[activeRoomId] ?? [] : []),
    [activeRoomId, messagesByRoom],
  );
  const roomMembers = useMemo(
    () => (activeRoomId ? membersByRoom[activeRoomId] ?? [] : []),
    [activeRoomId, membersByRoom],
  );
  const filteredRooms = useMemo(() => filterRooms(rooms, roomFilter), [roomFilter, rooms]);
  const visibleRoomMessages = useMemo(
    () =>
      roomMessages.filter((message) => {
        if (!message.expiresAt) {
          return true;
        }
        return new Date(message.expiresAt).getTime() > clock;
      }),
    [clock, roomMessages],
  );
  const memberHandleByUserId = useMemo(
    () =>
      roomMembers.reduce<Record<string, string>>((acc, member) => {
        acc[member.user_id] = member.handle;
        return acc;
      }, {}),
    [roomMembers],
  );
  const lastMessageByRoom = useMemo(
    () => {
      const entries = Object.entries(messagesByRoom).map(([roomId, messages]) => {
        const visible = messages.filter((message) => {
          if (!message.expiresAt) {
            return true;
          }
          return new Date(message.expiresAt).getTime() > clock;
        });
        return [roomId, visible[visible.length - 1] ?? null];
      });
      return Object.fromEntries(entries) as Record<string, ChatMessage | null>;
    },
    [clock, messagesByRoom],
  );
  const orderedRooms = useMemo(() => {
    const roomOrder = new Map(filteredRooms.map((room, index) => [room.room_id, index]));
    return [...filteredRooms].sort((a, b) => {
      const aUnread = unreadByRoom[a.room_id] ?? 0;
      const bUnread = unreadByRoom[b.room_id] ?? 0;
      if (aUnread !== bUnread) {
        return bUnread - aUnread;
      }

      const aLast = lastMessageByRoom[a.room_id];
      const bLast = lastMessageByRoom[b.room_id];
      const aTime = aLast ? Date.parse(aLast.createdAt) : Number.NEGATIVE_INFINITY;
      const bTime = bLast ? Date.parse(bLast.createdAt) : Number.NEGATIVE_INFINITY;

      if (aTime !== bTime) {
        return bTime - aTime;
      }

      return (roomOrder.get(a.room_id) ?? 0) - (roomOrder.get(b.room_id) ?? 0);
    });
  }, [filteredRooms, lastMessageByRoom, unreadByRoom]);
  const myMembership = useMemo(
    () => roomMembers.find((member) => member.user_id === user?.userId) ?? null,
    [roomMembers, user?.userId],
  );
  const canAddMembers = myMembership?.role === "ADMIN";
  const canDeleteAnyMessage = myMembership?.role === "ADMIN";
  const activeRoomExpiresInSeconds = useMemo(() => {
    if (!activeRoom?.expires_at) {
      return null;
    }
    return Math.max(Math.ceil((new Date(activeRoom.expires_at).getTime() - clock) / 1000), 0);
  }, [activeRoom?.expires_at, clock]);

  function handleLogout() {
    void logout().then(() => disconnect());
  }

  useEffect(() => {
    if (!token) {
      return;
    }
    void loadRooms(token).catch(() => undefined);
  }, [loadRooms, token]);

  useEffect(() => {
    if (!token) {
      return;
    }
    void connect(token);
    return () => disconnect();
  }, [connect, disconnect, token]);

  useEffect(() => {
    if (!activeRoom || !token) {
      return;
    }

    void joinRoom(token, activeRoom.room_id)
      .catch(() => undefined)
      .finally(() => {
        void loadRoomHistory(activeRoom.room_id, token);
        void loadMembers(token, activeRoom.room_id).catch(() => undefined);
      });
  }, [activeRoom, joinRoom, loadMembers, loadRoomHistory, token]);

  useEffect(() => {
    if (!token || rooms.length === 0) {
      return;
    }

    const roomIdsWithoutHistory = rooms
      .map((room) => room.room_id)
      .filter((roomId) => messagesByRoom[roomId] === undefined);

    if (roomIdsWithoutHistory.length === 0) {
      return;
    }

    void Promise.allSettled(
      roomIdsWithoutHistory.map((roomId) => loadRoomHistory(roomId, token)),
    );
  }, [loadRoomHistory, messagesByRoom, rooms, token]);

  useEffect(() => {
    const timer = window.setInterval(() => setClock(Date.now()), 1000);
    return () => window.clearInterval(timer);
  }, []);

  useEffect(() => {
    if (!activeRoomId) {
      return;
    }

    visibleRoomMessages
      .filter((message) => !message.isOwn && message.status !== "read")
      .forEach((message) => {
        markMessageRead(activeRoomId, message.id);
      });
  }, [activeRoomId, markMessageRead, visibleRoomMessages]);

  useEffect(() => {
    const anchor = messagesEndRef.current;
    if (!anchor) {
      return;
    }
    const frame1 = window.requestAnimationFrame(() => {
      const frame2 = window.requestAnimationFrame(() => {
        anchor.scrollIntoView({ block: "end" });
      });
      return () => window.cancelAnimationFrame(frame2);
    });
    return () => window.cancelAnimationFrame(frame1);
  }, [activeRoomId, mobileView, roomMessages.length]);

  useEffect(() => {
    setIsSelectionMode(false);
    setSelectedMessageIds([]);
  }, [activeRoomId]);

  function resetCreateRoomForm() {
    setNewRoomName("");
    setNewRoomDescription("");
    setNewRoomType("PUBLIC");
    setNewRoomTTL(60);
    setNewRoomZeroLogging(false);
    setCreateRoomFeedback(null);
  }

  async function handleCreateRoom() {
    if (!token) {
      return;
    }

    const trimmedName = newRoomName.trim();
    if (!trimmedName) {
      setCreateRoomFeedback("Nome da sala é obrigatório.");
      return;
    }

    const ttl = newRoomType === "TEMPORARY" ? Math.max(1, Math.floor(newRoomTTL || 60)) : 0;
    const created = await createRoom(token, {
      name: trimmedName,
      description: newRoomDescription.trim(),
      type: newRoomType,
      ttl,
      zero_logging: newRoomZeroLogging,
    }).catch((error) => {
      setCreateRoomFeedback(
        error instanceof Error ? error.message.replaceAll("\n", " ").trim() : "Falha ao criar sala.",
      );
      return null;
    });
    if (!created) {
      return;
    }
    resetCreateRoomForm();
    setIsCreateFormOpen(false);
    setActiveRoom(created.room_id);
  }

  async function handleJoinRoomById() {
    if (!token) {
      return;
    }

    const roomId = joinRoomIdInput.trim();
    if (!roomId) {
      setJoinRoomFeedback("Cole o ID da sala.");
      return;
    }

    const uuidPattern =
      /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
    if (!uuidPattern.test(roomId)) {
      setJoinRoomFeedback("ID inválido. Use o formato UUID da sala.");
      return;
    }

    setIsJoiningRoom(true);
    setJoinRoomFeedback(null);
    try {
      await joinRoom(token, roomId);
      await loadRooms(token);
      setActiveRoom(roomId);
      setJoinRoomIdInput("");
      setJoinRoomFeedback("Você entrou na sala.");
      setMobileView("chat");
      window.setTimeout(() => setJoinRoomFeedback(null), 3000);
    } catch (error) {
      if (error instanceof ApiError) {
        if (error.status === 404) {
          setJoinRoomFeedback("Sala não encontrada. Confira o ID.");
          return;
        }
        if (error.status === 403) {
          setJoinRoomFeedback("Só é possível entrar com ID em salas públicas.");
          return;
        }
      }
      setJoinRoomFeedback(
        error instanceof Error
          ? error.message.replaceAll("\n", " ").trim()
          : "Não foi possível entrar na sala.",
      );
    } finally {
      setIsJoiningRoom(false);
    }
  }

  async function handleCopyRoomId() {
    if (!activeRoom || activeRoom.type !== "PUBLIC") {
      return;
    }

    try {
      await navigator.clipboard.writeText(activeRoom.room_id);
      setRoomIdCopyFeedback("ID copiado!");
    } catch {
      setRoomIdCopyFeedback("Não foi possível copiar.");
    } finally {
      window.setTimeout(() => setRoomIdCopyFeedback(null), 2000);
    }
  }

  async function handleAddMember() {
    if (!token || !activeRoom || !canAddMembers) {
      return;
    }

    const handle = memberHandleInput.trim();
    if (!handle) {
      setAddMemberFeedback("Informe um handle válido (ex: usuario#1234).");
      return;
    }

    const result = await addMember(token, activeRoom.room_id, handle)
      .then(() => "Participante adicionado.")
      .catch((error) =>
        error instanceof Error ? error.message.replaceAll("\n", " ").trim() : "Falha ao adicionar participante.",
      );

    setAddMemberFeedback(result);
    if (result === "Participante adicionado.") {
      setMemberHandleInput("");
      void loadMembers(token, activeRoom.room_id).catch(() => undefined);
    }
  }

  function canDeleteMessage(message: ChatMessage) {
    return canDeleteAnyMessage || message.isOwn;
  }

  function toggleMessageSelection(messageId: string) {
    setSelectedMessageIds((current) =>
      current.includes(messageId)
        ? current.filter((id) => id !== messageId)
        : [...current, messageId],
    );
  }

  function handleDeleteSelectedMessages() {
    if (!activeRoomId || selectedMessageIds.length === 0) {
      return;
    }
    const deletableIds = visibleRoomMessages
      .filter((message) => selectedMessageIds.includes(message.id) && canDeleteMessage(message))
      .map((message) => message.id);
    if (deletableIds.length === 0) {
      return;
    }
    deleteMessages(activeRoomId, deletableIds);
    setSelectedMessageIds([]);
    setIsSelectionMode(false);
  }

  function openRoomInfoPanel() {
    if (!activeRoom) {
      return;
    }
    if (!isRightPanelOpen) {
      toggleRightPanel();
    }
    if (window.matchMedia("(max-width: 1023px)").matches) {
      setMobileView("info");
    }
  }

  function closeRoomInfoPanel() {
    if (isRightPanelOpen) {
      toggleRightPanel();
    }
    if (mobileView === "info") {
      setMobileView("chat");
    }
  }

  if (!user) {
    return null;
  }

  const showDesktopRightColumn = isRightPanelOpen;

  return (
    <div
      className={cn(
        "grid h-dvh min-w-0 grid-cols-1 overflow-hidden bg-[#07080d]",
        showDesktopRightColumn
          ? "lg:grid-cols-[minmax(0,320px)_minmax(0,1fr)_minmax(0,380px)]"
          : "lg:grid-cols-[minmax(0,320px)_minmax(0,1fr)]",
      )}
    >
      <aside
        className={`min-h-0 min-w-0 border-r border-white/8 bg-[#0d0e12] px-4 py-4 ${
          mobileView === "rooms" ? "block" : "hidden lg:block"
        }`}
      >
        <div className="flex h-full min-h-0 flex-col">
        <h1 className="text-3xl font-semibold tracking-wide text-[var(--text-0)]">
          SlickChat
        </h1>
        <div className="mt-3">
          <UserSessionBar
            handle={user.handle}
            userId={user.userId}
            avatarInputId={profileAvatarInputId}
            onLogout={handleLogout}
          />
        </div>
        <p className="mt-4 text-sm font-medium text-[var(--text-2)]">Suas salas</p>
        <div className="mt-3 grid gap-2">
          <button
            disabled={isCreatingRoom || isCreateFormOpen}
            onClick={() => {
              resetCreateRoomForm();
              setIsCreateFormOpen(true);
            }}
            className="w-full rounded-xl bg-gradient-to-r from-[var(--primary-500)] to-[var(--primary-400)] px-4 py-2.5 text-left text-base font-medium text-white transition hover:brightness-110 disabled:opacity-60"
          >
            Criar sala
          </button>
        </div>
        <div className="mt-2 space-y-1.5">
          <label className="text-xs text-[var(--text-3)]" htmlFor="join-room-id">
            Entrar com ID da sala
          </label>
          <div className="flex gap-2">
            <input
              id="join-room-id"
              value={joinRoomIdInput}
              onChange={(event) => setJoinRoomIdInput(event.target.value)}
              onKeyDown={(event) => {
                if (event.key === "Enter") {
                  event.preventDefault();
                  void handleJoinRoomById();
                }
              }}
              placeholder="UUID da sala pública"
              className="min-w-0 flex-1 rounded-lg border border-white/15 bg-[var(--bg-1)] px-2.5 py-2 text-xs text-[var(--text-0)] outline-none focus:border-[var(--primary-400)]"
            />
            <button
              type="button"
              disabled={isJoiningRoom || !joinRoomIdInput.trim()}
              onClick={() => void handleJoinRoomById()}
              className="shrink-0 rounded-lg border border-white/15 px-2.5 py-2 text-xs text-[var(--text-1)] transition hover:bg-white/10 disabled:opacity-50"
            >
              {isJoiningRoom ? "…" : "Entrar"}
            </button>
          </div>
          {joinRoomFeedback ? (
            <p className="text-xs text-[var(--text-2)]">{joinRoomFeedback}</p>
          ) : null}
        </div>
        {isCreateFormOpen ? (
          <div className="mt-3 space-y-2 rounded-xl border border-white/10 bg-[#14151a] p-3">
            <input
              value={newRoomName}
              onChange={(event) => setNewRoomName(event.target.value)}
              placeholder="Nome da sala"
              className="w-full rounded-lg border border-white/15 bg-[#1a1b20] px-3 py-2 text-sm text-[var(--text-0)] outline-none focus:border-[var(--primary-400)]"
            />
            <textarea
              value={newRoomDescription}
              onChange={(event) => setNewRoomDescription(event.target.value)}
              placeholder="Descrição (opcional)"
              rows={2}
              className="w-full resize-none rounded-lg border border-white/15 bg-[#1a1b20] px-3 py-2 text-sm text-[var(--text-0)] outline-none focus:border-[var(--primary-400)]"
            />
            <select
              value={newRoomType}
              onChange={(event) => {
                const nextType = event.target.value as "PUBLIC" | "PRIVATE" | "TEMPORARY";
                setNewRoomType(nextType);
                if (nextType !== "TEMPORARY") {
                  setNewRoomTTL(60);
                }
              }}
              className="w-full rounded-lg border border-white/15 bg-[#1a1b20] px-3 py-2 text-sm text-[var(--text-0)] outline-none focus:border-[var(--primary-400)]"
            >
              <option value="PUBLIC">PUBLIC</option>
              <option value="PRIVATE">PRIVATE</option>
              <option value="TEMPORARY">TEMPORARY</option>
            </select>
            {newRoomType === "TEMPORARY" ? (
              <input
                type="number"
                min={1}
                value={newRoomTTL}
                onChange={(event) => setNewRoomTTL(Number(event.target.value))}
                placeholder="TTL em segundos"
                className="w-full rounded-lg border border-white/15 bg-[#1a1b20] px-3 py-2 text-sm text-[var(--text-0)] outline-none focus:border-[var(--primary-400)]"
              />
            ) : null}
            <label className="flex items-center gap-2 text-xs text-[var(--text-2)]">
              <input
                type="checkbox"
                checked={newRoomZeroLogging}
                onChange={(event) => setNewRoomZeroLogging(event.target.checked)}
              />
              Zero logging
            </label>
            {createRoomFeedback ? (
              <p className="text-xs text-[var(--warning-500)]">{createRoomFeedback}</p>
            ) : null}
            <div className="flex gap-2">
              <button
                type="button"
                disabled={isCreatingRoom}
                onClick={() => void handleCreateRoom()}
                className="w-full rounded-lg bg-[var(--primary-500)] px-3 py-2 text-sm text-white transition-all duration-200 hover:-translate-y-0.5 hover:brightness-110 disabled:opacity-60"
              >
                {isCreatingRoom ? "Criando..." : "Criar"}
              </button>
              <button
                type="button"
                onClick={() => {
                  setIsCreateFormOpen(false);
                  resetCreateRoomForm();
                }}
                className="w-full rounded-lg border border-white/15 px-3 py-2 text-sm text-[var(--text-2)] transition-all duration-200 hover:bg-white/10"
              >
                Cancelar
              </button>
            </div>
          </div>
        ) : null}

        <div className="mt-4 grid grid-cols-3 gap-2 text-xs">
          {(["ALL", "TEMPORARY", "ZERO_LOGGING"] as const).map((filter) => (
            <button
              key={filter}
              onClick={() => setRoomFilter(filter)}
              className={`rounded-lg px-2 py-1 ${
                roomFilter === filter
                  ? "bg-[var(--primary-500)]/20 text-[var(--primary-200)]"
                  : "bg-[#14151a] text-[var(--text-3)]"
              }`}
            >
              {filter === "ALL" ? "Todos" : filter === "TEMPORARY" ? "TTL" : "Zero"}
            </button>
          ))}
        </div>

        <div className="mt-4 min-h-0 flex-1 space-y-1 overflow-y-auto pb-3">
          {isLoadingRooms ? (
            <p className="rounded-xl bg-[#14151a] px-3 py-2 text-sm text-[var(--text-3)]">
              Carregando salas...
            </p>
          ) : null}
          {!isLoadingRooms && orderedRooms.length === 0 ? (
            <p className="rounded-xl border border-dashed border-white/10 px-3 py-4 text-center text-sm text-[var(--text-3)]">
              Nenhuma sala ainda. Crie uma ou entre com o ID.
            </p>
          ) : null}
          {orderedRooms.map((room) => {
            const isActive = room.room_id === activeRoomId;
            const unreadCount = unreadByRoom[room.room_id] ?? 0;
            return (
              <button
                key={room.room_id}
                type="button"
                onClick={() => {
                  setActiveRoom(room.room_id);
                  setMobileView("chat");
                }}
                className={`w-full rounded-xl border p-2.5 text-left transition ${
                  isActive
                    ? "border-white/10 bg-[#212227]"
                    : "border-transparent bg-transparent hover:bg-[#14151a]"
                }`}
              >
                <div className="flex gap-3">
                  <RoomAvatar
                    name={room.name}
                    roomId={room.room_id}
                    objectKey={room.avatar_object_key}
                    token={token}
                    size="md"
                  />
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-base font-medium text-[var(--text-1)]">{room.name}</p>
                    <RoomTypeBadge type={room.type} zeroLogging={room.zero_logging} compact />
                    <p className="mt-1 truncate text-sm text-[var(--text-3)] [overflow-wrap:anywhere]">
                      {lastMessageByRoom[room.room_id]?.content
                        ? lastMessageByRoom[room.room_id]!.content.slice(0, 120)
                        : room.description?.trim() || "Abra para conversar"}
                    </p>
                  </div>
                  {!isActive && unreadCount > 0 ? (
                    <span
                      className="flex h-6 min-w-6 shrink-0 items-center justify-center self-center rounded-full bg-[var(--primary-500)] px-1.5 text-xs font-semibold text-white"
                      aria-label={`${unreadCount} mensagens não lidas`}
                    >
                      {unreadCount > 99 ? "99+" : unreadCount}
                    </span>
                  ) : null}
                </div>
              </button>
            );
          })}
        </div>
        </div>
      </aside>

      <main
        className={`flex min-h-0 min-w-0 flex-col overflow-hidden bg-[#07080d] ${
          mobileView === "chat" ? "flex" : "hidden lg:flex"
        }`}
      >
        <header className="flex items-center justify-between border-b border-white/8 px-4 py-3">
          <div className="flex min-w-0 items-start gap-2">
            <button
              type="button"
              className="mt-1 rounded-md border border-white/15 px-2 py-1 text-xs text-[var(--text-2)] lg:hidden"
              onClick={() => setMobileView("rooms")}
              aria-label="Voltar para salas"
            >
              <svg viewBox="0 0 24 24" className="h-4 w-4" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M15 18l-6-6 6-6" />
              </svg>
            </button>
            <button
              type="button"
              className="flex min-w-0 items-center gap-3 text-left"
              onClick={openRoomInfoPanel}
              disabled={!activeRoom}
            >
              {activeRoom ? (
                <RoomAvatar
                  name={activeRoom.name}
                  roomId={activeRoom.room_id}
                  objectKey={activeRoom.avatar_object_key}
                  token={token}
                  size="md"
                />
              ) : (
                <div className="h-11 w-11 shrink-0 rounded-2xl border border-dashed border-white/15 bg-[var(--bg-1)]" />
              )}
              <div className="min-w-0">
                <h1 className="truncate text-2xl font-semibold text-[var(--text-1)]">
                  {activeRoom?.name ?? "Selecione uma sala"}
                </h1>
                {activeRoom ? (
                  <RoomTypeBadge
                    type={activeRoom.type}
                    zeroLogging={activeRoom.zero_logging}
                    compact
                  />
                ) : null}
              </div>
            </button>
          </div>
          <div className="flex items-center gap-2">
            <span
              className={`rounded-lg px-2 py-1 text-xs ${
                connectionStatus === "connected"
                  ? "bg-[var(--success-500)]/20 text-[var(--success-500)]"
                  : "bg-[var(--warning-500)]/15 text-[var(--warning-500)]"
              }`}
            >
              {CONNECTION_LABELS[connectionStatus]}
            </span>
            {activeRoom ? (
              <button
                type="button"
                className="hidden rounded-md border border-white/20 px-2 py-1 text-xs text-[var(--text-2)] sm:inline-flex"
                onClick={() => {
                  setIsSelectionMode((current) => !current);
                  setSelectedMessageIds([]);
                }}
              >
                {isSelectionMode ? "Cancelar" : "Selecionar"}
              </button>
            ) : null}
            <button
              type="button"
              className="rounded-md border border-white/20 p-2 text-xs"
              onClick={openRoomInfoPanel}
              aria-label="Abrir informações da sala"
            >
              <svg viewBox="0 0 24 24" className="h-4 w-4" fill="none" stroke="currentColor" strokeWidth="2">
                <circle cx="12" cy="12" r="9" />
                <path d="M12 10v6M12 7h.01" />
              </svg>
            </button>
          </div>
        </header>

        {activeRoom?.zero_logging ? (
          <div className="border-b border-[var(--warning-500)]/30 bg-[var(--warning-500)]/10 py-2 text-center text-xs text-[var(--warning-500)]">
            Modo zero logging — nada do que você diz aqui é salvo no servidor
          </div>
        ) : null}
        {isSelectionMode ? (
          <div className="flex items-center justify-between border-b border-white/8 bg-[#10131a] px-4 py-2 text-xs text-[var(--text-2)]">
            <span>{selectedMessageIds.length} mensagem(ns) selecionada(s)</span>
            <button
              type="button"
              onClick={handleDeleteSelectedMessages}
              disabled={selectedMessageIds.length === 0}
              className="rounded-md border border-red-400/40 px-2 py-1 text-red-300 disabled:opacity-50"
            >
              Apagar selecionadas
            </button>
          </div>
        ) : null}

        <section className="flex min-h-0 min-w-0 flex-1 flex-col gap-3 overflow-x-hidden overflow-y-auto px-4 py-5">
          {visibleRoomMessages.map((message) => {
            const visual = getTemporaryVisualState(message, clock);
            const presence = presenceByUserId[message.authorId] ?? "offline";
            const displayHandle =
              memberHandleByUserId[message.authorId] ??
              (message.isOwn ? user.handle : message.authorHandle);
            const isDeletable = canDeleteMessage(message);
            const isSelected = selectedMessageIds.includes(message.id);
            return (
              <article
                key={message.id}
                className={`min-w-0 max-w-[min(100%,42rem)] shrink-0 rounded-2xl border px-4 py-3 shadow-sm sm:max-w-[min(100%,36rem)] lg:max-w-[min(100%,32rem)] ${
                  message.isOwn
                    ? "self-end border-[#6f00ff]/50 bg-gradient-to-br from-[#7a00ff] to-[#9400ff] text-white"
                    : "self-start border-white/10 bg-[#1a1b20] text-[var(--text-1)]"
                } ${visual.className} ${
                  isSelectionMode && isSelected ? "ring-2 ring-red-400/70" : ""
                }`}
                style={visual.style as CSSProperties}
                onClick={() => {
                  if (isSelectionMode && isDeletable) {
                    toggleMessageSelection(message.id);
                  }
                }}
              >
                <div className="mb-1 flex items-center gap-3 text-xs">
                  {isSelectionMode ? (
                    <input
                      type="checkbox"
                      checked={isSelected}
                      disabled={!isDeletable}
                      onChange={() => toggleMessageSelection(message.id)}
                      onClick={(event) => event.stopPropagation()}
                      aria-label={`Selecionar mensagem ${message.id}`}
                    />
                  ) : null}
                  <span className="opacity-80">{displayHandle}</span>
                  <span
                    className={`ml-auto h-2 w-2 shrink-0 rounded-full ${
                      presence === "online"
                        ? "bg-[var(--success-500)]"
                        : presence === "offline"
                          ? "bg-[var(--text-3)]"
                          : "bg-[var(--warning-500)]"
                    }`}
                    title={
                      presence === "online"
                        ? "Online"
                        : presence === "offline"
                          ? "Offline"
                          : "Desconhecido"
                    }
                    aria-hidden
                  />
                </div>
                {message.messageType === "IMAGE" && (message.imageObjectKey || message.imagePreviewUrl) ? (
                  <div className="space-y-2">
                    <MessageImageContent
                      objectKey={message.imageObjectKey ?? ""}
                      token={token}
                      previewUrl={message.imagePreviewUrl}
                      caption={message.content}
                    />
                    {message.content.trim() ? <MessageText content={message.content} /> : null}
                  </div>
                ) : (
                  <MessageText content={message.content} />
                )}
                <div className="mt-2 border-t border-white/15 pt-2 text-xs opacity-80">
                  <span>{new Date(message.createdAt).toLocaleTimeString("pt-BR", { hour: "2-digit", minute: "2-digit" })}</span>
                  <span className="ml-2">
                    {message.status === "read"
                      ? "✓✓ lida"
                      : message.status === "delivered"
                        ? "✓✓ entregue"
                        : message.status === "sent"
                          ? "✓ enviada"
                          : message.status === "failed"
                            ? "falha"
                            : "enviando"}
                  </span>
                  {visual.timeLeftSeconds !== null ? (
                    <span className="ml-2">{formatCountdown(visual.timeLeftSeconds)}</span>
                  ) : null}
                </div>
              </article>
            );
          })}
          {visibleRoomMessages.length === 0 ? (
            <div className="flex flex-col items-center justify-center gap-2 py-16 text-center">
              <div className="grid h-12 w-12 place-items-center rounded-2xl border border-white/10 bg-[var(--bg-1)] text-[var(--text-3)]">
                <svg viewBox="0 0 24 24" className="h-6 w-6" fill="none" stroke="currentColor" strokeWidth="1.5">
                  <path d="M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2z" />
                </svg>
              </div>
              <p className="text-base font-medium text-[var(--text-1)]">Comece a conversa</p>
              <p className="max-w-sm text-sm text-[var(--text-3)]">
                {activeRoom?.zero_logging
                  ? "Mensagens aparecem ao vivo e não ficam no histórico."
                  : activeRoom?.type === "TEMPORARY"
                    ? "Mensagens desta sala expiram após o TTL configurado."
                    : "Envie a primeira mensagem — ela chega em tempo real."}
              </p>
            </div>
          ) : null}
          <div ref={messagesEndRef} aria-hidden />
        </section>

        <footer className="border-t border-white/10 px-4 py-3">
          {connectionError ? (
            <p className="mb-2 text-xs text-[var(--warning-500)]">
              {connectionError === "content_too_long"
                ? `Mensagem muito longa (máx. ${MAX_MESSAGE_CONTENT_LENGTH} caracteres).`
                : connectionError}
            </p>
          ) : null}
          {activeRoomId && pendingAttachmentByRoom[activeRoomId] ? (
            <div className="mb-2 flex items-center gap-2 rounded-xl border border-white/10 bg-[var(--bg-1)] p-2">
              <img
                src={pendingAttachmentByRoom[activeRoomId]?.previewUrl}
                alt=""
                className="h-14 w-14 shrink-0 rounded-lg object-cover"
              />
              <p className="min-w-0 flex-1 truncate text-xs text-[var(--text-2)]">
                {pendingAttachmentByRoom[activeRoomId]?.file.name}
              </p>
              <button
                type="button"
                onClick={() => clearPendingAttachment(activeRoomId)}
                className="shrink-0 rounded-lg px-2 py-1 text-xs text-[var(--text-3)] transition hover:bg-white/10 hover:text-[var(--text-0)]"
                aria-label="Remover anexo"
              >
                ✕
              </button>
            </div>
          ) : null}
          <div className="flex items-center gap-2">
            <input
              id={messageFileInputId}
              type="file"
              accept="image/jpeg,image/png,image/webp,image/gif"
              className="sr-only"
              disabled={!activeRoomId || connectionStatus !== "connected"}
              onChange={(event) => {
                const file = event.target.files?.[0];
                if (file && activeRoomId) {
                  setPendingAttachment(activeRoomId, file);
                }
                event.target.value = "";
              }}
            />
            <button
              type="button"
              disabled={!activeRoomId || connectionStatus !== "connected"}
              onClick={() => document.getElementById(messageFileInputId)?.click()}
              className="grid h-11 w-11 shrink-0 place-items-center rounded-xl border border-white/15 text-[var(--text-2)] transition hover:bg-white/10 hover:text-[var(--text-0)] disabled:opacity-50"
              aria-label="Anexar imagem"
              title="Anexar imagem"
            >
              <svg viewBox="0 0 24 24" className="h-5 w-5" fill="none" stroke="currentColor" strokeWidth="1.8">
                <path d="M21.44 11.05l-9.19 9.19a6 6 0 01-8.49-8.49l9.19-9.19a4 4 0 015.66 5.66l-9.2 9.19a2 2 0 01-2.83-2.83l8.49-8.48" />
              </svg>
            </button>
            <input
              value={activeRoomId ? draftByRoom[activeRoomId] ?? "" : ""}
              maxLength={MAX_MESSAGE_CONTENT_LENGTH}
              onChange={(e) => activeRoomId && setDraft(activeRoomId, e.target.value)}
              onKeyDown={(event) => {
                if (event.key === "Enter" && !event.shiftKey && activeRoomId) {
                  event.preventDefault();
                  void sendComposer(activeRoomId);
                }
              }}
              disabled={!activeRoomId || connectionStatus !== "connected"}
              className="h-11 min-w-0 flex-1 rounded-xl border border-white/15 bg-[var(--bg-1)] px-3 text-base text-[var(--text-0)] outline-none transition focus:border-[var(--primary-400)] disabled:opacity-50"
              placeholder={
                !activeRoom
                  ? "Selecione uma sala"
                  : activeRoomId && pendingAttachmentByRoom[activeRoomId]
                    ? "Legenda (opcional)…"
                    : activeRoom.zero_logging
                      ? "Mensagem ao vivo — não será salva"
                      : "Escreva uma mensagem…"
              }
              aria-label="Mensagem"
            />
            <button
              type="button"
              onClick={() => activeRoomId && void sendComposer(activeRoomId)}
              disabled={
                !activeRoomId ||
                connectionStatus !== "connected" ||
                (!pendingAttachmentByRoom[activeRoomId ?? ""] &&
                  !(draftByRoom[activeRoomId ?? ""] ?? "").trim())
              }
              className="grid h-11 w-11 shrink-0 place-items-center rounded-xl bg-[var(--primary-500)] text-white transition hover:bg-[var(--primary-400)] disabled:cursor-not-allowed disabled:bg-[var(--bg-2)] disabled:text-[var(--text-3)]"
              aria-label="Enviar mensagem"
            >
              <svg viewBox="0 0 24 24" className="h-5 w-5 shrink-0 fill-current">
                <path d="M3.4 20.4l17.8-8.4c.9-.4.9-1.6 0-2L3.4 1.6c-.8-.4-1.7.4-1.4 1.3l2.1 6.4c.1.4.5.7.9.7h7.7a1 1 0 110 2H5a1 1 0 00-.9.7L2 19.1c-.3.9.6 1.7 1.4 1.3z" />
              </svg>
            </button>
          </div>
        </footer>
      </main>

      {isRightPanelOpen ? (
      <aside
        className={cn(
          "flex min-h-0 min-w-0 flex-col border-l border-white/8 bg-[var(--bg-0)]",
          mobileView === "info" ? "flex" : "hidden lg:flex",
        )}
      >
        <div className="flex shrink-0 items-center justify-between border-b border-white/8 px-4 py-3">
          <div className="flex items-center gap-2">
            <button
              type="button"
              className="rounded-md border border-white/15 px-2 py-1 text-xs text-[var(--text-2)] lg:hidden"
              onClick={() => setMobileView("chat")}
              aria-label="Voltar para chat"
            >
              <svg viewBox="0 0 24 24" className="h-4 w-4" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M15 18l-6-6 6-6" />
              </svg>
            </button>
            <div className="min-w-0">
              <h2 className="truncate text-lg font-semibold text-[var(--text-0)]">
                {activeRoom?.name ?? "Sala"}
              </h2>
              {activeRoom ? (
                <p
                  className={`truncate text-xs ${
                    activeRoom.type === "PUBLIC"
                      ? "text-[var(--success-500)]"
                      : activeRoom.type === "TEMPORARY"
                        ? "text-[var(--warning-500)]"
                        : activeRoom.type === "PRIVATE"
                          ? "text-cyan-400/80"
                          : "text-[var(--text-3)]"
                  }`}
                >
                  {formatRoomType(activeRoom.type)}
                  {activeRoom.zero_logging ? (
                    <span className="text-[var(--primary-300)]"> · Zero logging</span>
                  ) : null}
                </p>
              ) : (
                <p className="text-xs text-[var(--text-3)]">Selecione uma sala</p>
              )}
            </div>
          </div>
          <button
            type="button"
            className="rounded-md p-1 text-sm text-[var(--text-3)] transition hover:bg-white/10 hover:text-[var(--text-1)]"
            onClick={closeRoomInfoPanel}
            aria-label="Fechar painel"
          >
            ✕
          </button>
        </div>

        <div className="min-h-0 flex-1 overflow-y-auto px-4 py-4">
          {activeRoom ? (
            <div className="space-y-4">
              {token ? (
                <RoomMediaEditor
                  room={activeRoom}
                  token={token}
                  canEdit={canAddMembers}
                  onMediaUpdated={(patch) => patchRoomMedia(activeRoom.room_id, patch)}
                />
              ) : null}
              <div className="rounded-xl border border-white/10 bg-[var(--bg-1)] px-4 py-3 text-center">
                <RoomTypeBadge
                  type={activeRoom.type}
                  zeroLogging={activeRoom.zero_logging}
                />
                {activeRoom.description ? (
                  <p className="mt-3 text-sm leading-relaxed text-[var(--text-2)]">
                    {activeRoom.description}
                  </p>
                ) : (
                  <p className="mt-3 text-sm italic text-[var(--text-3)]">Sem descrição</p>
                )}
              </div>

              <div className="rounded-xl border border-white/10 bg-[var(--bg-1)] px-3 py-1">
                <p className="border-b border-white/8 py-2 text-xs font-medium uppercase tracking-wider text-[var(--text-3)]">
                  Detalhes
                </p>
                <MetaRow label="Tipo" value={formatRoomType(activeRoom.type)} />
                <MetaRow
                  label="Criada em"
                  value={formatDateTime(activeRoom.created_at)}
                />
                <MetaRow
                  label="Participantes"
                  value={String(roomMembers.length)}
                />
                <MetaRow
                  label="Mensagens visíveis"
                  value={String(visibleRoomMessages.length)}
                />
                {myMembership ? (
                  <MetaRow label="Seu papel" value={formatRole(myMembership.role)} />
                ) : null}
                {activeRoom.type === "TEMPORARY" ? (
                  <>
                    <MetaRow label="TTL" value={`${activeRoom.ttl}s`} />
                    {activeRoomExpiresInSeconds !== null ? (
                      <MetaRow
                        label="Expira em"
                        value={formatCountdown(activeRoomExpiresInSeconds)}
                      />
                    ) : null}
                  </>
                ) : null}
                {activeRoom.zero_logging ? (
                  <MetaRow label="Histórico" value="Não persiste (zero logging)" />
                ) : (
                  <MetaRow label="Histórico" value="Persistido no servidor" />
                )}
              </div>

              {activeRoom.type === "PUBLIC" ? (
                <div className="rounded-xl border border-[var(--success-500)]/20 bg-[var(--bg-1)] p-3">
                  <div className="flex items-center gap-2">
                    <svg
                      viewBox="0 0 24 24"
                      className="h-4 w-4 text-[var(--success-500)]"
                      fill="none"
                      stroke="currentColor"
                      strokeWidth="2"
                    >
                      <path d="M10 13a5 5 0 007.54.54l3-3a5 5 0 00-7.07-7.07l-1.72 1.71" />
                      <path d="M14 11a5 5 0 00-7.54-.54l-3 3a5 5 0 007.07 7.07l1.71-1.71" />
                    </svg>
                    <p className="text-sm font-medium text-[var(--text-0)]">Link de entrada</p>
                  </div>
                  <p className="mt-1.5 text-xs leading-relaxed text-[var(--text-3)]">
                    Salas públicas podem ser acessadas por ID. Compartilhe o UUID abaixo.
                  </p>
                  <p className="mt-2 rounded-lg border border-white/10 bg-[var(--bg-0)] px-2.5 py-2 font-mono text-[11px] leading-relaxed text-[var(--text-2)] break-all">
                    {activeRoom.room_id}
                  </p>
                  <button
                    type="button"
                    onClick={() => void handleCopyRoomId()}
                    className="mt-2 w-full rounded-lg border border-[var(--success-500)]/35 bg-[var(--success-500)]/10 px-3 py-2 text-sm font-medium text-[var(--success-500)] transition hover:bg-[var(--success-500)]/18"
                  >
                    Copiar ID da sala
                  </button>
                  {roomIdCopyFeedback ? (
                    <p className="mt-2 text-center text-xs text-[var(--success-500)]">
                      {roomIdCopyFeedback}
                    </p>
                  ) : null}
                </div>
              ) : null}

              <div>
                <h3 className="text-sm font-medium text-[var(--text-1)]">
                  Participantes
                  <span className="ml-1.5 font-normal text-[var(--text-3)]">({roomMembers.length})</span>
                </h3>
                <ul className="mt-2 space-y-2">
                  {roomMembers.map((member) => (
                    <ParticipantRow
                      key={member.user_id}
                      member={member}
                      isSelf={member.user_id === user.userId}
                      token={token}
                    />
                  ))}
                </ul>
              </div>

              <div className="rounded-xl border border-white/10 bg-[var(--bg-1)] p-3">
                <p className="text-sm font-medium text-[var(--text-1)]">Adicionar participante</p>
                {canAddMembers ? (
                  <>
                    <p className="mt-1 text-xs text-[var(--text-3)]">
                      Convide por handle (ex: joao#1234)
                    </p>
                    <div className="mt-3 flex gap-2">
                      <input
                        value={memberHandleInput}
                        onChange={(event) => setMemberHandleInput(event.target.value)}
                        placeholder="usuario#1234"
                        className="min-w-0 flex-1 rounded-lg border border-white/15 bg-[var(--bg-0)] px-3 py-2 text-sm text-[var(--text-0)] outline-none focus:border-[var(--primary-400)]"
                      />
                      <button
                        type="button"
                        onClick={() => void handleAddMember()}
                        disabled={!memberHandleInput.trim()}
                        className="shrink-0 rounded-lg bg-[var(--primary-500)] px-3 py-2 text-sm font-medium text-white transition hover:bg-[var(--primary-400)] disabled:cursor-not-allowed disabled:bg-[var(--bg-2)] disabled:text-[var(--text-3)]"
                      >
                        Adicionar
                      </button>
                    </div>
                    {addMemberFeedback ? (
                      <p className="mt-2 text-xs text-[var(--text-3)]">{addMemberFeedback}</p>
                    ) : null}
                  </>
                ) : (
                  <p className="mt-2 text-xs leading-relaxed text-[var(--text-3)]">
                    Apenas administradores podem convidar. Seu papel:{" "}
                    <span className="text-[var(--text-2)]">
                      {myMembership ? formatRole(myMembership.role) : "não membro"}
                    </span>
                    .
                  </p>
                )}
              </div>
            </div>
          ) : (
            <p className="mt-4 rounded-xl border border-dashed border-white/10 px-4 py-8 text-center text-sm text-[var(--text-3)]">
              Selecione uma sala para ver detalhes, participantes e o link de entrada.
            </p>
          )}
        </div>
      </aside>
      ) : null}
    </div>
  );
}
