import { create } from "zustand";

export type ChatMessage = {
  id: string;
  author: string;
  content: string;
  status: "sending" | "sent" | "delivered" | "read";
};

type ChatState = {
  activeRoomName: string;
  messages: ChatMessage[];
  draft: string;
  setDraft: (value: string) => void;
  sendDraft: () => void;
};

export const useChatStore = create<ChatState>((set, get) => ({
  activeRoomName: "Sala temporária",
  draft: "",
  messages: [
    {
      id: "1",
      author: "ghost#4451",
      content: "Bem-vindo ao SlickChat",
      status: "read",
    },
  ],
  setDraft: (value) => set({ draft: value }),
  sendDraft: () => {
    const content = get().draft.trim();
    if (!content) return;

    const msg: ChatMessage = {
      id: crypto.randomUUID(),
      author: "shadow#1827",
      content,
      status: "sending",
    };

    set((s) => ({ draft: "", messages: [...s.messages, msg] }));

    setTimeout(() => {
      set((s) => ({
        messages: s.messages.map((m) =>
          m.id === msg.id ? { ...m, status: "sent" } : m,
        ),
      }));
    }, 600);
  },
}));
