import type { ReactNode } from "react";

type EditableMediaOverlayProps = {
  children: ReactNode;
  inputId: string;
  accept: string;
  disabled?: boolean;
  uploading?: boolean;
  ariaLabel: string;
  onFileSelected: (file: File) => void;
  className?: string;
};

export function EditableMediaOverlay({
  children,
  inputId,
  accept,
  disabled,
  uploading,
  ariaLabel,
  onFileSelected,
  className = "",
}: EditableMediaOverlayProps) {
  if (disabled) {
    return <div className={className}>{children}</div>;
  }

  return (
    <label
      htmlFor={inputId}
      className={`group relative block cursor-pointer ${className}`}
    >
      {children}
      <span
        className={`pointer-events-none absolute inset-0 flex items-center justify-center rounded-[inherit] bg-black/0 transition group-hover:bg-black/45 ${
          uploading ? "bg-black/50" : ""
        }`}
      >
        <span
          className={`grid h-9 w-9 place-items-center rounded-full border border-white/25 bg-black/55 text-white opacity-0 shadow-lg transition group-hover:opacity-100 ${
            uploading ? "opacity-100" : ""
          }`}
          aria-hidden
        >
          <svg viewBox="0 0 24 24" className="h-4 w-4" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M12 20h9M16.5 3.5a2.12 2.12 0 013 3L7 19l-4 1 1-4L16.5 3.5z" />
          </svg>
        </span>
        {uploading ? (
          <span className="absolute bottom-2 text-[10px] text-white">Enviando…</span>
        ) : null}
      </span>
      <input
        id={inputId}
        type="file"
        accept={accept}
        className="sr-only"
        disabled={uploading}
        aria-label={ariaLabel}
        onChange={(event) => {
          const file = event.target.files?.[0];
          if (file) {
            onFileSelected(file);
          }
          event.target.value = "";
        }}
      />
    </label>
  );
}
