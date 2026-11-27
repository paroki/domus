import { Bell, BellOff, Download, Send, Smartphone } from "lucide-react";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { ApiError } from "~/shared/api/client";
import { IosInstallDialog } from "~/shared/pwa/IosInstallDialog";
import { usePushSubscription } from "~/shared/pwa/usePushSubscription";
import { usePwaInstall } from "~/shared/pwa/usePwaInstall";
import { Badge } from "~/shared/ui/Badge";
import { Button } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import {
  useSubscribePush,
  useTestPush,
  useUnsubscribePush,
  useUpdateMe,
} from "../api/actions";
import { useMe, useVapidPublicKey } from "../api/queries";

export function PwaDeviceCard() {
  const { t } = useTranslation();
  const { isStandalone, canInstallNative, isIosInstallable, promptInstall } =
    usePwaInstall();
  const {
    isSupported: pushSupported,
    permission,
    isSubscribed,
    loading: pushLoading,
    subscribeBrowser,
    unsubscribeBrowser,
  } = usePushSubscription();
  const { data: me } = useMe();
  const vapidQuery = useVapidPublicKey();
  const subscribePush = useSubscribePush();
  const unsubscribePush = useUnsubscribePush();
  const testPush = useTestPush();
  const updateMe = useUpdateMe();

  const [iosDialogOpen, setIosDialogOpen] = useState(false);
  const [statusMsg, setStatusMsg] = useState<string | null>(null);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [notifyPush, setNotifyPush] = useState(true);
  const [notifyWhatsapp, setNotifyWhatsapp] = useState(true);
  const [notifyTelegram, setNotifyTelegram] = useState(true);

  useEffect(() => {
    if (me) {
      setNotifyPush(me.notifyPush ?? true);
      setNotifyWhatsapp(me.notifyWhatsapp ?? true);
      setNotifyTelegram(me.notifyTelegram ?? true);
    }
  }, [me]);

  async function handleInstall() {
    if (canInstallNative) {
      await promptInstall();
    } else if (isIosInstallable) {
      setIosDialogOpen(true);
    }
  }

  async function handleTogglePushSubscription() {
    setErrorMsg(null);
    setStatusMsg(null);
    try {
      if (isSubscribed) {
        const endpoint = await unsubscribeBrowser();
        if (endpoint) {
          await unsubscribePush.mutateAsync({ endpoint });
        }
      } else {
        const vapidKey = vapidQuery.data?.publicKey;
        if (!vapidKey) {
          throw new Error("VAPID public key belum tersedia.");
        }
        const sub = await subscribeBrowser(vapidKey);
        await subscribePush.mutateAsync(sub);
      }
    } catch (err) {
      if (err instanceof ApiError) {
        setErrorMsg(err.details[0]?.message ?? err.message);
      } else if (err instanceof Error) {
        setErrorMsg(err.message);
      }
    }
  }

  async function handleTestPush() {
    setErrorMsg(null);
    setStatusMsg(null);
    try {
      await testPush.mutateAsync();
      setStatusMsg(t("pwa.testPushSuccess"));
    } catch (err) {
      if (err instanceof ApiError) {
        setErrorMsg(err.details[0]?.message ?? err.message);
      } else if (err instanceof Error) {
        setErrorMsg(err.message);
      }
    }
  }

  async function handlePrefToggle(
    field: "notifyPush" | "notifyWhatsapp" | "notifyTelegram",
    checked: boolean,
  ) {
    setErrorMsg(null);
    setStatusMsg(null);
    const prev =
      field === "notifyPush"
        ? notifyPush
        : field === "notifyWhatsapp"
          ? notifyWhatsapp
          : notifyTelegram;
    if (field === "notifyPush") setNotifyPush(checked);
    else if (field === "notifyWhatsapp") setNotifyWhatsapp(checked);
    else setNotifyTelegram(checked);

    try {
      await updateMe.mutateAsync({ [field]: checked });
    } catch (err) {
      if (field === "notifyPush") setNotifyPush(prev);
      else if (field === "notifyWhatsapp") setNotifyWhatsapp(prev);
      else setNotifyTelegram(prev);
      if (err instanceof ApiError) {
        setErrorMsg(err.details[0]?.message ?? err.message);
      }
    }
  }

  const pushBusy =
    pushLoading || subscribePush.isPending || unsubscribePush.isPending;

  return (
    <Card className="p-4 sm:p-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="flex min-w-0 flex-1 items-start gap-3">
          <span className="mt-0.5 flex size-9 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary">
            <Smartphone aria-hidden className="size-4" />
          </span>
          <div className="min-w-0 flex-1">
            <h2 className="text-lg font-bold text-fg">
              {t("pwa.deviceCardTitle")}
            </h2>
            <p className="mt-0.5 text-sm text-muted">
              {t("pwa.deviceCardDescription")}
            </p>
          </div>
        </div>

        <div className="flex flex-wrap items-center gap-2">
          {isStandalone && (
            <Badge tone="success">{t("pwa.installedBadge")}</Badge>
          )}
          {isSubscribed && (
            <Badge tone="success">{t("pwa.pushSubscribedBadge")}</Badge>
          )}
        </div>
      </div>

      {!isStandalone && (
        <div className="mt-4">
          {canInstallNative || isIosInstallable ? (
            <Button
              type="button"
              onClick={handleInstall}
              className="w-full sm:w-auto"
            >
              <Download aria-hidden className="size-4" />
              {t("pwa.installDeviceButton")}
            </Button>
          ) : (
            <p className="text-xs text-muted">{t("pwa.unsupportedHint")}</p>
          )}
        </div>
      )}

      <div className="mt-5 border-t border-line pt-5">
        <h3 className="text-sm font-semibold text-fg">{t("pwa.pushTitle")}</h3>

        {errorMsg && (
          <div
            role="alert"
            className="mt-3 rounded-lg bg-danger-bg px-3.5 py-2.5 text-sm font-medium text-danger-fg"
          >
            {errorMsg}
          </div>
        )}
        {statusMsg && (
          <div
            role="status"
            className="mt-3 rounded-lg bg-success-bg px-3.5 py-2.5 text-sm font-medium text-success-fg"
          >
            {statusMsg}
          </div>
        )}

        <div className="mt-3 flex flex-wrap items-center gap-2">
          {pushSupported ? (
            permission === "denied" ? (
              <p className="text-xs text-danger-fg">
                {t("pwa.pushBlockedHint")}
              </p>
            ) : (
              <Button
                type="button"
                variant={isSubscribed ? "outline" : "primary"}
                disabled={pushBusy}
                onClick={handleTogglePushSubscription}
                className="w-full sm:w-auto"
              >
                {isSubscribed ? (
                  <BellOff aria-hidden className="size-4" />
                ) : (
                  <Bell aria-hidden className="size-4" />
                )}
                {isSubscribed
                  ? t("pwa.pushDisableButton")
                  : t("pwa.pushEnableButton")}
              </Button>
            )
          ) : (
            <p className="text-xs text-muted">{t("pwa.pushUnsupportedHint")}</p>
          )}

          <Button
            type="button"
            variant="outline"
            disabled={testPush.isPending}
            onClick={handleTestPush}
            className="w-full sm:w-auto"
          >
            <Send aria-hidden className="size-4" />
            {t("pwa.testPushButton")}
          </Button>
        </div>

        <div className="mt-4 flex flex-col gap-1">
          <label className="flex min-h-11 cursor-pointer items-center gap-2.5 text-sm text-fg">
            <input
              type="checkbox"
              checked={notifyPush}
              disabled={updateMe.isPending}
              onChange={(e) =>
                void handlePrefToggle("notifyPush", e.target.checked)
              }
              className="size-4 shrink-0 rounded border-line"
            />
            {t("pwa.channelPushLabel")}
          </label>
          <label className="flex min-h-11 cursor-pointer items-center gap-2.5 text-sm text-fg">
            <input
              type="checkbox"
              checked={notifyWhatsapp}
              disabled={updateMe.isPending}
              onChange={(e) =>
                void handlePrefToggle("notifyWhatsapp", e.target.checked)
              }
              className="size-4 shrink-0 rounded border-line"
            />
            {t("pwa.channelWhatsappLabel")}
          </label>
          <label className="flex min-h-11 cursor-pointer items-center gap-2.5 text-sm text-fg">
            <input
              type="checkbox"
              checked={notifyTelegram}
              disabled={updateMe.isPending}
              onChange={(e) =>
                void handlePrefToggle("notifyTelegram", e.target.checked)
              }
              className="size-4 shrink-0 rounded border-line"
            />
            {t("pwa.channelTelegramLabel")}
          </label>
        </div>
      </div>

      <IosInstallDialog
        open={iosDialogOpen}
        onClose={() => setIosDialogOpen(false)}
      />
    </Card>
  );
}
