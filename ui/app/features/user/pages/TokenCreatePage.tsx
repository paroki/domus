import { ArrowLeft, Check, Copy, Plus } from "lucide-react";
import { type FormEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router";
import { ApiError } from "~/shared/api/client";
import i18n from "~/shared/i18n";
import { Button, buttonClass } from "~/shared/ui/Button";
import { InductiveFormSection, InductivePageHeader } from "~/shared/ui/iui";
import { useCreateToken } from "../api/actions";
import { ME_ROUTES } from "../model";
import type { Route } from "./+types/TokenCreatePage";

export function meta({}: Route.MetaArgs) {
  return [{ title: `${i18n.t("me.createTokenPageTitle")} · Domus` }];
}

export default function TokenCreatePage() {
  const { t } = useTranslation();
  const createToken = useCreateToken();

  const [name, setName] = useState("");
  const [expiryDays, setExpiryDays] = useState("30");
  const [createdToken, setCreatedToken] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

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

  async function handleCopy() {
    if (!createdToken) return;
    await navigator.clipboard.writeText(createdToken);
    setCopied(true);
  }

  return (
    <div className="mx-auto flex w-full max-w-3xl flex-col gap-6 px-4 py-6 sm:px-6 sm:py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: t("me.homeBreadcrumb"), to: "/" },
          { label: t("me.title"), to: ME_ROUTES.index },
          { label: t("me.createTokenPageTitle") },
        ]}
        title={t("me.createTokenPageTitle")}
        subtitle={t("me.createTokenPageSubtitle")}
      />

      <InductiveFormSection
        title={t("me.createTokenSectionTitle")}
        description={t("me.tokensSubtitle")}
      >
        {errorMsg && (
          <div
            role="alert"
            className="rounded-lg bg-danger-bg px-3.5 py-2.5 text-sm font-medium text-danger-fg"
          >
            {errorMsg}
          </div>
        )}

        {createdToken ? (
          <div className="flex flex-col gap-4">
            <div
              role="status"
              data-testid="pat-created-banner"
              className="rounded-lg border border-success-fg/30 bg-success-bg p-4"
            >
              <p className="mb-2 text-sm font-medium text-success-fg">
                {t("me.tokenCreatedHint")}
              </p>
              <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
                <code
                  data-testid="pat-plaintext-token"
                  className="min-w-0 flex-1 break-all rounded-md bg-canvas px-3 py-2 font-mono text-xs text-fg"
                >
                  {createdToken}
                </code>
                <Button
                  type="button"
                  variant="outline"
                  onClick={handleCopy}
                  className="w-full shrink-0 sm:w-auto"
                >
                  {copied ? (
                    <Check aria-hidden className="size-4" />
                  ) : (
                    <Copy aria-hidden className="size-4" />
                  )}
                  {copied ? t("me.tokenCopied") : t("me.copyToken")}
                </Button>
              </div>
            </div>

            <div>
              <Link
                to={ME_ROUTES.index}
                className={`${buttonClass("primary")} w-full sm:w-auto`}
              >
                <ArrowLeft aria-hidden className="size-4" />
                {t("me.backToProfile")}
              </Link>
            </div>
          </div>
        ) : (
          <form onSubmit={handleCreate} className="flex flex-col gap-4">
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

            <div className="flex flex-col-reverse gap-2 pt-2 sm:flex-row sm:items-center sm:justify-end">
              <Link
                to={ME_ROUTES.index}
                className={`${buttonClass("outline")} w-full sm:w-auto`}
              >
                {t("me.cancelEdit")}
              </Link>
              <Button
                type="submit"
                disabled={createToken.isPending}
                className="w-full sm:w-auto"
              >
                <Plus aria-hidden className="size-4" />
                {createToken.isPending
                  ? t("me.creatingToken")
                  : t("me.createToken")}
              </Button>
            </div>
          </form>
        )}
      </InductiveFormSection>
    </div>
  );
}
