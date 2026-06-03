import { useEffect, useId, useState } from "react";
import { uploadRoomMedia } from "../../../features/media/upload-room-media";
import type { Room } from "../../../shared/api/types";
import { EditableMediaOverlay } from "../../../shared/ui/editable-media-overlay";
import { RoomAvatar, RoomBanner } from "./room-ui";

type RoomMediaEditorProps = {
  room: Room;
  token: string;
  canEdit: boolean;
  onMediaUpdated: (patch: {
    avatar_object_key?: string;
    banner_object_key?: string;
  }) => void;
};

export function RoomMediaEditor({
  room,
  token,
  canEdit,
  onMediaUpdated,
}: RoomMediaEditorProps) {
  const avatarInputId = useId();
  const bannerInputId = useId();

  const [avatarPreview, setAvatarPreview] = useState<string | null>(null);
  const [bannerPreview, setBannerPreview] = useState<string | null>(null);
  const [isUploadingAvatar, setIsUploadingAvatar] = useState(false);
  const [isUploadingBanner, setIsUploadingBanner] = useState(false);
  const [feedback, setFeedback] = useState<string | null>(null);

  useEffect(() => {
    return () => {
      if (avatarPreview?.startsWith("blob:")) {
        URL.revokeObjectURL(avatarPreview);
      }
      if (bannerPreview?.startsWith("blob:")) {
        URL.revokeObjectURL(bannerPreview);
      }
    };
  }, [avatarPreview, bannerPreview]);

  async function handleFile(purpose: "room_avatar" | "room_banner", file: File) {
    const blobUrl = URL.createObjectURL(file);
    if (purpose === "room_avatar") {
      setAvatarPreview(blobUrl);
      setIsUploadingAvatar(true);
    } else {
      setBannerPreview(blobUrl);
      setIsUploadingBanner(true);
    }
    setFeedback(null);

    try {
      const result = await uploadRoomMedia(token, room.room_id, purpose, file);
      if (purpose === "room_avatar" && result.object_key) {
        onMediaUpdated({ avatar_object_key: result.object_key });
        setAvatarPreview(null);
        URL.revokeObjectURL(blobUrl);
        setFeedback("Avatar atualizado.");
      }
      if (purpose === "room_banner" && result.object_key) {
        onMediaUpdated({ banner_object_key: result.object_key });
        setBannerPreview(null);
        URL.revokeObjectURL(blobUrl);
        setFeedback("Banner atualizado.");
      }
    } catch (error) {
      if (purpose === "room_avatar") {
        setAvatarPreview(null);
      } else {
        setBannerPreview(null);
      }
      URL.revokeObjectURL(blobUrl);
      setFeedback(
        error instanceof Error ? error.message.replaceAll("\n", " ").trim() : "Falha no upload.",
      );
    } finally {
      if (purpose === "room_avatar") {
        setIsUploadingAvatar(false);
      } else {
        setIsUploadingBanner(false);
      }
      window.setTimeout(() => setFeedback(null), 4000);
    }
  }

  const bannerBlock = (
    <RoomBanner
      name={room.name}
      roomId={room.room_id}
      previewUrl={bannerPreview ?? undefined}
      objectKey={room.banner_object_key}
      token={token}
    />
  );

  return (
    <div className="overflow-hidden rounded-2xl border border-white/10 bg-[var(--bg-1)]">
      <div className="relative">
        {canEdit ? (
          <EditableMediaOverlay
            inputId={bannerInputId}
            accept="image/jpeg,image/png,image/webp,image/gif"
            disabled={!canEdit}
            uploading={isUploadingBanner}
            ariaLabel="Alterar banner da sala"
            onFileSelected={(file) => void handleFile("room_banner", file)}
          >
            {bannerBlock}
          </EditableMediaOverlay>
        ) : (
          bannerBlock
        )}
      </div>

      <div className="relative flex flex-col items-center px-4 pb-5 pt-2">
        <div className="relative -mt-12">
          {canEdit ? (
            <EditableMediaOverlay
              inputId={avatarInputId}
              accept="image/jpeg,image/png,image/webp,image/gif"
              uploading={isUploadingAvatar}
              ariaLabel="Alterar avatar da sala"
              onFileSelected={(file) => void handleFile("room_avatar", file)}
              className={`rounded-2xl ring-4 ring-[var(--bg-1)] ${
                room.type === "PUBLIC" ? "ring-[var(--success-500)]/20" : ""
              }`}
            >
              <RoomAvatar
                name={room.name}
                roomId={room.room_id}
                previewUrl={avatarPreview ?? undefined}
                objectKey={room.avatar_object_key}
                token={token}
                size="xl"
                className={isUploadingAvatar ? "opacity-70" : ""}
              />
            </EditableMediaOverlay>
          ) : (
            <RoomAvatar
              name={room.name}
              roomId={room.room_id}
              previewUrl={avatarPreview ?? undefined}
              objectKey={room.avatar_object_key}
              token={token}
              size="xl"
              className={`ring-4 ring-[var(--bg-1)] ${
                room.type === "PUBLIC" ? "ring-[var(--success-500)]/20" : ""
              }`}
            />
          )}
        </div>

        {feedback ? (
          <p className="mt-3 text-center text-xs text-[var(--text-2)]">{feedback}</p>
        ) : null}
      </div>
    </div>
  );
}
