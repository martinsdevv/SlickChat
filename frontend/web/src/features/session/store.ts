import { create } from "zustand";
import { persist } from "zustand/middleware";
import { apiRequest } from "../../shared/api/http-client";
import { useRoomsStore } from "../rooms/store";
import type {
  LoginRequest,
  LoginResponse,
  MeResponse,
  RegisterRequest,
  RegisterResponse,
} from "../../shared/api/types";

export type SessionUser = {
  userId: string;
  handle: string;
  username: string;
  discriminator: string;
};

type RegisterResult = {
  recoveryKey: string;
  handle: string;
};

type SessionState = {
  token: string | null;
  expiresAt: string | null;
  user: SessionUser | null;
  isBootstrapping: boolean;
  isAuthenticated: boolean;
  register: (input: RegisterRequest) => Promise<RegisterResult>;
  login: (input: LoginRequest) => Promise<void>;
  bootstrap: () => Promise<void>;
  logout: () => Promise<void>;
};

function parseHandle(handle: string | null | undefined) {
  const safeHandle = typeof handle === "string" ? handle : "";
  const [username, discriminator = "0000"] = safeHandle.split("#");
  return { username, discriminator };
}

function mapMeToUser(me: MeResponse): SessionUser {
  const { username, discriminator } = parseHandle(me.handle);
  return {
    userId: me.user_id,
    handle: me.handle,
    username,
    discriminator,
  };
}

export const useSessionStore = create<SessionState>()(
  persist(
    (set, get) => ({
      token: null,
      expiresAt: null,
      user: null,
      isBootstrapping: false,
      isAuthenticated: false,
      register: async (input) => {
        const response = await apiRequest<RegisterResponse>("/register", {
          method: "POST",
          body: input,
        });

        return {
          recoveryKey: response.recovery_key,
          handle: response.handle,
        };
      },
      login: async (input) => {
        useRoomsStore.getState().reset();

        const response = await apiRequest<LoginResponse>("/login", {
          method: "POST",
          body: input,
        });

        const { username, discriminator } = parseHandle(response.handle);
        set({
          token: response.token,
          expiresAt: response.expires_at,
          user: {
            userId: response.user_id,
            handle: response.handle,
            username,
            discriminator,
          },
          isAuthenticated: true,
        });
      },
      bootstrap: async () => {
        const token = get().token;
        if (!token) {
          set({
            isAuthenticated: false,
            user: null,
            expiresAt: null,
          });
          return;
        }

        set({ isBootstrapping: true });
        try {
          const me = await apiRequest<MeResponse>("/users/me", { token });
          set({
            user: mapMeToUser(me),
            isAuthenticated: true,
          });
        } catch {
          set({
            token: null,
            expiresAt: null,
            user: null,
            isAuthenticated: false,
          });
        } finally {
          set({ isBootstrapping: false });
        }
      },
      logout: async () => {
        const token = get().token;
        if (token) {
          try {
            await apiRequest<void>("/logout", { method: "POST", token });
          } catch {
            // Session cleanup on the client is still required when backend is unavailable.
          }
        }

        set({
          token: null,
          expiresAt: null,
          user: null,
          isAuthenticated: false,
        });
        useRoomsStore.getState().reset();
      },
    }),
    {
      name: "slickchat-session",
      partialize: (state) => ({
        token: state.token,
        expiresAt: state.expiresAt,
        user: state.user,
      }),
    },
  ),
);
