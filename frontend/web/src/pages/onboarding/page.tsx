import { useNavigate } from "react-router-dom";
import { useEffect, useMemo, useState, type FormEvent } from "react";
import { jsPDF } from "jspdf";
import { ApiError } from "../../shared/api/types";
import { useSessionStore } from "../../features/session/store";

type AuthMode = "register" | "login";

export function OnboardingPage() {
  const navigate = useNavigate();
  const register = useSessionStore((state) => state.register);
  const login = useSessionStore((state) => state.login);
  const user = useSessionStore((state) => state.user);
  const isAuthenticated = useSessionStore((state) => state.isAuthenticated);
  const [mode, setMode] = useState<AuthMode>("register");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [handle, setHandle] = useState("");
  const [paranoidMode, setParanoidMode] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [recoveryState, setRecoveryState] = useState<{
    handle: string;
    recoveryKey: string;
  } | null>(null);
  const [copyFeedback, setCopyFeedback] = useState<string | null>(null);
  const [isCopyingRecoveryKey, setIsCopyingRecoveryKey] = useState(false);

  const formTitle = useMemo(
    () => (mode === "register" ? "Criar Conta" : "Entrar"),
    [mode],
  );
  const isRegisterMode = mode === "register";

  useEffect(() => {
    if (isAuthenticated && user) {
      navigate("/chat", { replace: true });
    }
  }, [isAuthenticated, navigate, user]);

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setErrorMessage(null);
    setIsSubmitting(true);

    try {
      if (mode === "register") {
        const fallbackPassword = paranoidMode
          ? `paranoid-${crypto.randomUUID().replaceAll("-", "")}`
          : password;
        const result = await register({ username, password: fallbackPassword });
        setRecoveryState(result);
      } else {
        await login({ handle, password });
        navigate("/chat");
      }
    } catch (error) {
      if (error instanceof ApiError) {
        setErrorMessage(error.message);
      } else {
        setErrorMessage("Não foi possível concluir a autenticação.");
      }
    } finally {
      setIsSubmitting(false);
    }
  }

  async function handleCopyRecoveryKey() {
    if (!recoveryState) {
      return;
    }

    try {
      setIsCopyingRecoveryKey(true);
      await navigator.clipboard.writeText(recoveryState.recoveryKey);
      setCopyFeedback("Chave copiada!");
      window.setTimeout(() => setIsCopyingRecoveryKey(false), 900);
    } catch {
      setCopyFeedback("Não foi possível copiar automaticamente.");
      setIsCopyingRecoveryKey(false);
    } finally {
      window.setTimeout(() => setCopyFeedback(null), 2000);
    }
  }

  function handleDownloadRecoveryPdf() {
    if (!recoveryState) {
      return;
    }

    const pdf = new jsPDF();
    pdf.setFont("helvetica", "bold");
    pdf.setFontSize(18);
    pdf.text("SlickChat - Chave de Recuperacao", 14, 20);

    pdf.setFont("helvetica", "normal");
    pdf.setFontSize(12);
    pdf.text(`Handle: ${recoveryState.handle}`, 14, 32);
    pdf.text("Guarde esta chave em local seguro.", 14, 40);

    const wrappedKey = pdf.splitTextToSize(recoveryState.recoveryKey, 180);
    pdf.setFont("courier", "normal");
    pdf.text("Recovery key:", 14, 52);
    pdf.text(wrappedKey, 14, 60);

    pdf.setFont("helvetica", "normal");
    pdf.text("Esta chave aparece uma vez e e obrigatoria para recuperar a conta.", 14, 85);
    pdf.save(`slickchat-recovery-${recoveryState.handle.replace("#", "_")}.pdf`);
  }

  if (recoveryState) {
    return (
      <main className="mx-auto flex min-h-dvh w-full max-w-xl items-center justify-center px-4 py-8">
        <section className="w-full max-w-[420px] rounded-2xl border border-[var(--primary-500)]/80 bg-[#0d0d10] p-6 shadow-[0_0_0_1px_rgba(122,0,255,0.25)] transition-all duration-300">
          <div className="mb-5 flex items-center justify-center">
            <div className="grid h-14 w-14 place-items-center rounded-full bg-[var(--primary-500)]/25 text-2xl text-[var(--primary-300)]">
              ✓
            </div>
          </div>
          <h1 className="text-center text-3xl leading-tight text-[var(--text-0)] md:text-4xl md:font-medium">
            Conta Criada
          </h1>
          <p className="mt-3 text-center text-base text-[var(--text-2)]">
            Seu nome de usuário é
          </p>
          <p className="mt-1 text-center text-xl text-[var(--primary-200)]">
            {recoveryState.handle}
          </p>

          <div className="mt-8 rounded-2xl border border-[var(--primary-500)] bg-[var(--bg-1)] p-4">
            <p className="mb-4 text-lg text-[#ffbf00]">Chave de Recuperação - Salve agora</p>
            <p className="mb-4 max-w-full break-all rounded-xl border border-white/10 bg-[var(--bg-0)] px-4 py-3 font-mono text-sm text-[var(--text-1)]">
              {recoveryState.recoveryKey}
            </p>
            <div className="mb-4 grid grid-cols-2 gap-2">
              <button
                type="button"
                onClick={handleCopyRecoveryKey}
                className="rounded-lg border border-white/20 bg-[#262830] px-3 py-2 text-sm font-medium text-[var(--text-1)] transition-all duration-200 hover:-translate-y-0.5 hover:border-[var(--primary-400)] hover:bg-[#2c2e37] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--primary-400)]"
              >
                {isCopyingRecoveryKey ? "Copiado" : "Copiar chave"}
              </button>
              <button
                type="button"
                onClick={handleDownloadRecoveryPdf}
                className="rounded-lg border border-white/20 bg-[#262830] px-3 py-2 text-sm font-medium text-[var(--text-1)] transition-all duration-200 hover:-translate-y-0.5 hover:border-[var(--primary-400)] hover:bg-[#2c2e37] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--primary-400)]"
              >
                Baixar PDF
              </button>
            </div>
            {copyFeedback ? (
              <p className="mb-3 text-xs text-[var(--success-500)]">{copyFeedback}</p>
            ) : null}
            <ul className="space-y-1 text-sm text-[var(--text-2)]">
              <li>Esta chave só aparece uma vez</li>
              <li>Sem ela, você perde sua conta para sempre</li>
              <li>Guarde em um lugar seguro</li>
            </ul>
          </div>

          <button
            type="button"
            className="mt-8 w-full rounded-xl bg-gradient-to-r from-[#7a00ff] to-[#8d2cff] py-3 text-base font-semibold text-white transition-all duration-200 hover:-translate-y-0.5 hover:brightness-110 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--primary-400)]"
            onClick={() => {
              setRecoveryState(null);
              setMode("login");
              setHandle(recoveryState.handle);
              setPassword("");
              setParanoidMode(false);
            }}
          >
            Já Salvei Minha Chave
          </button>
        </section>
      </main>
    );
  }

  return (
    <main className="mx-auto flex min-h-dvh w-full max-w-xl items-center justify-center px-4 py-8">
      <form
        onSubmit={onSubmit}
        className="w-full max-w-[470px] rounded-2xl border border-white/10 bg-[#090a0f] p-6 shadow-[0_10px_30px_rgba(0,0,0,0.45)] transition-all duration-300 md:p-8"
      >
        <h1 className="text-center text-4xl font-semibold tracking-wide text-[var(--text-0)] md:text-5xl">
          SlickChat
        </h1>
        <p className="mt-2 text-center text-xl text-[var(--text-2)]">Anônimo. Efêmero. Privado.</p>

        <div className="mt-8 grid grid-cols-2 rounded-xl border border-white/10 bg-[var(--bg-1)] p-1">
          <button
            type="button"
            onClick={() => setMode("register")}
            className={`rounded-lg px-3 py-2 text-sm font-medium transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--primary-400)] ${
              isRegisterMode
                ? "bg-gradient-to-r from-[#7a00ff] to-[#8d2cff] text-white shadow-[0_6px_18px_rgba(122,0,255,0.35)]"
                : "text-[var(--text-2)] hover:bg-white/10 hover:text-[var(--text-1)]"
            }`}
          >
            Criar conta
          </button>
          <button
            type="button"
            onClick={() => setMode("login")}
            className={`rounded-lg px-3 py-2 text-sm font-medium transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--primary-400)] ${
              !isRegisterMode
                ? "bg-gradient-to-r from-[#7a00ff] to-[#8d2cff] text-white shadow-[0_6px_18px_rgba(122,0,255,0.35)]"
                : "text-[var(--text-2)] hover:bg-white/10 hover:text-[var(--text-1)]"
            }`}
          >
            Entrar
          </button>
        </div>

        <div className="transition-all duration-200 ease-out">
          {isRegisterMode ? (
            <>
            <label className="mb-2 mt-6 block text-base text-[var(--text-1)]">Nome de usuário</label>
            <input
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="Digite seu nome"
              required
              minLength={3}
              className="w-full rounded-xl border border-white/15 bg-[#1a1b20] px-4 py-3 text-base text-[var(--text-0)] outline-none transition-all duration-200 focus:border-[var(--primary-400)] focus:ring-2 focus:ring-[var(--primary-400)]/30"
            />
            <p className="mt-2 text-sm text-[var(--text-3)]">Você receberá um ID único tipo #1234</p>
            <label className="mt-5 flex items-center gap-3 rounded-xl border border-white/15 bg-[#1a1b20] px-4 py-4 text-base text-[var(--text-1)]">
              <input
                type="checkbox"
                checked={paranoidMode}
                onChange={(e) => {
                  const checked = e.target.checked;
                  setParanoidMode(checked);
                  if (checked) {
                    setPassword("");
                  }
                }}
                className="h-4 w-4 rounded border border-white/30 bg-[var(--bg-0)]"
              />
              <span>Modo Paranoico (só chave de recuperação, sem senha)</span>
            </label>
            </>
          ) : (
            <>
            <label className="mb-2 mt-6 block text-base text-[var(--text-1)]">Handle</label>
            <input
              value={handle}
              onChange={(e) => setHandle(e.target.value)}
              placeholder="ex: alice#1234"
              required
              className="w-full rounded-xl border border-white/15 bg-[#1a1b20] px-4 py-3 text-base text-[var(--text-0)] outline-none transition-all duration-200 focus:border-[var(--primary-400)] focus:ring-2 focus:ring-[var(--primary-400)]/30"
            />
            </>
          )}
        </div>

        {!isRegisterMode || !paranoidMode ? (
          <>
            <label className="mb-2 mt-5 block text-base text-[var(--text-1)]">Senha</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Digite uma senha forte"
              required
              minLength={6}
              className="w-full rounded-xl border border-white/15 bg-[#1a1b20] px-4 py-3 text-base text-[var(--text-0)] outline-none transition-all duration-200 focus:border-[var(--primary-400)] focus:ring-2 focus:ring-[var(--primary-400)]/30"
            />
          </>
        ) : null}

        {errorMessage ? (
          <p className="mt-3 rounded-lg border border-[var(--danger-500)]/60 bg-[var(--danger-500)]/10 px-3 py-2 text-sm text-[var(--danger-500)]">
            {errorMessage}
          </p>
        ) : null}

        <button
          type="submit"
          disabled={isSubmitting}
          className="mt-6 w-full rounded-xl bg-gradient-to-r from-[#7a00ff] to-[#8d2cff] py-3 text-lg font-semibold text-white transition-all duration-200 hover:-translate-y-0.5 hover:brightness-110 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--primary-400)] disabled:cursor-not-allowed disabled:opacity-50"
        >
          {isSubmitting ? "Aguarde..." : formTitle}
        </button>

        <hr className="my-6 border-white/10" />
        <p className="mt-2 text-center text-lg text-[var(--primary-400)]">
          Criptografia ponta-a-ponta. Zero coleta de metadados.
        </p>
      </form>
    </main>
  );
}
