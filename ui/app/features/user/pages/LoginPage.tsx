import { Church, Mail, MessageCircle, Send } from "lucide-react";
import { type FormEvent, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate, useSearchParams } from "react-router";
import { ApiError } from "~/shared/api/client";
import i18n from "~/shared/i18n";
import { Button, buttonClass } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import { useRequestOTP, useVerifyOTP } from "../api/actions";
import { useAuthProviders } from "../api/queries";
import { providerLabel, sanitizeRedirect } from "../model";
import type { Route } from "./+types/LoginPage";

export function meta({}: Route.MetaArgs) {
  return [{ title: `${i18n.t("auth.title")} · Domus` }];
}

type OtpChannel = "email" | "whatsapp" | "telegram";

function normalizeChannel(raw: string): OtpChannel {
  if (raw === "whatsapp" || raw === "telegram") return raw;
  return "email";
}

export default function LoginPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [params] = useSearchParams();
  const redirectTo = sanitizeRedirect(params.get("redirect"));

  const { data: authConfig } = useAuthProviders();
  const providers = authConfig?.providers ?? [];
  const otpChannels = authConfig?.otpChannels ?? ["email"];

  const requestOTP = useRequestOTP();
  const verifyOTP = useVerifyOTP();

  const [channel, setChannel] = useState<OtpChannel>("email");
  const [step, setStep] = useState<"target" | "code">("target");
  const [target, setTarget] = useState("");
  const [code, setCode] = useState("");
  const [cooldown, setCooldown] = useState(0);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  useEffect(() => {
    if (otpChannels.length > 0 && !otpChannels.includes(channel)) {
      setChannel(normalizeChannel(otpChannels[0] ?? "email"));
    }
  }, [channel, otpChannels]);

  useEffect(() => {
    if (cooldown <= 0) return;
    const timer = setTimeout(() => setCooldown((c) => c - 1), 1000);
    return () => clearTimeout(timer);
  }, [cooldown]);

  function handleSelectChannel(next: OtpChannel) {
    if (next === channel) return;
    setChannel(next);
    setStep("target");
    setTarget("");
    setCode("");
    setCooldown(0);
    setErrorMsg(null);
  }

  async function handleRequestOTP(e: FormEvent) {
    e.preventDefault();
    setErrorMsg(null);
    try {
      await requestOTP.mutateAsync({
        channel,
        target: target.trim(),
      });
      setStep("code");
      setCooldown(60);
    } catch (err) {
      setErrorMsg(err instanceof ApiError ? err.message : t("auth.sendingOtp"));
    }
  }

  async function handleVerifyOTP(e: FormEvent) {
    e.preventDefault();
    setErrorMsg(null);
    try {
      await verifyOTP.mutateAsync({
        channel,
        target: target.trim(),
        code: code.trim(),
      });
      navigate(redirectTo, { replace: true });
    } catch (err) {
      if (err instanceof ApiError) {
        setErrorMsg(err.message);
      }
    }
  }

  const channelMeta: Record<
    OtpChannel,
    {
      tabLabel: string;
      inputLabel: string;
      inputType: "email" | "tel" | "text";
      autoComplete: string;
      placeholder: string;
      Icon: typeof Mail;
    }
  > = {
    email: {
      tabLabel: t("auth.channelEmail"),
      inputLabel: t("auth.emailLabel"),
      inputType: "email",
      autoComplete: "email",
      placeholder: t("auth.emailPlaceholder"),
      Icon: Mail,
    },
    whatsapp: {
      tabLabel: t("auth.channelWhatsapp"),
      inputLabel: t("auth.whatsappLabel"),
      inputType: "tel",
      autoComplete: "tel",
      placeholder: t("auth.whatsappPlaceholder"),
      Icon: MessageCircle,
    },
    telegram: {
      tabLabel: t("auth.channelTelegram"),
      inputLabel: t("auth.telegramLabel"),
      inputType: "text",
      autoComplete: "username",
      placeholder: t("auth.telegramPlaceholder"),
      Icon: Send,
    },
  };

  const activeMeta = channelMeta[channel];
  const ChannelIcon = activeMeta.Icon;

  return (
    <main className="flex min-h-dvh items-center justify-center bg-canvas px-4 py-10">
      <Card className="w-full max-w-md p-6 sm:p-8">
        <div className="mb-6 flex flex-col items-center gap-2 text-center">
          <span className="flex size-12 items-center justify-center rounded-xl bg-primary text-on-primary">
            <Church aria-hidden className="size-6" />
          </span>
          <h1 className="text-2xl font-bold text-fg">{t("auth.title")}</h1>
          <p className="text-sm text-muted">{t("auth.subtitle")}</p>
        </div>

        {providers.length > 0 && (
          <div className="mb-6 flex flex-col gap-2.5">
            {providers.map((provider) => (
              <a
                key={provider}
                href={`/api/auth/${encodeURIComponent(provider)}?redirect=${encodeURIComponent(redirectTo)}`}
                className={buttonClass("outline")}
              >
                {t("auth.continueWith", {
                  provider: providerLabel(provider),
                })}
              </a>
            ))}

            <div className="relative my-2 flex items-center justify-center">
              <span className="h-px w-full bg-line" />
              <span className="bg-surface px-3 text-xs font-medium text-muted">
                {otpChannels.length > 1 ? t("auth.orOtp") : t("auth.orEmail")}
              </span>
              <span className="h-px w-full bg-line" />
            </div>
          </div>
        )}

        {otpChannels.length > 1 && (
          <div
            role="tablist"
            aria-label="OTP Channel"
            className="mb-5 grid grid-cols-3 gap-1 rounded-lg bg-canvas p-1 border border-line"
          >
            {otpChannels.map((raw) => {
              const ch = normalizeChannel(raw);
              const meta = channelMeta[ch];
              const isSelected = ch === channel;
              return (
                <button
                  key={ch}
                  type="button"
                  role="tab"
                  aria-selected={isSelected}
                  onClick={() => handleSelectChannel(ch)}
                  className={`rounded-md px-3 py-1.5 text-xs font-semibold transition-colors ${
                    isSelected
                      ? "bg-primary text-on-primary"
                      : "text-muted hover:text-fg"
                  }`}
                >
                  {meta.tabLabel}
                </button>
              );
            })}
          </div>
        )}

        {errorMsg && (
          <div
            role="alert"
            className="mb-4 rounded-lg bg-danger-bg px-3.5 py-2.5 text-sm font-medium text-danger-fg"
          >
            {errorMsg}
          </div>
        )}

        {step === "target" ? (
          <form onSubmit={handleRequestOTP} className="flex flex-col gap-4">
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              {activeMeta.inputLabel}
              <input
                type={activeMeta.inputType}
                required
                autoComplete={activeMeta.autoComplete}
                value={target}
                onChange={(e) => setTarget(e.target.value)}
                placeholder={activeMeta.placeholder}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg placeholder:text-muted"
              />
            </label>
            <Button type="submit" disabled={requestOTP.isPending}>
              <ChannelIcon aria-hidden className="size-4" />
              {requestOTP.isPending ? t("auth.sendingOtp") : t("auth.sendOtp")}
            </Button>
          </form>
        ) : (
          <form onSubmit={handleVerifyOTP} className="flex flex-col gap-4">
            <p className="text-sm text-muted">
              {t("auth.otpSentTo", { target: target.trim() })}
            </p>
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              {t("auth.otpLabel")}
              <input
                type="text"
                inputMode="numeric"
                pattern="[0-9]{6}"
                maxLength={6}
                required
                autoComplete="one-time-code"
                value={code}
                onChange={(e) => setCode(e.target.value)}
                placeholder={t("auth.otpPlaceholder")}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-center font-mono text-lg tracking-widest text-fg placeholder:text-muted"
              />
            </label>
            <Button type="submit" disabled={verifyOTP.isPending}>
              {verifyOTP.isPending
                ? t("auth.verifyingOtp")
                : t("auth.verifyOtp")}
            </Button>

            <div className="flex items-center justify-between text-sm">
              <button
                type="button"
                onClick={() => {
                  setStep("target");
                  setCode("");
                  setErrorMsg(null);
                }}
                className="font-medium text-primary hover:underline"
              >
                {t("auth.changeEmail")}
              </button>
              <button
                type="button"
                disabled={cooldown > 0 || requestOTP.isPending}
                onClick={handleRequestOTP}
                className="font-medium text-primary hover:underline disabled:text-muted disabled:no-underline"
              >
                {cooldown > 0
                  ? t("auth.resendIn", { seconds: cooldown })
                  : t("auth.resendOtp")}
              </button>
            </div>
          </form>
        )}
      </Card>
    </main>
  );
}
