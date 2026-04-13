const DEFAULT_API_BASE_URL = "/api";
const DEFAULT_WS_BASE_URL = "/ws";

function normalizeBaseUrl(value: string | undefined, fallback: string): string {
  if (!value || !value.trim()) {
    return fallback;
  }

  return value.replace(/\/+$/, "");
}

export const env = {
  apiBaseUrl: normalizeBaseUrl(
    import.meta.env.VITE_API_BASE_URL,
    DEFAULT_API_BASE_URL,
  ),
  wsBaseUrl: normalizeBaseUrl(
    import.meta.env.VITE_WS_BASE_URL,
    DEFAULT_WS_BASE_URL,
  ),
} as const;
