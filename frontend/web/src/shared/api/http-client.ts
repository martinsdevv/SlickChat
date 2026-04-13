import { env } from "../config/env";
import { ApiError } from "./types";

type RequestOptions = {
  method?: "GET" | "POST" | "DELETE";
  token?: string | null;
  query?: Record<string, string | number | boolean | undefined>;
  body?: unknown;
};

function createUrl(path: string, query?: RequestOptions["query"]): string {
  const prefixedPath = path.startsWith("/") ? path : `/${path}`;
  const base = env.apiBaseUrl.startsWith("http")
    ? env.apiBaseUrl
    : `${window.location.origin}${env.apiBaseUrl}`;
  const url = new URL(`${base}${prefixedPath}`);

  if (!query) {
    return url.toString();
  }

  Object.entries(query).forEach(([key, value]) => {
    if (value === undefined) {
      return;
    }

    url.searchParams.set(key, String(value));
  });

  return url.toString();
}

export async function apiRequest<T>(
  path: string,
  { method = "GET", token, body, query }: RequestOptions = {},
): Promise<T> {
  const response = await fetch(createUrl(path, query), {
    method,
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: body ? JSON.stringify(body) : undefined,
  });

  const contentType = response.headers.get("content-type") ?? "";
  const payload = contentType.includes("application/json")
    ? await response.json()
    : await response.text();

  if (!response.ok) {
    throw new ApiError({
      message:
        typeof payload === "string" && payload.trim()
          ? payload
          : `Request failed with status ${response.status}`,
      status: response.status,
      details: payload,
    });
  }

  return payload as T;
}
