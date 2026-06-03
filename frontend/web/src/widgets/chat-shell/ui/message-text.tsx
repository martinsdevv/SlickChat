type MessageTextProps = {
  content: string;
};

export function MessageText({ content }: MessageTextProps) {
  return (
    <p className="max-w-full whitespace-pre-wrap break-words text-base leading-relaxed [overflow-wrap:anywhere]">
      {content}
    </p>
  );
}
