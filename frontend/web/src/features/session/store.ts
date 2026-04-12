import { create } from "zustand";

type SessionState = {
  username: string;
  discriminator: string;
  setIdentity: (username: string, discriminator: string) => void;
};

export const useSessionStore = create<SessionState>((set) => ({
  username: "shadow",
  discriminator: "1827",
  setIdentity: (username, discriminator) => set({ username, discriminator }),
}));
