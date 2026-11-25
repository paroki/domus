import { AlertTriangle } from "lucide-react";
import { useEffect, useId } from "react";
import { Button } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import { getInitials } from "../model";

export { useDebouncedValue } from "~/shared/util/use-debounced-value";

export function UserAvatar({
  name,
  email,
  highlight = false,
}: {
  name?: string | null;
  email?: string | null;
  highlight?: boolean;
}) {
  return (
    <span
      aria-hidden
      className={`flex size-10 shrink-0 items-center justify-center rounded-full text-xs font-bold ${
        highlight ? "bg-primary/15 text-primary" : "bg-surface-2 text-fg"
      }`}
    >
      {getInitials(name, email)}
    </span>
  );
}

export function ConfirmDialog({
  open,
  title,
  description,
  confirmLabel,
  cancelLabel,
  tone = "danger",
  pending = false,
  onConfirm,
  onCancel,
}: {
  open: boolean;
  title: string;
  description: string;
  confirmLabel: string;
  cancelLabel: string;
  tone?: "danger" | "primary";
  pending?: boolean;
  onConfirm: () => void;
  onCancel: () => void;
}) {
  const titleId = useId();
  const descId = useId();

  useEffect(() => {
    if (!open) return;
    function onKeyDown(e: KeyboardEvent) {
      if (e.key === "Escape" && !pending) onCancel();
    }
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [open, pending, onCancel]);

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <button
        type="button"
        tabIndex={-1}
        aria-label={cancelLabel}
        disabled={pending}
        onClick={onCancel}
        className="fixed inset-0 cursor-default bg-black/50 backdrop-blur-[1px]"
      />
      <Card
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
        aria-describedby={descId}
        className="relative z-10 w-full max-w-md p-6 shadow-xl"
      >
        <div className="flex items-start gap-3.5">
          <span
            className={`flex size-10 shrink-0 items-center justify-center rounded-xl ${
              tone === "danger"
                ? "bg-danger-bg text-danger-fg"
                : "bg-warning-bg text-warning-fg"
            }`}
          >
            <AlertTriangle aria-hidden className="size-5" />
          </span>
          <div className="min-w-0 flex-1">
            <h2 id={titleId} className="text-base font-bold text-fg">
              {title}
            </h2>
            <p id={descId} className="mt-1.5 text-sm text-muted">
              {description}
            </p>
          </div>
        </div>
        <div className="mt-6 flex justify-end gap-2">
          <Button variant="outline" disabled={pending} onClick={onCancel}>
            {cancelLabel}
          </Button>
          <Button variant={tone} disabled={pending} onClick={onConfirm}>
            {confirmLabel}
          </Button>
        </div>
      </Card>
    </div>
  );
}
