import { KeyRound, Plus, Trash2 } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router";
import { ApiError } from "~/shared/api/client";
import { Badge } from "~/shared/ui/Badge";
import { Button, buttonClass } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import { useRevokeToken } from "../api/actions";
import { useMyTokens } from "../api/queries";
import { ME_ROUTES } from "../model";

export function TokenList() {
  const { t } = useTranslation();
  const { data, isLoading } = useMyTokens();
  const revokeToken = useRevokeToken();
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const tokens = data?.data ?? [];

  async function handleRevoke(id: string) {
    setErrorMsg(null);
    try {
      await revokeToken.mutateAsync(id);
    } catch (err) {
      if (err instanceof ApiError) setErrorMsg(err.message);
    }
  }

  return (
    <Card className="p-4 sm:p-6">
      <div className="mb-4 flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0 flex-1">
          <h2 className="text-lg font-bold text-fg">{t("me.tokensTitle")}</h2>
          <p className="mt-1 text-sm text-muted">{t("me.tokensSubtitle")}</p>
        </div>

        <Link
          to={ME_ROUTES.tokenNew}
          className={`${buttonClass("primary")} w-full shrink-0 sm:w-auto`}
        >
          <Plus aria-hidden className="size-4" />
          {t("me.createToken")}
        </Link>
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
      ) : tokens.length === 0 ? (
        <p className="text-sm text-muted">{t("me.tokensEmpty")}</p>
      ) : (
        <ul className="divide-y divide-line">
          {tokens.map((tok) => (
            <li
              key={tok.id}
              data-testid={`pat-row-${tok.name}`}
              className="flex flex-wrap items-center justify-between gap-3 py-3 first:pt-0 last:pb-0"
            >
              <div className="flex min-w-0 flex-1 items-start gap-3">
                <span className="mt-0.5 flex size-9 shrink-0 items-center justify-center rounded-lg bg-neutral-bg text-neutral-fg">
                  <KeyRound aria-hidden className="size-4" />
                </span>
                <div className="min-w-0 flex-1">
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="truncate text-sm font-semibold text-fg">
                      {tok.name}
                    </span>
                    <Badge tone="neutral">{tok.tokenPrefix}…</Badge>
                  </div>
                  <p className="truncate text-xs text-muted">
                    {t("me.tokenCreatedAt")}:{" "}
                    {tok.createdAt
                      ? new Date(tok.createdAt).toLocaleDateString("id-ID")
                      : "—"}{" "}
                    · {t("me.tokenExpiresAt")}:{" "}
                    {tok.expiresAt
                      ? new Date(tok.expiresAt).toLocaleDateString("id-ID")
                      : t("me.tokenExpiryNever")}{" "}
                    · {t("me.tokenLastUsedAt")}:{" "}
                    {tok.lastUsedAt
                      ? new Date(tok.lastUsedAt).toLocaleString("id-ID")
                      : t("me.tokenNeverUsed")}
                  </p>
                </div>
              </div>

              <Button
                variant="ghost"
                disabled={revokeToken.isPending}
                onClick={() => tok.id && handleRevoke(tok.id)}
                className="w-full sm:w-auto"
              >
                <Trash2 aria-hidden className="size-4" />
                {t("me.revokeToken")}
              </Button>
            </li>
          ))}
        </ul>
      )}
    </Card>
  );
}
