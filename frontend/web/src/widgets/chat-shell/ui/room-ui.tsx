import {
  useAuthMediaBlob,
  useRoomImageDisplay,
} from "../../../features/media/use-auth-media-blob";
import type { Room, RoomMember } from "../../../shared/api/types";
import { UserAvatarEditor } from "./user-avatar-editor";

/** Variações só na escala roxa do tema (definicoes_ui §18.3). */
const ROOM_GRADIENTS = [
  ["#7a00ff", "#a052ff"],
  ["#6a0dad", "#7a00ff"],
  ["#5c00cc", "#b388ff"],
  ["#4a0080", "#a052ff"],
] as const;

const SIZE_CLASSES = {
  sm: "h-9 w-9 text-xs",
  md: "h-11 w-11 text-sm",
  lg: "h-16 w-16 text-lg",
  xl: "h-24 w-24 text-2xl",
} as const;

export function hashString(value: string): number {
  let hash = 0;
  for (let i = 0; i < value.length; i += 1) {
    hash = (hash << 5) - hash + value.charCodeAt(i);
    hash |= 0;
  }
  return Math.abs(hash);
}

export function getRoomInitials(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  if (parts.length >= 2) {
    return `${parts[0][0] ?? ""}${parts[1][0] ?? ""}`.toUpperCase();
  }
  const trimmed = name.trim();
  return (trimmed.slice(0, 2) || "?").toUpperCase();
}

export function getHandleInitials(handle: string): string {
  const username = handle.split("#")[0] ?? handle;
  return getRoomInitials(username);
}

export function getRoomGradient(roomId: string): readonly [string, string] {
  return ROOM_GRADIENTS[hashString(roomId) % ROOM_GRADIENTS.length];
}

export function formatDateTime(iso: string): string {
  return new Date(iso).toLocaleString("pt-BR", {
    day: "2-digit",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function formatRoomType(type: Room["type"]): string {
  switch (type) {
    case "PUBLIC":
      return "Pública";
    case "PRIVATE":
      return "Privada";
    case "TEMPORARY":
      return "Temporária";
    case "DIRECT":
      return "Direta";
    default:
      return type;
  }
}

export function formatRole(role: RoomMember["role"]): string {
  switch (role) {
    case "ADMIN":
      return "Administrador";
    case "MODERATOR":
      return "Moderador";
    case "MEMBER":
      return "Membro";
    default:
      return role;
  }
}

type RoomAvatarProps = {
  name: string;
  roomId: string;
  /** Blob URL local (preview durante upload). */
  previewUrl?: string;
  objectKey?: string;
  token?: string | null;
  size?: keyof typeof SIZE_CLASSES;
  className?: string;
};

export function RoomAvatar({
  name,
  roomId,
  previewUrl,
  objectKey,
  token,
  size = "md",
  className = "",
}: RoomAvatarProps) {
  const [from, to] = getRoomGradient(roomId);
  const sizeClass = SIZE_CLASSES[size];
  const imageUrl = useRoomImageDisplay(previewUrl, objectKey, token);

  return (
    <div
      className={`relative shrink-0 overflow-hidden rounded-2xl ring-1 ring-white/10 ${sizeClass} ${className}`}
      style={
        imageUrl
          ? undefined
          : { background: `linear-gradient(135deg, ${from} 0%, ${to} 100%)` }
      }
      aria-hidden
    >
      {imageUrl ? (
        <img src={imageUrl} alt="" className="h-full w-full object-cover" />
      ) : (
        <>
          <span className="absolute inset-0 grid place-items-center font-semibold text-white/95">
            {getRoomInitials(name)}
          </span>
          <span className="absolute inset-0 bg-[url('data:image/svg+xml,%3Csvg xmlns=%22http://www.w3.org/2000/svg%22 width=%2240%22 height=%2240%22 viewBox=%220 0 40 40%22%3E%3Cpath fill=%22%23ffffff%22 fill-opacity=%220.04%22 d=%22M0 20h40M20 0v40%22/%3E%3C/svg%3E')] opacity-80" />
        </>
      )}
    </div>
  );
}

type RoomBannerProps = {
  name: string;
  roomId: string;
  previewUrl?: string;
  objectKey?: string;
  token?: string | null;
  className?: string;
};

export function RoomBanner({
  name,
  roomId,
  previewUrl,
  objectKey,
  token,
  className = "",
}: RoomBannerProps) {
  const [from, to] = getRoomGradient(roomId);
  const imageUrl = useRoomImageDisplay(previewUrl, objectKey, token);

  return (
    <div
      className={`relative h-28 w-full overflow-hidden bg-[var(--bg-2)] ${className}`}
      style={
        imageUrl
          ? undefined
          : { background: `linear-gradient(120deg, ${from}88 0%, ${to}44 45%, var(--bg-1) 100%)` }
      }
    >
      {imageUrl ? (
        <img src={imageUrl} alt="" className="h-full w-full object-cover" />
      ) : (
        <div className="absolute inset-0 bg-[url('data:image/svg+xml,%3Csvg xmlns=%22http://www.w3.org/2000/svg%22 width=%2260%22 height=%2260%22 viewBox=%220 0 60 60%22%3E%3Cpath fill=%22%23ffffff%22 fill-opacity=%220.03%22 d=%22M0 30h60M30 0v60%22/%3E%3C/svg%3E')] opacity-90" />
      )}
      <span className="sr-only">Banner da sala {name}</span>
    </div>
  );
}

type UserAvatarProps = {
  handle: string;
  userId?: string;
  objectKey?: string;
  token?: string | null;
  size?: keyof typeof SIZE_CLASSES;
};

export function UserAvatar({
  handle,
  userId,
  objectKey,
  token,
  size = "sm",
}: UserAvatarProps) {
  const seed = userId ?? handle;
  const [from, to] = getRoomGradient(seed);
  const sizeClass = SIZE_CLASSES[size];
  const { blobUrl } = useAuthMediaBlob(objectKey, token ?? null);

  return (
    <div
      className={`relative shrink-0 overflow-hidden rounded-full ring-1 ring-white/15 ${sizeClass}`}
      style={
        blobUrl
          ? undefined
          : {
              background: `linear-gradient(135deg, ${from} 0%, ${to} 55%, #1a1a1a 100%)`,
            }
      }
      aria-hidden
    >
      {blobUrl ? (
        <img src={blobUrl} alt="" className="h-full w-full object-cover" />
      ) : (
        <span className="absolute inset-0 grid place-items-center font-medium text-white/90">
          {getHandleInitials(handle)}
        </span>
      )}
    </div>
  );
}

type RoomTypeBadgeProps = {
  type: Room["type"];
  zeroLogging?: boolean;
  compact?: boolean;
};

function roomTypeBadgeClass(type: Room["type"]): string {
  switch (type) {
    case "PUBLIC":
      return "bg-[var(--success-500)]/15 text-[var(--success-500)]";
    case "PRIVATE":
      return "bg-cyan-500/12 text-cyan-300/90";
    case "TEMPORARY":
      return "bg-[var(--warning-500)]/15 text-[var(--warning-500)]";
    case "DIRECT":
      return "bg-[var(--bg-2)] text-[var(--text-2)]";
    default:
      return "bg-[var(--bg-2)] text-[var(--text-2)]";
  }
}

export function RoomTypeBadge({ type, zeroLogging, compact }: RoomTypeBadgeProps) {
  const typeClass = roomTypeBadgeClass(type);

  return (
    <div className={`flex flex-wrap items-center gap-1.5 ${compact ? "" : "mt-1"}`}>
      <span
        className={`rounded-md px-1.5 py-0.5 font-medium uppercase tracking-wide ${compact ? "text-[9px]" : "text-[10px]"} ${typeClass}`}
      >
        {type}
      </span>
      {zeroLogging ? (
        <span
          className={`rounded-md bg-[var(--primary-500)]/20 px-1.5 py-0.5 font-medium text-[var(--primary-200)] ${compact ? "text-[9px]" : "text-[10px]"}`}
        >
          ZERO
        </span>
      ) : null}
    </div>
  );
}

type UserSessionBarProps = {
  handle: string;
  userId?: string;
  avatarInputId: string;
  onLogout: () => void;
};

export function UserSessionBar({ handle, avatarInputId, onLogout }: UserSessionBarProps) {
  return (
    <div className="rounded-xl border border-white/10 bg-[var(--bg-1)] p-3">
      <div className="flex items-center gap-3">
        <UserAvatarEditor inputId={avatarInputId} />
        <div className="min-w-0 flex-1">
          <p className="text-[10px] font-medium uppercase tracking-wider text-[var(--text-3)]">
            Conectado como
          </p>
          <p className="truncate text-sm font-medium text-[var(--text-0)]">{handle}</p>
        </div>
        <button
          type="button"
          onClick={onLogout}
          title="Sair da sessão"
          className="shrink-0 rounded-lg border border-white/15 px-3 py-1.5 text-xs text-[var(--text-2)] transition hover:border-[var(--danger-500)]/40 hover:bg-[var(--danger-500)]/10 hover:text-[var(--danger-500)]"
        >
          Sair
        </button>
      </div>
    </div>
  );
}

type MetaRowProps = {
  label: string;
  value: string;
};

export function MetaRow({ label, value }: MetaRowProps) {
  return (
    <div className="flex items-start justify-between gap-3 py-2">
      <span className="text-xs text-[var(--text-3)]">{label}</span>
      <span className="text-right text-xs font-medium text-[var(--text-1)]">{value}</span>
    </div>
  );
}

type ParticipantRowProps = {
  member: RoomMember;
  isSelf?: boolean;
  token: string | null;
};

export function ParticipantRow({ member, isSelf, token }: ParticipantRowProps) {
  return (
    <li
      className={`flex items-center gap-3 rounded-xl px-3 py-2.5 ${
        isSelf ? "border border-[var(--primary-500)]/25 bg-[var(--primary-500)]/8" : "bg-[var(--bg-1)]"
      }`}
    >
      <UserAvatar
        handle={member.handle}
        userId={member.user_id}
        objectKey={member.avatar_object_key}
        token={token}
        size="sm"
      />
      <div className="min-w-0 flex-1">
        <p className="truncate text-sm text-[var(--text-0)]">
          {member.handle}
          {isSelf ? (
            <span className="ml-1.5 text-[10px] font-normal text-[var(--text-3)]">(você)</span>
          ) : null}
        </p>
        <p className="text-xs text-[var(--text-3)]">{formatRole(member.role)}</p>
      </div>
      <span
        className={`shrink-0 rounded-md px-1.5 py-0.5 text-[9px] font-medium uppercase tracking-wide ${
          member.role === "ADMIN"
            ? "bg-[var(--warning-500)]/12 text-[var(--warning-500)]"
            : member.role === "MODERATOR"
              ? "bg-cyan-500/12 text-cyan-300/80"
              : "bg-[var(--bg-2)] text-[var(--text-3)]"
        }`}
      >
        {member.role}
      </span>
    </li>
  );
}
