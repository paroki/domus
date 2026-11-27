import { Download, Smartphone } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { IosInstallDialog } from "~/shared/pwa/IosInstallDialog";
import { usePwaInstall } from "~/shared/pwa/usePwaInstall";
import { Badge } from "~/shared/ui/Badge";
import { Button } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";

export function PwaDeviceCard() {
  const { t } = useTranslation();
  const { isStandalone, canInstallNative, isIosInstallable, promptInstall } =
    usePwaInstall();
  const [iosDialogOpen, setIosDialogOpen] = useState(false);

  async function handleInstall() {
    if (canInstallNative) {
      await promptInstall();
    } else if (isIosInstallable) {
      setIosDialogOpen(true);
    }
  }

  return (
    <Card className="p-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="flex items-start gap-3">
          <span className="mt-0.5 flex size-9 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary">
            <Smartphone aria-hidden className="size-4" />
          </span>
          <div>
            <h2 className="text-lg font-bold text-fg">
              {t("pwa.deviceCardTitle")}
            </h2>
            <p className="mt-0.5 text-sm text-muted">
              {t("pwa.deviceCardDescription")}
            </p>
          </div>
        </div>

        {isStandalone && (
          <Badge tone="success">{t("pwa.installedBadge")}</Badge>
        )}
      </div>

      {!isStandalone && (
        <div className="mt-4">
          {canInstallNative || isIosInstallable ? (
            <Button type="button" onClick={handleInstall}>
              <Download aria-hidden className="size-4" />
              {t("pwa.installDeviceButton")}
            </Button>
          ) : (
            <p className="text-xs text-muted">{t("pwa.unsupportedHint")}</p>
          )}
        </div>
      )}

      <IosInstallDialog
        open={iosDialogOpen}
        onClose={() => setIosDialogOpen(false)}
      />
    </Card>
  );
}
