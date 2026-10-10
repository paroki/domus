import { GithubOutlined, GoogleOutlined } from "@ant-design/icons";
import { signIn } from "@domus/better-auth/client";
import { Alert, Button, Typography } from "antd";
import { useState } from "react";
import { Trans, useTranslation } from "react-i18next";
import { Link, useSearchParams } from "react-router";
import i18n from "~/i18n";
import { LanguageSwitcher } from "~/shared/layout/LanguageSwitcher";
import type { Route } from "../../routes/+types/login";

type Provider = "google" | "github";
const linkStyle = { color: "var(--domus-primary)", fontWeight: 500 } as const;

type ErrorCode = "access_denied" | "account_not_linked" | "oauth";

export function meta(_: Route.MetaArgs) {
  return [{ title: `${i18n.t("login.pageTitle")} | Domus` }];
}

// Cegah open redirect: hanya terima path internal.
function safeRedirect(value: string | null): string {
  if (!value?.startsWith("/") || value.startsWith("//")) return "/";
  return value;
}

const ERROR_CODES: readonly ErrorCode[] = [
  "access_denied",
  "account_not_linked",
  "oauth",
];

/** Kode error dari query string; kode yang tidak dikenal jatuh ke `oauth`. */
function toErrorCode(value: string | null): ErrorCode | null {
  if (!value) return null;
  return ERROR_CODES.find((code) => code === value) ?? "oauth";
}

export default function Login() {
  const { t } = useTranslation();
  const [params] = useSearchParams();
  const [pending, setPending] = useState<Provider | null>(null);
  // Simpan kode (bukan teks) supaya pesan ikut berganti saat bahasa diubah.
  const [error, setError] = useState<ErrorCode | null>(() =>
    toErrorCode(params.get("error")),
  );

  const redirectTo = safeRedirect(params.get("redirect"));

  async function doSignIn(provider: Provider) {
    setError(null);
    setPending(provider);
    try {
      const { error } = await signIn.social({
        provider,
        callbackURL: `${window.location.origin}${redirectTo}`,
        errorCallbackURL: `${window.location.origin}/login`,
      });
      if (error) {
        setError("oauth");
        setPending(null);
      }
      // Sukses: browser diarahkan ke penyedia, loading dibiarkan jalan.
    } catch {
      setError("oauth");
      setPending(null);
    }
  }

  return (
    <main className="min-h-screen grid place-items-center px-4 py-10">
      <div className="fixed top-3 right-3">
        <LanguageSwitcher />
      </div>
      <section
        aria-labelledby="login-title"
        className="glass glass-strong w-full max-w-100 p-8"
        style={{ borderRadius: "var(--domus-radius-lg)" }}
      >
        <div className="flex flex-col items-center text-center">
          <div
            aria-hidden="true"
            className="grid place-items-center size-11 mb-4 text-xl font-medium"
            style={{
              background: "var(--domus-brand)",
              color: "#fff",
              borderRadius: "var(--domus-radius)",
            }}
          >
            D
          </div>
          <Typography.Title
            id="login-title"
            level={2}
            style={{ margin: 0, fontWeight: 500 }}
          >
            {t("login.title")}
          </Typography.Title>
          <Typography.Paragraph
            style={{
              margin: "6px 0 0",
              color: "var(--domus-text-secondary)",
            }}
          >
            {t("login.subtitle")}
          </Typography.Paragraph>
        </div>

        {error && (
          <Alert
            className="mt-5"
            type="error"
            showIcon
            title={t(`login.errors.${error}`)}
            closable={{
              onClose: () => setError(null),
            }}
          />
        )}

        <div className="mt-6 flex flex-col gap-3">
          <Button
            size="large"
            block
            type="primary"
            icon={<GoogleOutlined />}
            loading={pending === "google"}
            disabled={pending !== null && pending !== "google"}
            onClick={() => doSignIn("google")}
          >
            {t("login.continueWithGoogle")}
          </Button>
          <Button
            size="large"
            block
            icon={<GithubOutlined />}
            loading={pending === "github"}
            disabled={pending !== null && pending !== "github"}
            onClick={() => doSignIn("github")}
          >
            {t("login.continueWithGithub")}
          </Button>
        </div>

        <Typography.Paragraph
          style={{
            margin: "20px 0 0",
            textAlign: "center",
            fontSize: 13,
            color: "var(--domus-text-muted)",
          }}
        >
          {t("login.noAccess")}
        </Typography.Paragraph>

        <p
          className="mt-5 pt-4 text-center"
          style={{
            fontSize: 13,
            lineHeight: 1.6,
            color: "var(--domus-text-secondary)",
            borderTop: "1px solid var(--domus-border-soft)",
          }}
        >
          <Trans
            i18nKey="login.agreement"
            components={{
              terms: <Link to="/terms" style={linkStyle} />,
              privacy: <Link to="/privacy" style={linkStyle} />,
            }}
          />
        </p>
      </section>
    </main>
  );
}
