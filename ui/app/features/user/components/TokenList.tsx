import { Check, Copy, KeyRound, Plus, Trash2 } from "lucide-react";
import { type FormEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { ApiError } from "~/shared/api/client";
import { Badge } from "~/shared/ui/Badge";
import { Button } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import { useCreateToken, useRevokeToken } from "../api/actions";
import { useMyTokens } from "../api/queries";

export function TokenList() {
  const { t } = useTranslation();
  const { data, isLoading } = useMyTokens();
  const createToken = useCreateToken();
  const revokeToken = useRevokeToken();

  const [name, setName] = useState("");
  const [expiryDays, setExpiryDays] = useState("30");
  const [createdToken, setCreatedToken] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const tokens = data?.data ?? [];

  async function handleCreate(e: FormEvent) {
    e.preventDefault();
    setErrorMsg(null);
    setCreatedToken(null);
    setCopied(false);

    const parsedDays = Number(expiryDays);
    try {
      const created = await createToken.mutateAsync({
        name: name.trim(),
        expiresInDays: parsedDays > 0 ? parsedDays : undefined,
      });
      setCreatedToken(created.token ?? null);
      setName("");
    } catch (err) {
      if (err instanceof ApiError) setErrorMsg(err.message);
    }
  }

  async function handleRevoke(id: string) {
    setErrorMsg(null);
    try {
      await revokeToken.mutateAsync(id);
    } catch (err) {
      if (err instanceof ApiError) setErrorMsg(err.message);
    }
  }

  async function handleCopy() {
    if (!createdToken) return;
    await navigator.clipboard.writeText(createdToken);
    setCopied(true);
  }

  return (
    <Card className="p-6">
      <div className="mb-4">
        <h2 className="text-lg font-bold text-fg">{t("me.tokensTitle")}</h2>
        <p className="mt-1 text-sm text-muted">{t("me.tokensSubtitle")}</p>
      </div>

      {errorMsg && (
        <div
          role="alert"
          className="mb-4 rounded-lg bg-danger-bg px-3.5 py-2.5 text-sm font-medium text-danger-fg"
        >
          {errorMsg}
        </div>
      )}

      {createdToken && (
        <div
          role="status"
          data-testid="pat-created-banner"
          className="mb-4 rounded-lg border border-success-fg/30 bg-success-bg p-4"
        >
          <p className="mb-2 text-sm font-medium text-success-fg">
            {t("me.tokenCreatedHint")}
          </p>
          <div className="flex flex-wrap items-center gap-2">
            <code
              data-testid="pat-plaintext-token"
              className="flex-1 break-all rounded-md bg-canvas px-3 py-2 font-mono text-xs text-fg"
            >
              {createdToken}
            </code>
            <Button type="button" variant="outline" onClick={handleCopy}>
              {copied ? (
                <Check aria-hidden className="size-4" />
              ) : (
                <Copy aria-hidden className="size-4" />
              )}
              {copied ? t("me.tokenCopied") : t("me.copyToken")}
            </Button>
          </div>
        </div>
      )}

      <form
        onSubmit={handleCreate}
        className="mb-6 grid gap-3 sm:grid-cols-[1fr_180px_auto] sm:items-end"
      >
        <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
          {t("me.tokenNameLabel")}
          <input
            type="text"
            required
            maxLength={100}
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder={t("me.tokenNamePlaceholder")}
            className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
          />
        </label>

        <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
          {t("me.tokenExpiryLabel")}
          <select
            value={expiryDays}
            onChange={(e) => setExpiryDays(e.target.value)}
            className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
          >
            <option value="30">{t("me.tokenExpiry30")}</option>
            <option value="90">{t("me.tokenExpiry90")}</option>
            <option value="365">{t("me.tokenExpiry365")}</option>
            <option value="0">{t("me.tokenExpiryNever")}</option>
          </select>
        </label>

        <Button type="submit" disabled={createToken.isPending}>
          <Plus aria-hidden className="size-4" />
          {createToken.isPending ? t("me.creatingToken") : t("me.createToken")}
        </Button>
      </form>

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
              <div className="flex items-start gap-3">
                <span className="mt-0.5 flex size-9 items-center justify-center rounded-lg bg-neutral-bg text-neutral-fg">
                  <KeyRound aria-hidden className="size-4" />
                </span>
                <div>
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="text-sm font-semibold text-fg">
                      {tok.name}
                    </span>
                    <Badge tone="neutral">{tok.tokenPrefix}…</Badge>
                  </div>
                  <p className="text-xs text-muted">
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
