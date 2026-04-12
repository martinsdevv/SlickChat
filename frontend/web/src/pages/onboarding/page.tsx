import { useNavigate } from "react-router-dom";

export function OnboardingPage() {
  const navigate = useNavigate();

  return (
    <main className="mx-auto flex min-h-dvh w-full max-w-md flex-col justify-center px-6">
      <h1 className="mb-2 text-2xl font-semibold">SlickChat</h1>
      <p className="mb-6 text-sm text-[var(--text-2)]">
        Privacidade e efemeridade por padrão.
      </p>

      <label className="mb-2 text-sm">Username</label>
      <input
        defaultValue="shadow"
        className="mb-4 rounded-lg border border-white/20 bg-[var(--bg-1)] px-3 py-2 text-sm outline-none focus:border-[var(--primary-400)]"
      />

      <button
        onClick={() => navigate("/chat")}
        className="rounded-lg bg-[var(--primary-500)] px-4 py-2 text-sm font-medium"
      >
        Entrar
      </button>
    </main>
  );
}
