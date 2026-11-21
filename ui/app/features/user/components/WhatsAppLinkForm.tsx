import { MessageCircle } from "lucide-react";
import { type FormEvent, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { ApiError } from "~/shared/api/client";
import { Button } from "~/shared/ui/Button";
import { useRequestLinkAccountOTP, useVerifyLinkAccount } from "../api/actions";
import type { UserAccount } from "../api/types";

interface WhatsAppLinkFormProps {
  initialPhone?: string;
  onLinked?: (acc: UserAccount) => void;
  submitLabel?: string;
}

export function WhatsAppLinkForm({
  initialPhone = "",
  onLinked,
  submitLabel,
}: WhatsAppLinkFormProps) {
  const { t } = useTranslation();
  const requestOTP = useRequestLinkAccountOTP();
  const verifyLink = useVerifyLinkAccount();

  const [step, setStep] = useState<"phone" | "code">("phone");
  const [phone, setPhone] = useState(initialPhone);
  const [code, setCode] = useState("");
  const [cooldown, setCooldown] = useState(0);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  useEffect(() => {
    if (cooldown <= 0) return;
    const timer = setTimeout(() => setCooldown((c) => c - 1), 1000);
    return () => clearTimeout(timer);
  }, [cooldown]);

  async function handleRequestOTP(e: FormEvent) {
    e.preventDefault();
    setErrorMsg(null);
    try {
      await requestOTP.mutateAsync({
        channel: "whatsapp",
        target: phone.trim(),
      });
      setStep("code");
      setCooldown(60);
    } catch (err) {
      setErrorMsg(err instanceof ApiError ? err.message : t("auth.sendingOtp"));
    }
  }

  async function handleVerify(e: FormEvent) {
    e.preventDefault();
    setErrorMsg(null);
    try {
      const account = await verifyLink.mutateAsync({
        channel: "whatsapp",
        target: phone.trim(),
        code: code.trim(),
      });
      onLinked?.(account);
    } catch (err) {
      if (err instanceof ApiError) {
        setErrorMsg(err.message);
      }
    }
  }

  return (
    <div className="flex flex-col gap-3">
      {errorMsg && (
        <div
          role="alert"
          className="rounded-lg bg-danger-bg px-3.5 py-2.5 text-sm font-medium text-danger-fg"
        >
          {errorMsg}
        </div>
      )}

      {step === "phone" ? (
        <form onSubmit={handleRequestOTP} className="flex flex-col gap-3">
          <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
            {t("me.whatsappNumberLabel")}
            <input
              type="tel"
              required
              autoComplete="tel"
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
              placeholder={t("me.whatsappNumberPlaceholder")}
              className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg placeholder:text-muted"
            />
          </label>
          <div>
            <Button type="submit" disabled={requestOTP.isPending}>
              <MessageCircle aria-hidden className="size-4" />
              {requestOTP.isPending
                ? t("auth.sendingOtp")
                : t("me.whatsappSendOtp")}
            </Button>
          </div>
        </form>
      ) : (
        <form onSubmit={handleVerify} className="flex flex-col gap-3">
          <p className="text-sm text-muted">
            {t("me.whatsappOtpSentTo", { target: phone.trim() })}
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
          <div>
            <Button type="submit" disabled={verifyLink.isPending}>
              {verifyLink.isPending
                ? t("auth.verifyingOtp")
                : (submitLabel ?? t("me.whatsappVerifyLink"))}
            </Button>
          </div>
          <div className="flex items-center justify-between text-sm">
            <button
              type="button"
              onClick={() => {
                setStep("phone");
                setCode("");
                setErrorMsg(null);
              }}
              className="font-medium text-primary hover:underline"
            >
              {t("me.whatsappChangeNumber")}
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
    </div>
  );
}
