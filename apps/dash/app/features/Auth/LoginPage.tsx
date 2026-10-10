import { GithubOutlined, GoogleOutlined } from "@ant-design/icons";
import { signIn } from "@domus/better-auth/client";
import { Alert, Button, Typography } from "antd";
import { useState } from "react";
import { Link, useSearchParams } from "react-router";
import type { Route } from "../../routes/+types/login";

type Provider = "google" | "github";

export function meta(_: Route.MetaArgs) {
  return [{ title: "Masuk | Domus" }];
}

// Cegah open redirect: hanya terima path internal.
function safeRedirect(value: string | null): string {
  if (!value?.startsWith("/") || value.startsWith("//")) return "/";
  return value;
}

const ERROR_MESSAGES: Record<string, string> = {
  access_denied: "Izin login dibatalkan. Coba lagi kalau mau lanjut.",
  account_not_linked:
    "Email ini sudah terdaftar lewat penyedia lain. Masuk dengan penyedia yang dulu dipakai.",
  oauth: "Login gagal. Coba lagi sebentar lagi.",
};

export default function Login() {
  const [params] = useSearchParams();
  const [pending, setPending] = useState<Provider | null>(null);
  const [error, setError] = useState<string | null>(() => {
    const code = params.get("error");
    return code ? (ERROR_MESSAGES[code] ?? ERROR_MESSAGES.oauth) : null;
  });

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
        setError(ERROR_MESSAGES.oauth);
        setPending(null);
      }
      // Sukses: browser diarahkan ke penyedia, loading dibiarkan jalan.
    } catch {
      setError(ERROR_MESSAGES.oauth);
      setPending(null);
    }
  }

  return (
    <main className="min-h-screen grid place-items-center px-4 py-10">
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
            Masuk ke Domus
          </Typography.Title>
          <Typography.Paragraph
            style={{
              margin: "6px 0 0",
              color: "var(--domus-text-secondary)",
            }}
          >
            Kelola keuskupan, paroki, dan lingkungan di satu tempat.
          </Typography.Paragraph>
        </div>

        {error && (
          <Alert
            className="mt-5"
            type="error"
            showIcon
            title={error}
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
            Lanjutkan dengan Google
          </Button>
          <Button
            size="large"
            block
            icon={<GithubOutlined />}
            loading={pending === "github"}
            disabled={pending !== null && pending !== "github"}
            onClick={() => doSignIn("github")}
          >
            Lanjutkan dengan GitHub
          </Button>
        </div>

        <Typography.Paragraph
          style={{
            margin: "20px 0 0",
            textAlign: "center",
            fontSize: 12,
            color: "var(--domus-text-muted)",
          }}
        >
          Belum punya akses? Hubungi admin keuskupan atau paroki kamu.
        </Typography.Paragraph>

        <p
          className="mt-5 pt-4 text-center"
          style={{
            fontSize: 12,
            lineHeight: 1.6,
            color: "var(--domus-text-secondary)",
            borderTop: "1px solid var(--domus-border-soft)",
          }}
        >
          Dengan masuk, kamu menyetujui{" "}
          <Link
            to="/terms"
            style={{ color: "var(--domus-primary)", fontWeight: 500 }}
          >
            Ketentuan Layanan
          </Link>{" "}
          dan{" "}
          <Link
            to="/privacy"
            style={{ color: "var(--domus-primary)", fontWeight: 500 }}
          >
            Kebijakan Privasi
          </Link>
          .
        </p>
      </section>
    </main>
  );
}
