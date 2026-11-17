import type { HTMLAttributes } from "react";

export type BadgeTone = "success" | "warning" | "danger" | "neutral";

const tones: Record<BadgeTone, string> = {
  success: "bg-success-bg text-success-fg",
  warning: "bg-warning-bg text-warning-fg",
  danger: "bg-danger-bg text-danger-fg",
  neutral: "bg-neutral-bg text-neutral-fg",
};

export function Badge({
  tone = "neutral",
  className = "",
  ...props
}: HTMLAttributes<HTMLSpanElement> & { tone?: BadgeTone }) {
  return (
    <span
      className={`inline-flex items-center rounded-full px-3 py-1 text-[13px] font-semibold ${tones[tone]} ${className}`}
      {...props}
    />
  );
}
