import { useEffect } from "react";
import { createPortal } from "react-dom";

type ImageLightboxProps = {
  src: string;
  alt: string;
  caption?: string;
  open: boolean;
  onClose: () => void;
};

export function ImageLightbox({ src, alt, caption, open, onClose }: ImageLightboxProps) {
  useEffect(() => {
    if (!open) {
      return;
    }

    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        onClose();
      }
    };

    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    window.addEventListener("keydown", onKeyDown);

    return () => {
      document.body.style.overflow = previousOverflow;
      window.removeEventListener("keydown", onKeyDown);
    };
  }, [open, onClose]);

  if (!open) {
    return null;
  }

  return createPortal(
    <div
      className="fixed inset-0 z-[100] flex items-center justify-center bg-black/85 p-4 backdrop-blur-sm"
      role="dialog"
      aria-modal="true"
      aria-label="Visualização da imagem"
      onClick={onClose}
    >
      <button
        type="button"
        className="absolute right-4 top-4 rounded-full border border-white/20 bg-black/50 p-2 text-white transition hover:bg-white/15"
        onClick={onClose}
        aria-label="Fechar"
      >
        <svg viewBox="0 0 24 24" className="h-6 w-6" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="M18 6L6 18M6 6l12 12" />
        </svg>
      </button>

      <div
        className="flex max-h-full max-w-full flex-col items-center gap-3"
        onClick={(event) => event.stopPropagation()}
      >
        <img
          src={src}
          alt={alt}
          className="max-h-[min(85vh,900px)] max-w-[min(95vw,1200px)] rounded-lg object-contain shadow-2xl"
        />
        {caption?.trim() ? (
          <p className="max-w-[min(95vw,1200px)] text-center text-sm text-white/90">{caption}</p>
        ) : null}
      </div>
    </div>,
    document.body,
  );
}
