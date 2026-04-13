import { useEffect, useMemo, useState, type CSSProperties } from "react";
import { useChatStore, type ChatMessage } from "../../../features/chat/store";
import { filterRooms, useRoomsStore } from "../../../features/rooms/store";
import { useSessionStore } from "../../../features/session/store";
import { useUIStore } from "../../../features/ui/store";

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
    loadRooms,
    createRoom,
    joinRoom,
    addMember,
    loadMembers,
    reset,
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
    sendMessage,
    markMessageRead,
  } = useChatStore();
  const {
    isSidebarOpen,
    isRightPanelOpen,
    mobileView,
    toggleSidebar,
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
  const myMembership = useMemo(
    () => roomMembers.find((member) => member.user_id === user.userId) ?? null,
    [roomMembers, user.userId],
  );
  const canAddMembers = myMembership?.role === "ADMIN";

  useEffect(() => {
    if (!token) {
      return;
    }
    reset();
    void loadRooms(token).catch(() => undefined);
  }, [loadRooms, reset, token]);

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

  if (!user) {
    return null;
  }

  return (
    <div className="grid h-dvh grid-cols-1 overflow-hidden bg-[#07080d] lg:grid-cols-[320px_1fr_340px]">
      <aside
        className={`${
          isSidebarOpen ? "block" : "hidden"
        } min-h-0 border-r border-white/8 bg-[#0d0e12] px-4 py-4 ${
          mobileView === "chat" ? "hidden lg:block" : "block"
        }`}
      >
        <div className="flex h-full min-h-0 flex-col">
        <h1 className="text-3xl font-semibold tracking-wide text-[var(--text-0)]">
          SlickChat
        </h1>
        <div className="mt-4 grid gap-2">
          <button
            disabled={isCreatingRoom || isCreateFormOpen}
            onClick={() => {
              resetCreateRoomForm();
              setIsCreateFormOpen(true);
            }}
            className="w-full rounded-xl bg-gradient-to-r from-[#7a00ff] to-[#8d2cff] px-4 py-2.5 text-left text-base font-medium text-white disabled:opacity-60"
          >
            Criar Sala
          </button>
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
          {filteredRooms.map((room) => {
            const isActive = room.room_id === activeRoomId;
            return (
              <button
                key={room.room_id}
                type="button"
                onClick={() => {
                  setActiveRoom(room.room_id);
                  setMobileView("chat");
                }}
                className={`w-full rounded-xl border px-3 py-2 text-left transition ${
                  isActive
                    ? "border-white/10 bg-[#212227]"
                    : "border-transparent bg-transparent hover:bg-[#14151a]"
                }`}
              >
                <p className="text-base font-medium text-[var(--text-1)]">{room.name}</p>
                <div className="mt-1 flex items-center gap-2 text-[10px] uppercase tracking-wide">
                  <span
                    className={`rounded px-1.5 py-0.5 ${
                      room.type === "PRIVATE"
                        ? "bg-fuchsia-500/20 text-fuchsia-300"
                        : room.type === "PUBLIC"
                          ? "bg-emerald-500/20 text-emerald-300"
                          : "bg-amber-500/20 text-amber-300"
                    }`}
                  >
                    {room.type}
                  </span>
                  {room.zero_logging ? (
                    <span className="rounded bg-[#7a00ff]/20 px-1.5 py-0.5 text-[#d3b3ff]">ZERO</span>
                  ) : null}
                </div>
                <p className="mt-1 truncate text-sm text-[var(--text-3)]">
                  {lastMessageByRoom[room.room_id]?.content ?? "Sem mensagens ainda"}
                </p>
                {room.description ? (
                  <p className="mt-1 line-clamp-2 text-xs text-[var(--text-3)]">{room.description}</p>
                ) : null}
              </button>
            );
          })}
        </div>
        </div>
      </aside>

      <main
        className={`flex min-h-0 flex-col bg-[#07080d] ${
          mobileView === "rooms" ? "hidden lg:flex" : "flex"
        }`}
      >
        <header className="flex items-center justify-between border-b border-white/8 px-4 py-3">
          <div>
            <h1 className="text-2xl font-semibold text-[var(--text-1)]">
              {activeRoom?.name ?? "Selecione uma sala"}
            </h1>
            {activeRoom ? (
              <div className="mt-1 flex items-center gap-2 text-[10px] uppercase tracking-wide">
                <span
                  className={`rounded px-1.5 py-0.5 ${
                    activeRoom.type === "PRIVATE"
                      ? "bg-fuchsia-500/20 text-fuchsia-300"
                      : activeRoom.type === "PUBLIC"
                        ? "bg-emerald-500/20 text-emerald-300"
                        : "bg-amber-500/20 text-amber-300"
                  }`}
                >
                  {activeRoom.type}
                </span>
                {activeRoom.zero_logging ? (
                  <span className="rounded bg-[#7a00ff]/20 px-1.5 py-0.5 text-[#d3b3ff]">ZERO</span>
                ) : null}
              </div>
            ) : null}
          </div>
          <div className="flex items-center gap-2">
            <span
              className={`rounded-lg px-2 py-1 text-xs ${
                connectionStatus === "connected"
                  ? "bg-[var(--success-500)]/20 text-[var(--success-500)]"
                  : "bg-[var(--warning-500)]/15 text-[var(--warning-500)]"
              }`}
            >
              {connectionStatus}
            </span>
            <button
              type="button"
              className="rounded-md border border-white/15 px-2 py-1 text-xs text-[var(--text-2)] lg:hidden"
              onClick={() => setMobileView("rooms")}
            >
              Voltar
            </button>
            <button
              className="rounded-md border border-white/20 px-3 py-2 text-xs"
              onClick={toggleSidebar}
            >
              Menu
            </button>
            <button
              className="rounded-md border border-white/20 px-3 py-2 text-xs"
              onClick={toggleRightPanel}
            >
              Painel
            </button>
          </div>
        </header>

        {activeRoom?.zero_logging ? (
          <div className="border-b border-[#674b00] bg-[#2f2300]/45 py-1 text-center text-xs text-[#f7bf31]">
            Aviso: nada do que voce diz aqui é salvo
          </div>
        ) : null}

        <section className="min-h-0 flex-1 space-y-3 overflow-x-hidden overflow-y-auto px-4 py-5">
          {visibleRoomMessages.map((message) => {
            const visual = getTemporaryVisualState(message, clock);
            const presence = presenceByUserId[message.authorId] ?? "unknown";
            const displayHandle =
              memberHandleByUserId[message.authorId] ??
              (message.isOwn ? user.handle : message.authorHandle);
            return (
              <article
                key={message.id}
                className={`min-w-0 w-fit max-w-[78%] sm:max-w-[70%] lg:max-w-[62%] rounded-2xl border px-4 py-3 shadow-sm ${
                  message.isOwn
                    ? "ml-auto border-[#6f00ff]/50 bg-gradient-to-br from-[#7a00ff] to-[#9400ff] text-white"
                    : "border-white/10 bg-[#1a1b20] text-[var(--text-1)]"
                } ${visual.className}`}
                style={visual.style as CSSProperties}
              >
                <div className="mb-1 flex items-center gap-3 text-xs">
                  <span className="opacity-80">{displayHandle}</span>
                  <span
                    className={`ml-auto ${
                      presence === "online"
                        ? "text-[var(--success-500)]"
                        : presence === "offline"
                          ? "text-[var(--text-3)]"
                          : "text-[var(--warning-500)]"
                    }`}
                  >
                    {presence}
                  </span>
                </div>
                <p className="whitespace-pre-wrap break-all text-base leading-relaxed">{message.content}</p>
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
            <p className="rounded-xl border border-dashed border-white/10 bg-[#0e0f13] px-3 py-2 text-sm text-[var(--text-3)]">
              Sem mensagens ainda nesta sala.
            </p>
          ) : null}
        </section>

        <footer className="border-t border-white/10 px-4 py-3">
          <label className="mb-2 block text-xs text-[var(--text-3)]">
            {activeRoom?.zero_logging
              ? "Enviar mensagem (não será salva)"
              : "Enviar mensagem"}
          </label>
          {connectionError ? (
            <p className="mb-2 text-xs text-[var(--warning-500)]">{connectionError}</p>
          ) : null}
          <div className="flex items-center gap-2">
            <button
              type="button"
              className="grid h-11 w-11 place-items-center rounded-xl border border-white/10 text-[var(--text-2)] transition-all duration-200 hover:bg-white/10"
              aria-label="Anexar"
            >
              <svg viewBox="0 0 24 24" className="h-5 w-5" fill="none" stroke="currentColor" strokeWidth="1.8">
                <path d="M21.44 11.05l-8.49 8.49a5 5 0 01-7.07-7.07l9.19-9.19a3.5 3.5 0 114.95 4.95L9.77 18.48a2 2 0 01-2.83-2.83l8.49-8.49" />
              </svg>
            </button>
            <input
              value={activeRoomId ? draftByRoom[activeRoomId] ?? "" : ""}
              onChange={(e) => activeRoomId && setDraft(activeRoomId, e.target.value)}
              onKeyDown={(event) => {
                if (event.key === "Enter" && activeRoomId) {
                  event.preventDefault();
                  sendMessage(activeRoomId);
                }
              }}
              className="h-11 w-full rounded-xl border border-white/15 bg-[#1a1b20] px-3 text-base text-[var(--text-0)] outline-none transition focus:border-[var(--primary-400)]"
              placeholder={
                activeRoom?.zero_logging ? "Enviar mensagem (não será salva)" : "Enviar mensagem"
              }
              aria-label="Mensagem"
            />
            <button
              type="button"
              className="grid h-11 w-11 place-items-center rounded-xl border border-white/10 text-[var(--text-2)] transition-all duration-200 hover:bg-white/10"
              aria-label="Áudio"
            >
              <svg viewBox="0 0 24 24" className="h-5 w-5" fill="none" stroke="currentColor" strokeWidth="1.8">
                <rect x="9" y="3" width="6" height="12" rx="3" />
                <path d="M5 11a7 7 0 0014 0M12 18v3M8 21h8" />
              </svg>
            </button>
            <button
              onClick={() => activeRoomId && sendMessage(activeRoomId)}
              disabled={!activeRoomId}
              className="grid h-11 w-11 place-items-center rounded-xl bg-[#1f2128] text-[var(--text-1)] transition-all duration-200 hover:brightness-110 disabled:opacity-60"
              aria-label="Enviar"
            >
              <svg viewBox="0 0 24 24" className="h-5 w-5" fill="currentColor">
                <path d="M3.4 20.4l17.8-8.4c.9-.4.9-1.6 0-2L3.4 1.6c-.8-.4-1.7.4-1.4 1.3l2.1 6.4c.1.4.5.7.9.7h7.7a1 1 0 110 2H5a1 1 0 00-.9.7L2 19.1c-.3.9.6 1.7 1.4 1.3z" />
              </svg>
            </button>
          </div>
        </footer>
      </main>

      <aside className={`${isRightPanelOpen ? "hidden lg:block" : "hidden"} border-l border-white/8 bg-[#0c0d12] p-4`}>
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-lg font-semibold text-[var(--text-1)]">Info da Sala</h2>
          <button
            type="button"
            className="text-sm text-[var(--text-3)]"
            onClick={toggleRightPanel}
          >
            ✕
          </button>
        </div>
        {activeRoom ? (
          <>
            <p className="text-xl text-[var(--text-1)]">{activeRoom.name}</p>
            <div className="mt-2 flex items-center gap-2 text-[10px] uppercase tracking-wide">
              <span
                className={`rounded px-1.5 py-0.5 ${
                  activeRoom.type === "PRIVATE"
                    ? "bg-fuchsia-500/20 text-fuchsia-300"
                    : activeRoom.type === "PUBLIC"
                      ? "bg-emerald-500/20 text-emerald-300"
                      : "bg-amber-500/20 text-amber-300"
                }`}
              >
                {activeRoom.type}
              </span>
              {activeRoom.zero_logging ? (
                <span className="rounded bg-[#7a00ff]/20 px-1.5 py-0.5 text-[#d3b3ff]">ZERO</span>
              ) : null}
            </div>
            {activeRoom.description ? (
              <p className="mt-2 text-sm text-[var(--text-2)]">{activeRoom.description}</p>
            ) : null}

            <div className="mt-6">
              <h3 className="text-sm font-medium text-[var(--text-1)]">
                Participantes ({roomMembers.length})
              </h3>
              <ul className="mt-2 space-y-2">
                {roomMembers.map((member) => (
                  <li key={member.user_id} className="rounded-lg bg-[#15161c] px-3 py-2">
                    <p className="text-sm text-[var(--text-1)]">{member.handle}</p>
                    <p className="text-xs text-[var(--text-3)]">{member.role}</p>
                  </li>
                ))}
              </ul>
            </div>

            <div className="mt-6 rounded-xl border border-white/10 bg-[#13141a] p-3">
              <p className="text-sm text-[var(--text-1)]">Adicionar participante</p>
              {canAddMembers ? (
                <>
                  <p className="mt-1 text-xs text-[var(--text-3)]">Digite o handle do usuário (ex: joao#1234)</p>
                  <div className="mt-3 flex gap-2">
                    <input
                      value={memberHandleInput}
                      onChange={(event) => setMemberHandleInput(event.target.value)}
                      placeholder="usuario#1234"
                      className="w-full rounded-lg border border-white/15 bg-[#1a1b20] px-3 py-2 text-sm text-[var(--text-0)] outline-none focus:border-[var(--primary-400)]"
                    />
                    <button
                      type="button"
                      onClick={() => void handleAddMember()}
                      className="rounded-lg border border-white/15 px-3 py-2 text-sm text-[var(--text-2)] transition-all duration-200 hover:bg-white/10"
                    >
                      Adicionar
                    </button>
                  </div>
                  {addMemberFeedback ? (
                    <p className="mt-2 text-xs text-[var(--text-3)]">{addMemberFeedback}</p>
                  ) : null}
                </>
              ) : (
                <p className="mt-1 text-xs text-[var(--text-3)]">
                  Somente admins podem adicionar usuários. Seu papel nesta sala: {myMembership?.role ?? "desconhecido"}.
                </p>
              )}
            </div>

            <div className="mt-6 rounded-xl border border-white/10 bg-[#13141a] p-3">
              <p className="text-sm text-[var(--text-2)]">Conectado como</p>
              <p className="text-base text-[var(--text-0)]">{user.handle}</p>
              <button
                type="button"
                onClick={() => logout().then(() => disconnect())}
                className="mt-3 w-full rounded-lg border border-white/15 px-3 py-2 text-sm text-[var(--text-2)]"
              >
                Sair da sessão
              </button>
            </div>
          </>
        ) : (
          <p className="text-sm text-[var(--text-3)]">
            Selecione uma sala para visualizar detalhes e participantes.
          </p>
        )}
      </aside>
    </div>
  );
}
