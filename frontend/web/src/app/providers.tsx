import { type PropsWithChildren, useEffect } from "react";
import { useSessionStore } from "../features/session/store";

export function AppProviders({ children }: PropsWithChildren) {
  const bootstrap = useSessionStore((state) => state.bootstrap);

  useEffect(() => {
    document.documentElement.classList.add("dark");
    void bootstrap();
  }, [bootstrap]);

  return children;
}
