import { type PropsWithChildren, useEffect } from "react";

export function AppProviders({ children }: PropsWithChildren) {
  useEffect(() => {
    document.documentElement.classList.add("dark");
  }, []);

  return children;
}
