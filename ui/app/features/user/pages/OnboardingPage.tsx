import { ArrowLeft, CheckCircle2, Church, LogOut } from "lucide-react";
import {
  type FormEvent,
  useCallback,
  useEffect,
  useRef,
  useState,
} from "react";
import { useTranslation } from "react-i18next";
import { Navigate, useNavigate, useSearchParams } from "react-router";
import { ApiError } from "~/shared/api/client";
import i18n from "~/shared/i18n";
import { Badge } from "~/shared/ui/Badge";
import { Button } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import { useLogout, useUpdateMe } from "../api/actions";
import { useAuthProviders, useMe } from "../api/queries";
import { TelegramLinkPanel } from "../components/TelegramLinkPanel";
import { WhatsAppLinkForm } from "../components/WhatsAppLinkForm";
import { buildOnboardingSteps, sanitizeRedirect } from "../model";
import type { Route } from "./+types/OnboardingPage";

export function meta({}: Route.MetaArgs) {
  return [{ title: `${i18n.t("onboarding.title")} · Domus` }];
}

export default function OnboardingPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [params] = useSearchParams();
  const nextPath = sanitizeRedirect(params.get("next"));

  const { data: me } = useMe();
  const providersQuery = useAuthProviders();
  const updateMe = useUpdateMe();
  const logout = useLogout();

  const [stepIndex, setStepIndex] = useState(0);
  const [name, setName] = useState("");
  const [phone, setPhone] = useState("");
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const initializedRef = useRef(false);
  const completingRef = useRef(false);

  useEffect(() => {
    if (me && !initializedRef.current) {
      initializedRef.current = true;
      setName(me.name ?? "");
      setPhone(me.phone ?? "");
    }
  }, [me]);

  const otpChannels = providersQuery.data?.otpChannels ?? ["email"];
  const steps = buildOnboardingSteps(otpChannels);
  const safeIndex = Math.min(stepIndex, steps.length - 1);
  const currentStep = steps[safeIndex] ?? "biodata";
  const isFinalStep = safeIndex >= steps.length - 1;

  const completeOnboarding = useCallback(async () => {
    if (completingRef.current) return;
    completingRef.current = true;
    setErrorMsg(null);
    try {
      await updateMe.mutateAsync({ onboarded: true });
      navigate(nextPath, { replace: true });
    } catch (err) {
      completingRef.current = false;
      if (err instanceof ApiError) {
        setErrorMsg(err.details[0]?.message ?? err.message);
      }
    }
  }, [navigate, nextPath, updateMe]);

  const telegramLinked = Boolean(me?.linkedProviders?.includes("telegram"));
  const whatsappLinked = Boolean(me?.linkedProviders?.includes("whatsapp"));

  useEffect(() => {
    if (currentStep === "telegram" && telegramLinked) {
      void completeOnboarding();
    }
  }, [completeOnboarding, currentStep, telegramLinked]);

  if (me?.onboarded === true) {
    return <Navigate to={nextPath} replace />;
  }

  async function handleLogout() {
    await logout.mutateAsync();
    navigate("/login", { replace: true });
  }

  async function handleBiodataSubmit(e: FormEvent) {
    e.preventDefault();
    setErrorMsg(null);
    try {
      if (steps.length === 1) {
        await updateMe.mutateAsync({
          name: name.trim(),
          phone: phone.trim(),
          onboarded: true,
        });
        navigate(nextPath, { replace: true });
      } else {
        await updateMe.mutateAsync({
          name: name.trim(),
          phone: phone.trim(),
        });
        setStepIndex((i) => i + 1);
      }
    } catch (err) {
      if (err instanceof ApiError) {
        setErrorMsg(err.details[0]?.message ?? err.message);
      }
    }
  }

  function advanceOrFinish() {
    if (isFinalStep) {
      void completeOnboarding();
    } else {
      setErrorMsg(null);
      setStepIndex((i) => i + 1);
    }
  }

  function handlePrevious() {
    setErrorMsg(null);
    setStepIndex((i) => Math.max(0, i - 1));
  }

  return (
    <div className="flex min-h-dvh flex-col bg-canvas">
      <header className="border-b border-line bg-surface">
        <div className="mx-auto flex h-16 max-w-2xl items-center justify-between px-4 sm:px-6">
          <div className="flex items-center gap-2.5">
            <span className="flex size-9 items-center justify-center rounded-lg bg-primary text-on-primary">
              <Church aria-hidden className="size-5" />
            </span>
            <span className="text-lg font-bold tracking-tight text-fg">
              Domus
            </span>
          </div>

          <Button
            type="button"
            variant="ghost"
            disabled={logout.isPending}
            onClick={handleLogout}
          >
            <LogOut aria-hidden className="size-4" />
            {t("onboarding.logout")}
          </Button>
        </div>
      </header>

      <main className="flex flex-1 items-center justify-center px-4 py-10">
        <Card className="w-full max-w-lg p-6 sm:p-8">
          {providersQuery.isLoading ? (
            <div
              role="status"
              aria-live="polite"
              className="py-12 text-center text-sm font-medium text-muted"
            >
              {t("onboarding.loading")}
            </div>
          ) : (
            <div className="flex flex-col gap-6">
              <div className="flex flex-col gap-2">
                <div className="flex items-center justify-between gap-2">
                  <span className="text-xs font-semibold uppercase tracking-wide text-primary">
                    {t("onboarding.stepIndicator", {
                      current: safeIndex + 1,
                      total: steps.length,
                    })}
                  </span>
                  <div className="flex items-center gap-1.5" aria-hidden>
                    {steps.map((stepKey, idx) => (
                      <span
                        key={stepKey}
                        className={`h-1.5 w-6 rounded-full transition-colors ${
                          idx <= safeIndex ? "bg-primary" : "bg-line"
                        }`}
                      />
                    ))}
                  </div>
                </div>
                <h1 className="text-2xl font-bold text-fg">
                  {t("onboarding.title")}
                </h1>
                <p className="text-sm text-muted">{t("onboarding.subtitle")}</p>
              </div>

              {errorMsg && (
                <div
                  role="alert"
                  className="rounded-lg bg-danger-bg px-3.5 py-2.5 text-sm font-medium text-danger-fg"
                >
                  {errorMsg}
                </div>
              )}

              {currentStep === "biodata" && (
                <form
                  onSubmit={handleBiodataSubmit}
                  className="flex flex-col gap-4"
                >
                  <div>
                    <h2 className="text-lg font-bold text-fg">
                      {t("onboarding.biodataTitle")}
                    </h2>
                    <p className="text-sm text-muted">
                      {t("onboarding.biodataDescription")}
                    </p>
                  </div>

                  <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                    {t("onboarding.nameLabel")}
                    <input
                      type="text"
                      required
                      value={name}
                      onChange={(e) => setName(e.target.value)}
                      placeholder={t("onboarding.namePlaceholder")}
                      className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg placeholder:text-muted"
                    />
                  </label>

                  <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                    {t("onboarding.phoneLabel")}
                    <input
                      type="tel"
                      value={phone}
                      onChange={(e) => setPhone(e.target.value)}
                      placeholder={t("onboarding.phonePlaceholder")}
                      className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg placeholder:text-muted"
                    />
                  </label>

                  <div className="flex justify-end pt-2">
                    <Button type="submit" disabled={updateMe.isPending}>
                      {updateMe.isPending
                        ? t("onboarding.saving")
                        : steps.length === 1
                          ? t("onboarding.saveAndFinish")
                          : t("onboarding.saveAndContinue")}
                    </Button>
                  </div>
                </form>
              )}

              {currentStep === "whatsapp" && (
                <div className="flex flex-col gap-4">
                  <div>
                    <h2 className="text-lg font-bold text-fg">
                      {t("onboarding.whatsappTitle")}
                    </h2>
                    <p className="text-sm text-muted">
                      {t("onboarding.whatsappDescription")}
                    </p>
                  </div>

                  {whatsappLinked ? (
                    <div className="flex flex-col gap-2 rounded-xl border border-line bg-success-bg/50 p-4">
                      <div className="flex items-center gap-2">
                        <CheckCircle2
                          aria-hidden
                          className="size-5 text-success-fg"
                        />
                        <Badge tone="success">
                          {t("onboarding.whatsappLinkedTitle")}
                        </Badge>
                      </div>
                      {me?.phone && (
                        <p className="text-sm text-fg">
                          {t("onboarding.whatsappLinkedNumber", {
                            phone: me.phone,
                          })}
                        </p>
                      )}
                    </div>
                  ) : (
                    <WhatsAppLinkForm
                      initialPhone={phone || me?.phone || ""}
                      onLinked={(acc) => {
                        if (acc.subject) setPhone(acc.subject);
                        advanceOrFinish();
                      }}
                    />
                  )}

                  <div className="flex items-center justify-between gap-3 pt-2">
                    <Button
                      type="button"
                      variant="outline"
                      onClick={handlePrevious}
                    >
                      <ArrowLeft aria-hidden className="size-4" />
                      {t("onboarding.previous")}
                    </Button>

                    {whatsappLinked ? (
                      <Button
                        type="button"
                        disabled={updateMe.isPending}
                        onClick={advanceOrFinish}
                      >
                        {isFinalStep
                          ? t("onboarding.finish")
                          : t("onboarding.next")}
                      </Button>
                    ) : (
                      <Button
                        type="button"
                        variant="ghost"
                        disabled={updateMe.isPending}
                        onClick={advanceOrFinish}
                      >
                        {t("onboarding.skip")}
                      </Button>
                    )}
                  </div>
                </div>
              )}

              {currentStep === "telegram" && (
                <div className="flex flex-col gap-4">
                  <div>
                    <h2 className="text-lg font-bold text-fg">
                      {t("onboarding.telegramTitle")}
                    </h2>
                    <p className="text-sm text-muted">
                      {t("onboarding.telegramDescription")}
                    </p>
                  </div>

                  {telegramLinked ? (
                    <div
                      role="status"
                      className="rounded-xl border border-line bg-success-bg px-4 py-3 text-sm font-medium text-success-fg"
                    >
                      <p>{t("onboarding.telegramLinkedSuccess")}</p>
                      <p className="mt-1 text-xs opacity-90">
                        {t("onboarding.telegramFinishing")}
                      </p>
                    </div>
                  ) : (
                    <TelegramLinkPanel
                      autoStart
                      onLinked={() => {
                        void completeOnboarding();
                      }}
                    />
                  )}

                  <div className="flex items-center justify-between gap-3 pt-2">
                    <Button
                      type="button"
                      variant="outline"
                      onClick={handlePrevious}
                    >
                      <ArrowLeft aria-hidden className="size-4" />
                      {t("onboarding.previous")}
                    </Button>

                    <Button
                      type="button"
                      variant={telegramLinked ? "primary" : "ghost"}
                      disabled={updateMe.isPending}
                      onClick={() => {
                        void completeOnboarding();
                      }}
                    >
                      {telegramLinked
                        ? t("onboarding.finish")
                        : t("onboarding.skip")}
                    </Button>
                  </div>
                </div>
              )}
            </div>
          )}
        </Card>
      </main>
    </div>
  );
}
