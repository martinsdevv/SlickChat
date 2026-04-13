import { useEffect, useMemo, useState, type CSSProperties } from "react";
import { useChatStore, type ChatMessage } from "../../../features/chat/store";
import { filterRooms, useRoomsStore } from "../../../features/rooms/store";
import { useSessionStore } from "../../../features/session/store";
import { useUIStore } from "../../../features/ui/store";

function formatRoomMode(roomType: string, zeroLogging: boolean) {
  if (zeroLogging) {
    return "Nada do que você diz aqui é salvo";
  }
  if (roomType === "TEMPORARY") {
    return "Mensagens desaparecem em instantes";
  }
  return "Mensagens ficam salvas";
}

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
    loadMembers,
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

  const activeRoom = rooms.find((room) => room.room_id === activeRoomId) ?? null;
  const roomMessages = useMemo(
    () => (activeRoomId ? messagesByRoom[activeRoomId] ?? [] : []),
    [activeRoomId, messagesByRoom],
  );
  const roomMembers = activeRoomId ? membersByRoom[activeRoomId] ?? [] : [];
  const filteredRooms = useMemo(() => filterRooms(rooms, roomFilter), [roomFilter, rooms]);

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
    if (!activeRoomId || !token) {
      return;
    }

    void joinRoom(token, activeRoomId)
      .catch(() => undefined)
      .finally(() => {
        void loadRoomHistory(activeRoomId, token);
        void loadMembers(token, activeRoomId).catch(() => undefined);
      });
  }, [activeRoomId, joinRoom, loadMembers, loadRoomHistory, token]);

  useEffect(() => {
    const timer = window.setInterval(() => setClock(Date.now()), 1000);
    return () => window.clearInterval(timer);
  }, []);

  useEffect(() => {
    if (!activeRoomId) {
      return;
    }

    roomMessages
      .filter((message) => !message.isOwn && message.status !== "read")
      .forEach((message) => {
        markMessageRead(activeRoomId, message.id);
      });
  }, [activeRoomId, markMessageRead, roomMessages]);

  async function handleCreateRoom(roomType: "PUBLIC" | "TEMPORARY") {
    if (!token) {
      return;
    }

    const now = new Date().toISOString().slice(11, 16).replace(":", "");
    const created = await createRoom(token, {
      name: roomType === "TEMPORARY" ? `Privacidade-${now}` : `Geral-${now}`,
      type: roomType,
      ttl: roomType === "TEMPORARY" ? 60 : 0,
      paranoid_mode: false,
      zero_logging: roomType === "TEMPORARY",
    }).catch(() => null);
    if (!created) {
      return;
    }
    setActiveRoom(created.room_id);
  }

  if (!user) {
    return null;
  }

  return (
    <div className="grid min-h-dvh grid-cols-1 bg-[#07080d] lg:grid-cols-[320px_1fr_340px]">
      <aside
        className={`${
          isSidebarOpen ? "block" : "hidden"
        } border-r border-white/8 bg-[#0d0e12] px-4 py-4 ${
          mobileView === "chat" ? "hidden lg:block" : "block"
        }`}
      >
        <h1 className="text-3xl font-semibold tracking-wide text-[var(--text-0)]">
          SlickChat
        </h1>
        <div className="mt-4 grid gap-2">
          <button
            disabled={isCreatingRoom}
            onClick={() => handleCreateRoom("PUBLIC")}
            className="w-full rounded-xl bg-gradient-to-r from-[#7a00ff] to-[#8d2cff] px-4 py-2.5 text-left text-base font-medium text-white disabled:opacity-60"
          >
            Nova Sala
          </button>
          <button
            disabled={isCreatingRoom}
            onClick={() => handleCreateRoom("TEMPORARY")}
            className="w-full rounded-xl border border-white/10 bg-[#1c1d22] px-4 py-2.5 text-left text-base font-medium text-[var(--text-1)] disabled:opacity-60"
          >
            Nova Sala Temporária
          </button>
        </div>

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

        <div className="mt-4 space-y-1 overflow-y-auto pb-3">
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
                <p className="text-base font-medium text-[var(--text-1)]">
                  {room.type === "TEMPORARY" ? "Temp " : room.zero_logging ? "Priv " : "Sala "}
                  {room.name}
                </p>
                <p className="text-sm text-[var(--text-3)]">
                  {formatRoomMode(room.type, room.zero_logging)}
                </p>
              </button>
            );
          })}
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
              {activeRoom?.type === "TEMPORARY" ? "Temporaria - " : activeRoom?.zero_logging ? "Privada - " : ""}
              {activeRoom?.name ?? "Selecione uma sala"}
            </h1>
            <p className="text-sm text-[var(--text-3)]">
              {activeRoom
                ? formatRoomMode(activeRoom.type, activeRoom.zero_logging)
                : "Escolha uma sala para começar"}
            </p>
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
            Aviso: nada do que voce diz aqui e salvo
          </div>
        ) : null}

        <section className="min-h-0 flex-1 space-y-3 overflow-y-auto px-4 py-5">
          {roomMessages.map((message) => {
            const visual = getTemporaryVisualState(message, clock);
            const presence = presenceByUserId[message.authorId] ?? "unknown";
            return (
              <article
                key={message.id}
                className={`max-w-[85%] rounded-2xl border px-4 py-3 shadow-sm ${
                  message.isOwn
                    ? "ml-auto border-[#6f00ff]/50 bg-gradient-to-br from-[#7a00ff] to-[#9400ff] text-white"
                    : "border-white/10 bg-[#1a1b20] text-[var(--text-1)]"
                } ${visual.className}`}
                style={visual.style as CSSProperties}
              >
                <div className="mb-1 flex items-center justify-between gap-4 text-xs">
                  <span className="opacity-80">{message.authorHandle}</span>
                  <span
                    className={`${
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
                <p className="text-base leading-relaxed">{message.content}</p>
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
          {roomMessages.length === 0 ? (
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
          <div className="flex items-end gap-2">
            <button
              type="button"
              className="rounded-xl border border-white/10 px-3 py-2 text-[var(--text-2)]"
              aria-label="Anexar"
            >
              Anexar
            </button>
            <input
              value={activeRoomId ? draftByRoom[activeRoomId] ?? "" : ""}
              onChange={(e) => activeRoomId && setDraft(activeRoomId, e.target.value)}
              className="w-full rounded-xl border border-white/15 bg-[#1a1b20] px-3 py-2 text-base text-[var(--text-0)] outline-none transition focus:border-[var(--primary-400)]"
              placeholder={
                activeRoom?.zero_logging ? "Enviar mensagem (não será salva)" : "Enviar mensagem"
              }
              aria-label="Mensagem"
            />
            <button
              type="button"
              className="rounded-xl border border-white/10 px-3 py-2 text-[var(--text-2)]"
              aria-label="Áudio"
            >
              Audio
            </button>
            <button
              onClick={() => activeRoomId && sendMessage(activeRoomId)}
              disabled={!activeRoomId}
              className="rounded-xl bg-[#1f2128] px-4 py-2 text-sm font-medium text-[var(--text-1)] disabled:opacity-60"
            >
              Enviar
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
            <p className="text-sm text-[var(--text-3)]">
              {formatRoomMode(activeRoom.type, activeRoom.zero_logging)}
            </p>

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
