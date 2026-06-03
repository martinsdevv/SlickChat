const ALLOWED_TYPES = new Set([
  "image/jpeg",
  "image/png",
  "image/webp",
  "image/gif",
]);

export const MEDIA_LIMITS = {
  userAvatar: 5 * 1024 * 1024,
  roomAvatar: 5 * 1024 * 1024,
  roomBanner: 8 * 1024 * 1024,
  messageImage: 15 * 1024 * 1024,
} as const;

function normalizeMime(type: string): string {
  return type === "image/jpg" ? "image/jpeg" : type;
}

export function resolveImageContentType(file: File): string | null {
  if (file.type) {
    const normalized = normalizeMime(file.type.toLowerCase());
    if (ALLOWED_TYPES.has(normalized)) {
      return normalized;
    }
  }

  const name = file.name.toLowerCase();
  if (name.endsWith(".jpg") || name.endsWith(".jpeg")) {
    return "image/jpeg";
  }
  if (name.endsWith(".png")) {
    return "image/png";
  }
  if (name.endsWith(".webp")) {
    return "image/webp";
  }
  if (name.endsWith(".gif")) {
    return "image/gif";
  }

  return null;
}

/** Alguns navegadores reportam file.size = 0 até ler o arquivo. */
export async function getReliableFileSize(file: File): Promise<number> {
  if (file.size > 0) {
    return file.size;
  }
  const buffer = await file.arrayBuffer();
  return buffer.byteLength;
}

export async function validateImageFile(
  file: File,
  maxBytes: number,
  maxLabel: string,
): Promise<string | null> {
  if (!resolveImageContentType(file)) {
    return "Use JPEG, PNG, WebP ou GIF.";
  }

  const size = await getReliableFileSize(file);
  if (size <= 0) {
    return "Não foi possível ler o arquivo.";
  }
  if (size > maxBytes) {
    return `Arquivo muito grande (máx. ${maxLabel}).`;
  }
  return null;
}
