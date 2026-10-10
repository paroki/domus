import { getSession } from "@domus/better-auth/client";
import { redirect } from "react-router";

/**
 * Guard untuk route yang butuh login. Dipanggil dari `clientLoader`.
 * Kalau belum ada sesi, lempar redirect ke `/login?redirect=<path asal>`
 * supaya user kembali ke halaman tujuan setelah masuk.
 */
export async function requireSession(request: Request) {
  const session = await getSession()
    .then(({ data }) => data)
    .catch(() => null);

  if (!session) {
    const { pathname, search } = new URL(request.url);
    const target = `${pathname}${search}`;
    const to =
      target === "/"
        ? "/login"
        : `/login?redirect=${encodeURIComponent(target)}`;
    throw redirect(to);
  }

  return session;
}
