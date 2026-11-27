import { Download, X } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import { IosInstallDialog } from "./IosInstallDialog";
import { usePwaInstall } from "./usePwaInstall";

export function PwaInstallBanner() {
  const { t } = useTranslation();
  const {
    isStandalone,
    canInstallNative,
    isIosInstallable,
    isDismissed,
    promptInstall,
    dismissBanner,
  } = usePwaInstall();
  const [iosDialogOpen, setIosDialogOpen] = useState(false);

  const visible =
    !isStandalone && (canInstallNative || isIosInstallable) && !isDismissed;

  if (!visible && !iosDialogOpen) {
    return null;
  }

  async function handleInstall() {
    if (canInstallNative) {
      await promptInstall();
    } else if (isIosInstallable) {
      setIosDialogOpen(true);
    }
  }

  return (
    <>
      {visible && (
        <div className="fixed inset-x-0 bottom-[max(1rem,env(safe-area-inset-bottom))] z-40 mx-auto max-w-lg px-4">
          <Card className="flex items-center justify-between gap-3 border-primary/20 p-3.5 shadow-lg">
            <div className="flex min-w-0 flex-1 items-center gap-3">
              <img
                src="/logo-192.png"
                alt=""
                aria-hidden
                className="size-10 shrink-0 rounded-xl"
              />
              <div className="min-w-0">
                <p className="text-sm font-semibold text-fg">
                  {t("pwa.bannerTitle")}
                </p>
                <p className="text-xs text-muted">{t("pwa.bannerSubtitle")}</p>
              </div>
            </div>

            <div className="flex shrink-0 items-center gap-1.5">
              <Button type="button" onClick={handleInstall}>
                <Download aria-hidden className="size-4" />
                {t("pwa.installAction")}
              </Button>
              <Button
                type="button"
                variant="ghost"
                aria-label={t("pwa.dismissLabel")}
                onClick={dismissBanner}
                className="size-11 p-0"
              >
                <X aria-hidden className="size-4" />
              </Button>
            </div>
          </Card>
        </div>
      )}

      <IosInstallDialog
        open={iosDialogOpen}
        onClose={() => setIosDialogOpen(false)}
      />
    </>
  );
}
