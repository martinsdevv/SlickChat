import { useEffect, useState } from "react";
import { env } from "../../shared/config/env";

function mediaObjectRequestUrl(objectKey: string): string {
  const base = env.apiBaseUrl.startsWith("http")
    ? env.apiBaseUrl
    : `${window.location.origin}${env.apiBaseUrl}`;
  const url = new URL(`${base}/media/object`);
  url.searchParams.set("object_key", objectKey);
  return url.toString();
}

export function useAuthMediaBlob(objectKey: string | undefined, token: string | null) {
  const [blobUrl, setBlobUrl] = useState<string | null>(null);
  const [failed, setFailed] = useState(false);

  useEffect(() => {
    if (!objectKey || !token) {
      setBlobUrl(null);
      setFailed(false);
      return;
    }

    let cancelled = false;
    let objectUrl: string | null = null;

    void (async () => {
      try {
        const response = await fetch(mediaObjectRequestUrl(objectKey), {
          headers: { Authorization: `Bearer ${token}` },
        });
        if (!response.ok) {
          if (!cancelled) {
            setFailed(true);
            setBlobUrl(null);
          }
          return;
        }
        const blob = await response.blob();
        if (cancelled) {
          return;
        }
        objectUrl = URL.createObjectURL(blob);
        setBlobUrl(objectUrl);
        setFailed(false);
      } catch {
        if (!cancelled) {
          setFailed(true);
          setBlobUrl(null);
        }
      }
    })();

    return () => {
      cancelled = true;
      if (objectUrl) {
        URL.revokeObjectURL(objectUrl);
      }
      setBlobUrl((previous) => {
        if (previous && previous !== objectUrl) {
          URL.revokeObjectURL(previous);
        }
        return null;
      });
    };
  }, [objectKey, token]);

  return { blobUrl, failed };
}

function useRoomImageDisplay(
  previewUrl: string | undefined,
  objectKey: string | undefined,
  token: string | null | undefined,
): string | undefined {
  const { blobUrl } = useAuthMediaBlob(objectKey, token ?? null);
  if (previewUrl?.startsWith("blob:")) {
    return previewUrl;
  }
  return blobUrl ?? undefined;
}

export { useRoomImageDisplay };
