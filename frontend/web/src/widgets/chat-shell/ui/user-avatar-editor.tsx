import { useState } from "react";
import { useAuthMediaBlob } from "../../../features/media/use-auth-media-blob";
import { uploadUserAvatar } from "../../../features/media/upload-user-avatar";
import { useSessionStore } from "../../../features/session/store";
import { EditableMediaOverlay } from "../../../shared/ui/editable-media-overlay";
import { UserAvatar } from "./room-ui";

type UserAvatarEditorProps = {
  inputId: string;
};

export function UserAvatarEditor({ inputId }: UserAvatarEditorProps) {
  const { user, token, patchUserAvatar, refreshProfile } = useSessionStore();
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const [isUploading, setIsUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (!user || !token) {
    return null;
  }

  async function handleFile(file: File) {
    const blobUrl = URL.createObjectURL(file);
    setPreviewUrl(blobUrl);
    setIsUploading(true);
    setError(null);

    try {
      const result = await uploadUserAvatar(token!, file);
      patchUserAvatar(result.object_key);
      setPreviewUrl(null);
      URL.revokeObjectURL(blobUrl);
      void refreshProfile();
    } catch (uploadError) {
      setPreviewUrl(null);
      URL.revokeObjectURL(blobUrl);
      const raw =
        uploadError instanceof Error
          ? uploadError.message.replaceAll("\n", " ").trim()
          : "Não foi possível salvar a foto.";
      setError(
        raw === "invalid file size for avatar"
          ? "Arquivo vazio ou maior que 5 MB."
          : raw,
      );
    } finally {
      setIsUploading(false);
    }
  }

  return (
    <div className="flex flex-col items-center gap-1">
      <EditableMediaOverlay
        inputId={inputId}
        accept="image/jpeg,image/png,image/webp,image/gif"
        uploading={isUploading}
        ariaLabel="Alterar foto de perfil"
        onFileSelected={(file) => void handleFile(file)}
        className="shrink-0 rounded-full"
      >
        {previewUrl ? (
          <img
            src={previewUrl}
            alt=""
            className="h-11 w-11 rounded-full object-cover ring-1 ring-white/15"
          />
        ) : (
          <ProfileAvatarImage
            handle={user.handle}
            userId={user.userId}
            objectKey={user.avatarObjectKey}
            token={token}
          />
        )}
      </EditableMediaOverlay>
      {error ? (
        <p className="max-w-[140px] text-center text-[10px] text-[var(--danger-500)]">{error}</p>
      ) : null}
    </div>
  );
}

type ProfileAvatarImageProps = {
  handle: string;
  userId: string;
  objectKey?: string;
  token: string;
};

function ProfileAvatarImage({ handle, userId, objectKey, token }: ProfileAvatarImageProps) {
  const { blobUrl } = useAuthMediaBlob(objectKey, token);
  if (blobUrl) {
    return (
      <img
        src={blobUrl}
        alt=""
        className="h-11 w-11 rounded-full object-cover ring-1 ring-white/15"
      />
    );
  }
  return <UserAvatar handle={handle} userId={userId} size="md" />;
}
