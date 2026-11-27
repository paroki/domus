import { Share, X } from "lucide-react";
import { useEffect, useId } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";

interface IosInstallDialogProps {
  open: boolean;
  onClose: () => void;
}

export function IosInstallDialog({ open, onClose }: IosInstallDialogProps) {
  const { t } = useTranslation();
  const titleId = useId();
  const descId = useId();

  useEffect(() => {
    if (!open) return;
    function handleKeyDown(e: KeyboardEvent) {
      if (e.key === "Escape") {
        onClose();
      }
    }
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [open, onClose]);

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <Card
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
        aria-describedby={descId}
        className="w-full max-w-md p-4 shadow-xl sm:p-6"
      >
        <div className="flex items-start justify-between gap-3">
          <h2 id={titleId} className="text-lg font-bold text-fg">
            {t("pwa.iosDialogTitle")}
          </h2>
          <Button
            type="button"
            variant="ghost"
            aria-label={t("pwa.closeLabel")}
            onClick={onClose}
            className="size-11 p-0"
          >
            <X aria-hidden className="size-4" />
          </Button>
        </div>

        <p id={descId} className="mt-2 text-sm text-muted">
          {t("pwa.iosDialogSubtitle")}
        </p>

        <ol className="mt-4 flex flex-col gap-3 text-sm text-fg">
          <li className="flex items-start gap-2.5">
            <span className="flex size-6 shrink-0 items-center justify-center rounded-full bg-primary/10 text-xs font-bold text-primary">
              1
            </span>
            <span className="flex items-center gap-1.5">
              {t("pwa.iosStep1")}
              <Share aria-hidden className="inline size-4 text-primary" />
            </span>
          </li>
          <li className="flex items-start gap-2.5">
            <span className="flex size-6 shrink-0 items-center justify-center rounded-full bg-primary/10 text-xs font-bold text-primary">
              2
            </span>
            <span>{t("pwa.iosStep2")}</span>
          </li>
          <li className="flex items-start gap-2.5">
            <span className="flex size-6 shrink-0 items-center justify-center rounded-full bg-primary/10 text-xs font-bold text-primary">
              3
            </span>
            <span>{t("pwa.iosStep3")}</span>
          </li>
        </ol>

        <div className="mt-6 flex justify-end">
          <Button type="button" onClick={onClose} className="w-full sm:w-auto">
            {t("pwa.iosGotIt")}
          </Button>
        </div>
      </Card>
    </div>
  );
}
