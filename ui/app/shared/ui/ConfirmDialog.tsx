import { useEffect, useId } from "react";
import { Button, type ButtonVariant } from "./Button";
import { Card } from "./Card";

interface ConfirmDialogProps {
  open: boolean;
  title: string;
  description: string;
  confirmLabel?: string;
  cancelLabel?: string;
  confirmVariant?: ButtonVariant;
  pending?: boolean;
  onConfirm: () => void;
  onCancel: () => void;
}

export function ConfirmDialog({
  open,
  title,
  description,
  confirmLabel = "Hapus",
  cancelLabel = "Batal",
  confirmVariant = "danger",
  pending = false,
  onConfirm,
  onCancel,
}: ConfirmDialogProps) {
  const titleId = useId();
  const descId = useId();

  useEffect(() => {
    if (!open) return;
    function handleKeyDown(e: KeyboardEvent) {
      if (e.key === "Escape" && !pending) {
        onCancel();
      }
    }
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [open, pending, onCancel]);

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <Card
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
        aria-describedby={descId}
        className="w-full max-w-md p-6 shadow-xl"
      >
        <h2 id={titleId} className="text-lg font-bold text-fg">
          {title}
        </h2>
        <p id={descId} className="mt-2 text-sm text-muted">
          {description}
        </p>
        <div className="mt-6 flex justify-end gap-2">
          <Button variant="outline" disabled={pending} onClick={onCancel}>
            {cancelLabel}
          </Button>
          <Button
            variant={confirmVariant}
            disabled={pending}
            onClick={onConfirm}
          >
            {confirmLabel}
          </Button>
        </div>
      </Card>
    </div>
  );
}
