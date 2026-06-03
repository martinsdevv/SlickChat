import { useState } from "react";
import { useAuthMediaBlob } from "../../../features/media/use-auth-media-blob";
import { ImageLightbox } from "./image-lightbox";

type MessageImageContentProps = {
  objectKey: string;
  token: string | null;
  previewUrl?: string;
  caption?: string;
};

export function MessageImageContent({
  objectKey,
  token,
  previewUrl,
  caption,
}: MessageImageContentProps) {
  const [expanded, setExpanded] = useState(false);
  const { blobUrl, failed } = useAuthMediaBlob(objectKey, token);
  const localPreview = previewUrl?.startsWith("blob:") ? previewUrl : undefined;
  const src = blobUrl ?? localPreview;

  if (!src) {
    return (
      <div className="flex h-40 w-56 max-w-full items-center justify-center rounded-lg bg-black/20 text-xs text-white/70">
        {failed ? "Imagem indisponível" : "Carregando…"}
      </div>
    );
  }

  return (
    <>
      <button
        type="button"
        className="block max-w-full cursor-zoom-in rounded-lg text-left transition hover:brightness-110 focus:outline-none focus-visible:ring-2 focus-visible:ring-white/40"
        onClick={(event) => {
          event.stopPropagation();
          setExpanded(true);
        }}
        aria-label="Expandir imagem"
      >
        <img
          src={src}
          alt={caption?.trim() || "Imagem enviada"}
          className="max-h-64 max-w-full rounded-lg object-contain"
        />
      </button>

      <ImageLightbox
        src={src}
        alt={caption?.trim() || "Imagem enviada"}
        caption={caption}
        open={expanded}
        onClose={() => setExpanded(false)}
      />
    </>
  );
}
