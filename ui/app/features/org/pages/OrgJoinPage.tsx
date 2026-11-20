import { AlertTriangle, Building2, CheckCircle2, Mail } from "lucide-react";
import { type FormEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link, useNavigate, useParams } from "react-router";
import { useMe } from "~/shared/api/session";
import i18n from "~/shared/i18n";
import { Badge } from "~/shared/ui/Badge";
import { Button, buttonClass } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import {
  useOrgRequestOTP,
  useOrgVerifyOTP,
  useRedeemOrgInvite,
} from "../api/actions";
import { resolveOrgError } from "../api/errors";
import { useOrgAuthProviders, useOrgInvitePreview } from "../api/queries";
import { inviteStatusMeta, kindLabel, providerLabel } from "../model";
import type { Route } from "./+types/OrgJoinPage";

export function meta({}: Route.MetaArgs) {
  return [{ title: `${i18n.t("org.joinTitle")} · Domus` }];
}

export default function OrgJoinPage() {
  const { t } = useTranslation();
  const { code = "" } = useParams();
  const navigate = useNavigate();

  const previewQuery = useOrgInvitePreview(code);
  const meQuery = useMe();
  const providersQuery = useOrgAuthProviders();

  const redeemInvite = useRedeemOrgInvite(code);
  const requestOTP = useOrgRequestOTP();
  const verifyOTP = useOrgVerifyOTP();

  const [step, setStep] = useState<"email" | "code">("email");
  const [email, setEmail] = useState("");
  const [name, setName] = useState("");
  const [otpCode, setOtpCode] = useState("");
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const preview = previewQuery.data;
  const me = meQuery.data;
  const providers = providersQuery.data ?? [];

  async function handleRedeemAuthenticated() {
    if (!preview) return;
    setErrorMsg(null);
    try {
      await redeemInvite.mutateAsync();
      navigate(`/org/${preview.orgId}`, { replace: true });
    } catch (err) {
      setErrorMsg(resolveOrgError(err));
    }
  }

  async function handleRequestOTP(e: FormEvent) {
    e.preventDefault();
    setErrorMsg(null);
    try {
      await requestOTP.mutateAsync({
        email: email.trim(),
        inviteCode: code,
      });
      setStep("code");
    } catch (err) {
      setErrorMsg(resolveOrgError(err));
    }
  }

  async function handleVerifyOTP(e: FormEvent) {
    e.preventDefault();
    if (!preview) return;
    setErrorMsg(null);
    try {
      await verifyOTP.mutateAsync({
        email: email.trim(),
        code: otpCode.trim(),
        name: name.trim() || undefined,
        inviteCode: code,
      });
      navigate(`/org/${preview.orgId}`, { replace: true });
    } catch (err) {
      setErrorMsg(resolveOrgError(err));
    }
  }

  if (previewQuery.isLoading || meQuery.isLoading) {
    return (
      <main className="flex min-h-dvh items-center justify-center bg-canvas px-4 py-10">
        <p className="text-sm text-muted">Memuat undangan organisasi…</p>
      </main>
    );
  }

  if (previewQuery.isError || !preview) {
    return (
      <main className="flex min-h-dvh items-center justify-center bg-canvas px-4 py-10">
        <Card className="w-full max-w-md p-6 text-center">
          <AlertTriangle
            aria-hidden
            className="mx-auto mb-3 size-8 text-danger-fg"
          />
          <h1 className="text-lg font-bold text-fg">
            {t("org.invalidInviteTitle")}
          </h1>
          <p className="mt-1 text-sm text-muted">
            {resolveOrgError(previewQuery.error)}
          </p>
          <Link to="/" className={`${buttonClass("outline")} mt-5`}>
            {t("org.backHome")}
          </Link>
        </Card>
      </main>
    );
  }

  const statusMeta = inviteStatusMeta(preview.status ?? "");

  return (
    <main className="flex min-h-dvh items-center justify-center bg-canvas px-4 py-10">
      <Card className="w-full max-w-md p-6 sm:p-8">
        <div className="mb-6 flex flex-col items-center gap-2 text-center">
          <span className="flex size-12 items-center justify-center rounded-xl bg-primary text-on-primary">
            <Building2 aria-hidden className="size-6" />
          </span>
          <h1 className="text-xl font-bold text-fg">{preview.orgName}</h1>
          <p className="text-sm text-muted">{preview.parishName}</p>
          <div className="mt-1 flex flex-wrap items-center justify-center gap-2">
            <Badge tone="neutral">{kindLabel(preview.orgKind ?? "")}</Badge>
            <Badge tone={statusMeta.tone}>{statusMeta.label}</Badge>
            <Badge tone="neutral">
              {preview.role === "admin"
                ? t("org.orgAdminBadge")
                : t("org.orgMemberBadge")}
            </Badge>
            {preview.defaultPosition && (
              <Badge tone="neutral">{preview.defaultPosition}</Badge>
            )}
          </div>
          <p className="mt-1 text-xs text-muted">
            Berlaku hingga{" "}
            {preview.expiresAt
              ? new Date(preview.expiresAt).toLocaleString("id-ID")
              : "-"}
          </p>
        </div>

        {errorMsg && (
          <div
            role="alert"
            className="mb-4 rounded-lg bg-danger-bg px-3.5 py-2.5 text-sm font-medium text-danger-fg"
          >
            {errorMsg}
          </div>
        )}

        {!preview.valid ? (
          <div className="flex flex-col items-center gap-4 text-center">
            <p className="text-sm text-muted">
              Tautan undangan ini berstatus <strong>{statusMeta.label}</strong>{" "}
              dan tidak dapat digunakan lagi.
            </p>
            <Link to="/" className={buttonClass("outline")}>
              {t("org.backHome")}
            </Link>
          </div>
        ) : me ? (
          <div className="flex flex-col gap-4">
            <div className="rounded-lg border border-line bg-canvas p-3.5 text-sm text-fg">
              Masuk sebagai <strong>{me.name}</strong> ({me.email})
            </div>
            <Button
              disabled={redeemInvite.isPending}
              onClick={handleRedeemAuthenticated}
            >
              <CheckCircle2 aria-hidden className="size-4" />
              {redeemInvite.isPending
                ? t("org.joining")
                : t("org.joinButton", { orgName: preview.orgName })}
            </Button>
          </div>
        ) : (
          <div className="flex flex-col gap-5">
            {providers.length > 0 && (
              <div className="flex flex-col gap-2.5">
                {providers.map((provider) => (
                  <a
                    key={provider}
                    href={`/api/auth/${encodeURIComponent(provider)}?redirect=${encodeURIComponent(`/org/${preview.orgId}`)}&inviteCode=${encodeURIComponent(code)}`}
                    className={buttonClass("outline")}
                  >
                    {t("auth.continueWith", {
                      provider: providerLabel(provider),
                    })}
                  </a>
                ))}
                <div className="relative my-1 flex items-center justify-center">
                  <span className="h-px w-full bg-line" />
                  <span className="bg-surface px-3 text-xs font-medium text-muted">
                    {t("auth.orEmail")}
                  </span>
                  <span className="h-px w-full bg-line" />
                </div>
              </div>
            )}

            {step === "email" ? (
              <form onSubmit={handleRequestOTP} className="flex flex-col gap-3">
                <label className="flex flex-col gap-1 text-sm font-medium text-fg">
                  Nama lengkap (untuk pengguna baru)
                  <input
                    type="text"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    placeholder="Nama lengkap Anda"
                    className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
                  />
                </label>
                <label className="flex flex-col gap-1 text-sm font-medium text-fg">
                  {t("auth.emailLabel")}
                  <input
                    type="email"
                    required
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    placeholder={t("auth.emailPlaceholder")}
                    className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
                  />
                </label>
                <Button type="submit" disabled={requestOTP.isPending}>
                  <Mail aria-hidden className="size-4" />
                  {requestOTP.isPending
                    ? t("auth.sendingOtp")
                    : t("auth.sendOtp")}
                </Button>
              </form>
            ) : (
              <form onSubmit={handleVerifyOTP} className="flex flex-col gap-3">
                <p className="text-sm text-muted">
                  {t("auth.otpSentTo", { email: email.trim() })}
                </p>
                <label className="flex flex-col gap-1 text-sm font-medium text-fg">
                  {t("auth.otpLabel")}
                  <input
                    type="text"
                    inputMode="numeric"
                    pattern="[0-9]{6}"
                    maxLength={6}
                    required
                    value={otpCode}
                    onChange={(e) => setOtpCode(e.target.value)}
                    placeholder={t("auth.otpPlaceholder")}
                    className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm tracking-widest text-fg"
                  />
                </label>
                <Button type="submit" disabled={verifyOTP.isPending}>
                  {verifyOTP.isPending
                    ? t("org.joining")
                    : t("org.joinButton", { orgName: preview.orgName })}
                </Button>
                <Button
                  variant="ghost"
                  onClick={() => {
                    setStep("email");
                    setOtpCode("");
                  }}
                >
                  {t("auth.changeEmail")}
                </Button>
              </form>
            )}
          </div>
        )}
      </Card>
    </main>
  );
}
