import type { ReactElement } from "react";
import { Navigate, Route, Routes } from "react-router-dom";
import { OnboardingPage } from "../pages/onboarding/page";
import { ChatPage } from "../pages/chat/page";
import { useSessionStore } from "../features/session/store";

function ProtectedRoute({ children }: { children: ReactElement }) {
  const isAuthenticated = useSessionStore((state) => state.isAuthenticated);
  const isBootstrapping = useSessionStore((state) => state.isBootstrapping);
  const token = useSessionStore((state) => state.token);
  const user = useSessionStore((state) => state.user);

  if (isBootstrapping) {
    return (
      <main className="grid min-h-dvh place-items-center bg-[var(--bg-0)] text-sm text-[var(--text-2)]">
        Validando sessão...
      </main>
    );
  }

  if (!isAuthenticated || !token || !user) {
    return <Navigate to="/" replace />;
  }

  return children;
}

export function AppRouter() {
  return (
    <Routes>
      <Route path="/" element={<OnboardingPage />} />
      <Route
        path="/chat"
        element={
          <ProtectedRoute>
            <ChatPage />
          </ProtectedRoute>
        }
      />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
