import { Link2, Unlink } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { ApiError } from "~/shared/api/client";
import { Badge } from "~/shared/ui/Badge";
import { Button } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import { useUnlinkAccount } from "../api/actions";
import { useMyAccounts } from "../api/queries";
import { providerLabel } from "../model";

export function AccountList() {
  const { t } = useTranslation();
  const { data, isLoading } = useMyAccounts();
  const unlinkAccount = useUnlinkAccount();
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const accounts = data?.data ?? [];
  const onlyOneLeft = accounts.length <= 1;

  async function handleUnlink(id: string) {
    setErrorMsg(null);
    try {
      await unlinkAccount.mutateAsync(id);
    } catch (err) {
      if (err instanceof ApiError) setErrorMsg(err.message);
    }
  }

  return (
    <Card className="p-6">
      <h2 className="mb-4 text-lg font-bold text-fg">
        {t("me.accountsTitle")}
      </h2>

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
                <div className="flex items-center gap-3">
                  <span className="flex size-9 items-center justify-center rounded-lg bg-neutral-bg text-neutral-fg">
                    <Link2 aria-hidden className="size-4" />
                  </span>
                  <div>
                    <div className="flex items-center gap-2">
                      <Badge tone="neutral">
                        {providerLabel(acc.provider ?? "")}
                      </Badge>
                    </div>
                    <p className="mt-1 text-sm text-fg">
                      {acc.email || acc.subject}
                    </p>
                  </div>
                </div>

                <Button
                  variant="ghost"
                  disabled={onlyOneLeft || unlinkAccount.isPending}
                  onClick={() => acc.id && handleUnlink(acc.id)}
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
