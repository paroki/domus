import { requireSession } from "~/lib/auth";
import type { Route } from "./+types/_app";

/** Semua route di bawah `_app` butuh login; yang belum masuk dialihkan ke /login. */
export async function clientLoader({ request }: Route.ClientLoaderArgs) {
  await requireSession(request);
  return null;
}

export { default } from "~/shared/layout/AppShell";
