import { randomUUID } from "node:crypto";
// NOTE: sesuaikan import ini dengan export instance auth di @domus/better-auth
import { auth } from "@domus/better-auth/auth";
import { env } from "./env";

async function testHelpers() {
  const ctx = await auth.$context;
  if (!ctx.test) {
    throw new Error(
      "testUtils plugin belum aktif. Jalankan auth server dengan E2E=true.",
    );
  }
  return ctx.test;
}

/** Bikin user test di DB + cookie sesi siap pakai buat Playwright. */
export async function createSession(overrides: Record<string, unknown> = {}) {
  const t = await testHelpers();
  const user = t.createUser({
    email: `e2e-${randomUUID()}@example.com`,
    name: "E2E User",
    ...overrides,
  });
  const saved = await t.saveUser(user);
  const cookies = await t.getCookies({
    userId: saved.id,
    domain: new URL(env.baseUrl).hostname,
  });
  return { user: saved, cookies };
}

export async function removeUser(id: string) {
  const t = await testHelpers();
  await t.deleteUser(id);
}
