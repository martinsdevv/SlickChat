import { env } from "../../shared/config/env";
import { ApiError } from "../../shared/api/types";

function buildApiUrl(path: string): string {
  const base = env.apiBaseUrl.startsWith("http")
    ? env.apiBaseUrl
    : `${window.location.origin}${env.apiBaseUrl}`;
  const normalizedBase = base.endsWith("/") ? base : `${base}/`;
  const relative = path.startsWith("/") ? path.slice(1) : path;
  return new URL(relative, normalizedBase).toString();
}

function isLocalMinioUrl(uploadUrl: string): boolean {
  try {
    const host = new URL(uploadUrl).hostname;
    return host === "localhost" || host === "127.0.0.1";
  } catch {
    return false;
  }
}

function resolveUploadTarget(
  uploadUrl: string,
  objectKey: string,
  uploadViaApi?: boolean,
): { targetUrl: string; useApiProxy: boolean } {
  if (uploadViaApi || uploadUrl.startsWith("/")) {
    return { targetUrl: buildApiUrl(uploadUrl), useApiProxy: true };
  }

  if (isLocalMinioUrl(uploadUrl)) {
    if (!objectKey) {
      throw new ApiError({
        message:
          "Upload indisponível remotamente. Reinicie a API (make demo-restart-api) com deploy/.env carregado.",
        status: 503,
      });
    }
    const proxyPath = `/media/upload-put?object_key=${encodeURIComponent(objectKey)}`;
    return { targetUrl: buildApiUrl(proxyPath), useApiProxy: true };
  }

  return { targetUrl: uploadUrl, useApiProxy: false };
}

/** PUT do arquivo — via API (túnel / remoto) ou presigned MinIO (somente dev local direto). */
export async function putUploadedFile(
  uploadUrl: string,
  token: string,
  file: File,
  contentType: string,
  objectKey: string,
  uploadViaApi?: boolean,
): Promise<void> {
  const { targetUrl, useApiProxy } = resolveUploadTarget(
    uploadUrl,
    objectKey,
    uploadViaApi,
  );

  const response = await fetch(targetUrl, {
    method: "PUT",
    headers: {
      "Content-Type": contentType,
      ...(useApiProxy && token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: file,
  });

  if (!response.ok) {
    const detail = await response.text();
    throw new ApiError({
      message: detail.trim() || `Upload falhou (${response.status})`,
      status: response.status,
    });
  }
}
