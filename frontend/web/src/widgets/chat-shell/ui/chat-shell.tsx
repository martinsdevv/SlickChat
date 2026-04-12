import { useChatStore } from "../../../features/chat/store";
import { useSessionStore } from "../../../features/session/store";
import { useUIStore } from "../../../features/ui/store";

const rooms = ["Sala normal", "Sala temporária", "Sala privada efêmera"];

export function ChatShell() {
  const { username, discriminator } = useSessionStore();
  const { activeRoomName, messages, draft, setDraft, sendDraft } = useChatStore();
  const { isSidebarOpen, isRightPanelOpen, toggleSidebar, toggleRightPanel } = useUIStore();

  return (
    <div className="grid min-h-dvh grid-cols-1 lg:grid-cols-[280px_1fr_320px]">
      <aside className={`${isSidebarOpen ? "block" : "hidden"} border-r border-white/10 bg-[var(--bg-1)] p-4 lg:block`}>
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-sm font-semibold text-[var(--text-1)]">Conversas</h2>
          <button
            className="rounded-md bg-[var(--primary-500)] px-3 py-2 text-xs font-medium"
            onClick={toggleSidebar}
          >
            Ocultar
          </button>
        </div>
        <ul className="space-y-2">
          {rooms.map((room) => (
            <li
              key={room}
              className={`cursor-pointer rounded-lg px-3 py-2 text-sm ${
                room === activeRoomName
                  ? "bg-[var(--primary-400)]/20 text-[var(--text-0)]"
                  : "bg-[var(--bg-2)] text-[var(--text-2)]"
              }`}
            >
              {room}
            </li>
          ))}
        </ul>
      </aside>

      <main className="flex min-h-0 flex-col bg-[var(--bg-0)]">
        <header className="flex items-center justify-between border-b border-white/10 px-4 py-3">
          <div>
            <h1 className="text-base font-semibold">{activeRoomName}</h1>
            <p className="text-xs text-[var(--text-2)]">Modo efêmero ativo</p>
          </div>
          <div className="flex gap-2">
            <button className="rounded-md border border-white/20 px-3 py-2 text-xs" onClick={toggleSidebar}>
              Sidebar
            </button>
            <button className="rounded-md border border-white/20 px-3 py-2 text-xs" onClick={toggleRightPanel}>
              Painel
            </button>
          </div>
        </header>

        <section className="min-h-0 flex-1 space-y-3 overflow-y-auto p-4">
          {messages.map((message) => (
            <article key={message.id} className="rounded-lg bg-[var(--bg-1)] p-3">
              <div className="mb-1 flex items-center justify-between text-xs text-[var(--text-2)]">
                <span>{message.author}</span>
                <span>{message.status}</span>
              </div>
              <p className="text-sm">{message.content}</p>
            </article>
          ))}
        </section>

        <footer className="border-t border-white/10 p-4">
          <label className="mb-2 block text-xs text-[var(--text-2)]">
            Enviar mensagem (não será salva)
          </label>
          <div className="flex gap-2">
            <input
              value={draft}
              onChange={(e) => setDraft(e.target.value)}
              className="w-full rounded-lg border border-white/20 bg-[var(--bg-1)] px-3 py-2 text-sm outline-none focus:border-[var(--primary-400)]"
              placeholder="Digite sua mensagem"
              aria-label="Mensagem"
            />
            <button
              onClick={sendDraft}
              className="rounded-lg bg-[var(--primary-500)] px-4 py-2 text-sm font-medium"
            >
              Enviar
            </button>
          </div>
        </footer>
      </main>

      <aside className={`${isRightPanelOpen ? "hidden lg:block" : "hidden"} border-l border-white/10 bg-[var(--bg-1)] p-4`}>
        <h2 className="mb-2 text-sm font-semibold">Detalhes da sala</h2>
        <p className="text-xs text-[var(--text-2)]">
          Usuário ativo: {username}#{discriminator}
        </p>
        <p className="mt-2 text-xs text-[var(--text-2)]">
          TTL padrão: 5 minutos
        </p>
      </aside>
    </div>
  );
}
