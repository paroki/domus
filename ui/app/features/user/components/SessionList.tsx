import { Laptop, Trash2 } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { ApiError } from "~/shared/api/client";
import { Badge } from "~/shared/ui/Badge";
import { Button } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import { useRevokeSession } from "../api/actions";
import { useMySessions } from "../api/queries";
import { formatUserAgent } from "../model";

export function SessionList() {
  const { t } = useTranslation();
  const { data, isLoading } = useMySessions();
  const revokeSession = useRevokeSession();
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const sessions = data?.data ?? [];

  async function handleRevoke(id: string) {
    setErrorMsg(null);
    try {
      await revokeSession.mutateAsync(id);
    } catch (err) {
      if (err instanceof ApiError) setErrorMsg(err.message);
    }
  }

  return (
    <Card className="p-4 sm:p-6">
      <h2 className="mb-4 text-lg font-bold text-fg">
        {t("me.sessionsTitle")}
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
      ) : sessions.length === 0 ? (
        <p className="text-sm text-muted">{t("me.sessionsEmpty")}</p>
      ) : (
        <ul className="divide-y divide-line">
          {sessions.map((sess) => (
            <li
              key={sess.id}
              className="flex flex-wrap items-center justify-between gap-3 py-3 first:pt-0 last:pb-0"
            >
              <div className="flex min-w-0 flex-1 items-start gap-3">
                <span className="mt-0.5 flex size-9 shrink-0 items-center justify-center rounded-lg bg-neutral-bg text-neutral-fg">
                  <Laptop aria-hidden className="size-4" />
                </span>
                <div className="min-w-0 flex-1">
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="truncate text-sm font-semibold text-fg">
                      {formatUserAgent(sess.userAgent ?? "")}
                    </span>
                    {sess.isCurrent && (
                      <Badge tone="success">{t("me.currentDevice")}</Badge>
                    )}
                  </div>
                  <p className="truncate text-xs text-muted">
                    {sess.ipAddress || "—"} · {t("me.lastActive")}:{" "}
                    {sess.lastActiveAt
                      ? new Date(sess.lastActiveAt).toLocaleString("id-ID")
                      : "—"}
                  </p>
                </div>
              </div>

              <Button
                variant="ghost"
                disabled={revokeSession.isPending}
                onClick={() => sess.id && handleRevoke(sess.id)}
                className="w-full sm:w-auto"
              >
                <Trash2 aria-hidden className="size-4" />
                {t("me.revokeSession")}
              </Button>
            </li>
          ))}
        </ul>
      )}
    </Card>
  );
}
