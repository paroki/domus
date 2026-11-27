import { Link2, MessageCircle, Send, Unlink } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router";
import { ApiError } from "~/shared/api/client";
import { Badge } from "~/shared/ui/Badge";
import { Button, buttonClass } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import { useUnlinkAccount } from "../api/actions";
import { useAuthProviders, useMyAccounts } from "../api/queries";
import { ME_ROUTES, providerLabel } from "../model";

export function AccountList() {
  const { t } = useTranslation();
  const { data, isLoading } = useMyAccounts();
  const { data: authConfig } = useAuthProviders();
  const unlinkAccount = useUnlinkAccount();
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const accounts = data?.data ?? [];
  const onlyOneLeft = accounts.length <= 1;
  const otpChannels = authConfig?.otpChannels ?? ["email"];
  const canLinkWhatsapp = otpChannels.includes("whatsapp");
  const canLinkTelegram = otpChannels.includes("telegram");
  const hasWhatsapp = accounts.some((a) => a.provider === "whatsapp");
  const hasTelegram = accounts.some((a) => a.provider === "telegram");

  async function handleUnlink(id: string) {
    setErrorMsg(null);
    try {
      await unlinkAccount.mutateAsync(id);
    } catch (err) {
      if (err instanceof ApiError) setErrorMsg(err.message);
    }
  }

  return (
    <Card className="p-4 sm:p-6">
      <div className="mb-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <h2 className="text-lg font-bold text-fg">{t("me.accountsTitle")}</h2>

        {(canLinkWhatsapp || canLinkTelegram) && (
          <div className="flex flex-col gap-2 sm:flex-row sm:flex-wrap sm:items-center">
            {canLinkWhatsapp && (
              <Link
                to={ME_ROUTES.linkWhatsapp}
                className={`${buttonClass("outline")} w-full sm:w-auto`}
              >
                <MessageCircle aria-hidden className="size-4" />
                {hasWhatsapp ? t("me.replaceWhatsapp") : t("me.linkWhatsapp")}
              </Link>
            )}
            {canLinkTelegram && (
              <Link
                to={ME_ROUTES.linkTelegram}
                className={`${buttonClass("outline")} w-full sm:w-auto`}
              >
                <Send aria-hidden className="size-4" />
                {hasTelegram ? t("me.replaceTelegram") : t("me.linkTelegram")}
              </Link>
            )}
          </div>
        )}
      </div>

      {errorMsg && (
        <div
          role="alert"
          className="mb-4 rounded-lg bg-danger-bg px-3.5 py-2.5 text-sm font-medium text-danger-fg"
        >
          {errorMsg}
        </div>
      )}

      {isLoading ? (
        <p className="text-sm text-muted">{t("auth.loading")}</p>
      ) : accounts.length === 0 ? (
        <p className="text-sm text-muted">{t("me.accountsEmpty")}</p>
      ) : (
        <>
          <ul className="divide-y divide-line">
            {accounts.map((acc) => (
              <li
                key={acc.id}
                className="flex flex-wrap items-center justify-between gap-3 py-3 first:pt-0 last:pb-0"
              >
                <div className="flex min-w-0 flex-1 items-center gap-3">
                  <span className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-neutral-bg text-neutral-fg">
                    <Link2 aria-hidden className="size-4" />
                  </span>
                  <div className="min-w-0 flex-1">
                    <div className="flex items-center gap-2">
                      <Badge tone="neutral">
                        {providerLabel(acc.provider ?? "")}
                      </Badge>
                    </div>
                    <p className="mt-1 truncate text-sm text-fg">
                      {acc.email || acc.displayName || acc.subject}
                    </p>
                  </div>
                </div>

                <Button
                  variant="ghost"
                  disabled={onlyOneLeft || unlinkAccount.isPending}
                  onClick={() => acc.id && handleUnlink(acc.id)}
                  className="w-full sm:w-auto"
                >
                  <Unlink aria-hidden className="size-4" />
                  {t("me.unlinkAccount")}
                </Button>
              </li>
            ))}
          </ul>
          {onlyOneLeft && (
            <p className="mt-3 text-xs text-muted">{t("me.lastAccountHint")}</p>
          )}
        </>
      )}
    </Card>
  );
}
