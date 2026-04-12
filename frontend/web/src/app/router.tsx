import { Navigate, Route, Routes } from "react-router-dom";
import { OnboardingPage } from "../pages/onboarding/page";
import { ChatPage } from "../pages/chat/page";

export function AppRouter() {
  return (
    <Routes>
      <Route path="/" element={<OnboardingPage />} />
      <Route path="/chat" element={<ChatPage />} />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
