import { createHash, randomBytes, randomUUID } from "node:crypto";
import { SQL } from "bun";
import { config } from "./config";
import type { TestClient } from "./test-client";

export interface SessionCookie {
  name: string;
  value: string;
  domain: string;
  path: string;
  httpOnly: boolean;
  sameSite: "Lax" | "Strict" | "None";
}

export interface AuthSession {
  userId: string;
  rawToken: string;
  cookies: SessionCookie[];
}

function sha256Hex(value: string): string {
  return createHash("sha256").update(value).digest("hex");
}

function openDb(): SQL {
  return new SQL(config.appDbUrl);
}

export async function loginAsUser(options: {
  email: string;
  name?: string;
  onboarded?: boolean;
  activeParokiId?: string;
}): Promise<AuthSession> {
  const db = openDb();
  try {
    const now = new Date().toISOString();
    const expiresAt = new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString();
    const email = options.email.trim().toLowerCase();
    const name = options.name ?? email;
    const onboarded = options.onboarded ?? true;

    const existing = await db<{ id: string }[]>`
      SELECT id FROM users WHERE email = ${email} AND deleted_at IS NULL
    `;

    let userId: string;
    if (existing.length > 0 && existing[0]) {
      userId = existing[0].id;
      await db`
        UPDATE users SET status = 'active', onboarded = ${onboarded}, updated_at = ${now} WHERE id = ${userId}
      `;
    } else {
      userId = randomUUID();
      await db`
        INSERT INTO users (id, created_at, updated_at, email, name, status, onboarded)
        VALUES (${userId}, ${now}, ${now}, ${email}, ${name}, 'active', ${onboarded})
      `;
    }

    const account = await db<{ id: string }[]>`
      SELECT id FROM user_accounts WHERE provider = ${"email"} AND subject = ${email}
    `;

    if (account.length === 0) {
      await db`
        INSERT INTO user_accounts (id, created_at, updated_at, provider, subject, email, display_name, user_id)
        VALUES (${randomUUID()}, ${now}, ${now}, 'email', ${email}, ${email}, ${name}, ${userId})
      `;
    }

    const rawToken = randomBytes(32).toString("base64url");
    await db`
      INSERT INTO user_sessions (id, created_at, updated_at, token_hash, user_agent, ip_address, active_paroki_id, last_active_at, expires_at, user_id)
      VALUES (${randomUUID()}, ${now}, ${now}, ${sha256Hex(rawToken)}, 'Playwright E2E', '127.0.0.1', ${options.activeParokiId ?? null}, ${now}, ${expiresAt}, ${userId})
    `;

    return {
      userId,
      rawToken,
      cookies: [
        {
          name: "domus_session",
          value: rawToken,
          domain: "localhost",
          path: "/",
          httpOnly: true,
          sameSite: "Lax",
        },
      ],
    };
  } finally {
    await db.close();
  }
}

export async function loginAsSuperAdmin(options?: {
  email?: string;
  name?: string;
  onboarded?: boolean;
  activeParokiId?: string;
}): Promise<AuthSession> {
  return loginAsUser({
    email: options?.email ?? config.superAdminEmail,
    name: options?.name ?? config.superAdminName,
    onboarded: options?.onboarded ?? true,
    activeParokiId: options?.activeParokiId,
  });
}

let parokiBootstrapCounter = 0;

export async function bootstrapParokiAdminSession(
  client: TestClient,
  options?: {
    email?: string;
    name?: string;
    parokiCodeSuffix?: string;
  },
): Promise<{
  session: AuthSession;
  keuskupan: { id: string; kode: string; nama: string };
  paroki: { id: string; kode: string; nama: string };
}> {
  parokiBootstrapCounter += 1;
  const suffix =
    options?.parokiCodeSuffix ??
    `${Date.now().toString(36)}-${parokiBootstrapCounter}`;

  const superAdmin = await loginAsSuperAdmin();
  client.setToken(superAdmin.rawToken);

  const keuskupanKode = `e2e-d-${suffix}`;
  const keuskupanNama = `Keuskupan E2E ${suffix}`;
  const keuskupanRes = await client.post<{
    data?: { id?: string; kode?: string; nama?: string };
  }>("/ref/keuskupan", {
    kode: keuskupanKode,
    nama: keuskupanNama,
  });
  if (keuskupanRes.status !== 201 || !keuskupanRes.data?.data?.id) {
    throw new Error(
      `Failed to bootstrap keuskupan: ${keuskupanRes.status} ${JSON.stringify(keuskupanRes.data)}`,
    );
  }
  const keuskupanId = keuskupanRes.data.data.id;

  const parokiKode = `e2e-p-${suffix}`;
  const parokiNama = `Paroki E2E ${suffix}`;
  const parokiRes = await client.post<{
    data?: { id?: string; kode?: string; nama?: string };
  }>("/ref/paroki", {
    keuskupanId,
    kode: parokiKode,
    nama: parokiNama,
  });
  if (parokiRes.status !== 201 || !parokiRes.data?.data?.id) {
    throw new Error(
      `Failed to bootstrap paroki: ${parokiRes.status} ${JSON.stringify(parokiRes.data)}`,
    );
  }
  const parokiId = parokiRes.data.data.id;

  const email = options?.email ?? "parokiadmin.e2e@domus.test";
  const name = options?.name ?? "Admin Paroki E2E";
  const memberRes = await client.post(`/ref/paroki/${parokiId}/members`, {
    email,
    name,
    role: "admin",
  });
  if (memberRes.status !== 201 && memberRes.status !== 409) {
    throw new Error(
      `Failed to appoint paroki admin: ${memberRes.status} ${JSON.stringify(memberRes.data)}`,
    );
  }

  const session = await loginAsUser({
    email,
    name,
    onboarded: true,
    activeParokiId: parokiId,
  });
  client.setToken(session.rawToken);

  return {
    session,
    keuskupan: { id: keuskupanId, kode: keuskupanKode, nama: keuskupanNama },
    paroki: { id: parokiId, kode: parokiKode, nama: parokiNama },
  };
}

export async function bootstrapOrgAdminSession(
  client: TestClient,
  options?: {
    email?: string;
    name?: string;
    orgName?: string;
    periodName?: string;
  },
): Promise<{
  session: AuthSession;
  paroki: { id: string; kode: string; nama: string };
  org: { id: string; code: string; name: string };
  period: { id: string; name: string };
  parokiAdminToken: string;
}> {
  const { session: parokiAdminSession, paroki } =
    await bootstrapParokiAdminSession(client);
  client.setParokiId(paroki.id);
  const suffix = `${Date.now().toString(36)}-${parokiBootstrapCounter}`;

  const periodName = options?.periodName ?? "Periode 2025-2028";
  const periodRes = await client.post<{
    data?: { id?: string; name?: string };
  }>("/ref/periods", {
    type: "fiscal",
    code: `fy-2025-2028-${suffix}`,
    name: periodName,
    startDate: "2025-01-01",
    endDate: "2028-12-31",
    isCurrent: true,
  });
  if (periodRes.status !== 201 || !periodRes.data?.data?.id) {
    throw new Error(
      `Failed to bootstrap period: ${periodRes.status} ${JSON.stringify(periodRes.data)}`,
    );
  }

  const lookupRes = await client.post("/ref/lookups", {
    category: "org_position",
    code: `sekretaris-e2e-${suffix}`,
    label: "Sekretaris E2E",
    sortOrder: 0,
    global: false,
  });
  if (lookupRes.status !== 201 && lookupRes.status !== 409) {
    throw new Error(
      `Failed to bootstrap lookup: ${lookupRes.status} ${JSON.stringify(lookupRes.data)}`,
    );
  }

  const orgCode = `liturgi-e2e-${suffix}`;
  const orgName = options?.orgName ?? "Seksi Liturgi E2E";
  const orgRes = await client.post<{
    data?: { id?: string; code?: string; name?: string };
  }>("/orgs", {
    kind: "section",
    code: orgCode,
    name: orgName,
  });
  if (orgRes.status !== 201 || !orgRes.data?.data?.id) {
    throw new Error(
      `Failed to bootstrap org: ${orgRes.status} ${JSON.stringify(orgRes.data)}`,
    );
  }
  const orgId = orgRes.data.data.id;

  const email = options?.email ?? "orgadmin.e2e@domus.test";
  const name = options?.name ?? "Admin Organisasi E2E";
  const memberRes = await client.post(`/orgs/${orgId}/members`, {
    email,
    name,
    role: "admin",
  });
  if (memberRes.status !== 201 && memberRes.status !== 409) {
    throw new Error(
      `Failed to appoint org admin: ${memberRes.status} ${JSON.stringify(memberRes.data)}`,
    );
  }

  const session = await loginAsUser({
    email,
    name,
    onboarded: true,
    activeParokiId: paroki.id,
  });
  client.setToken(session.rawToken);

  return {
    session,
    paroki,
    org: { id: orgId, code: orgCode, name: orgName },
    period: { id: periodRes.data.data.id, name: periodName },
    parokiAdminToken: parokiAdminSession.rawToken,
  };
}

export async function bootstrapOrgMemberSession(
  client: TestClient,
  options?: {
    email?: string;
    name?: string;
    orgName?: string;
    otherOrgName?: string;
  },
): Promise<{
  session: AuthSession;
  paroki: { id: string; kode: string; nama: string };
  orgA: { id: string; code: string; name: string };
  orgB: { id: string; code: string; name: string };
  orgAdminToken: string;
  parokiAdminToken: string;
}> {
  const adminBoot = await bootstrapOrgAdminSession(client, {
    orgName: options?.orgName,
  });
  const orgAdminToken = adminBoot.session.rawToken;
  const parokiAdminToken = adminBoot.parokiAdminToken;

  client.setParokiId(adminBoot.paroki.id);
  client.setToken(orgAdminToken);

  const email = options?.email ?? "member.e2e@domus.test";
  const name = options?.name ?? "Anggota Organisasi E2E";
  const memberRes = await client.post(`/orgs/${adminBoot.org.id}/members`, {
    email,
    name,
    role: "member",
    position: "Anggota Koor",
  });
  if (memberRes.status !== 201 && memberRes.status !== 409) {
    throw new Error(
      `Failed to bootstrap org member: ${memberRes.status} ${JSON.stringify(memberRes.data)}`,
    );
  }

  client.setToken(parokiAdminToken);
  const suffix = `${Date.now().toString(36)}-${parokiBootstrapCounter}`;
  const orgBCode = `katekese-e2e-${suffix}`;
  const orgBName = options?.otherOrgName ?? "Seksi Katekese E2E";
  const orgBRes = await client.post<{
    data?: { id?: string; code?: string; name?: string };
  }>("/orgs", {
    kind: "section",
    code: orgBCode,
    name: orgBName,
  });
  if (orgBRes.status !== 201 || !orgBRes.data?.data?.id) {
    throw new Error(
      `Failed to bootstrap sibling org: ${orgBRes.status} ${JSON.stringify(orgBRes.data)}`,
    );
  }
  const orgBId = orgBRes.data.data.id;

  const orgBAdminRes = await client.post(`/orgs/${orgBId}/members`, {
    email: "orgadmin.e2e@domus.test",
    name: "Admin Organisasi E2E",
    role: "admin",
  });
  if (orgBAdminRes.status !== 201 && orgBAdminRes.status !== 409) {
    throw new Error(
      `Failed to appoint sibling org admin: ${orgBAdminRes.status} ${JSON.stringify(orgBAdminRes.data)}`,
    );
  }

  const session = await loginAsUser({
    email,
    name,
    onboarded: true,
    activeParokiId: adminBoot.paroki.id,
  });
  client.setToken(session.rawToken);

  return {
    session,
    paroki: adminBoot.paroki,
    orgA: adminBoot.org,
    orgB: { id: orgBId, code: orgBCode, name: orgBName },
    orgAdminToken,
    parokiAdminToken,
  };
}

export async function seedOtpCode(
  email: string,
  code = "123456",
  channel = "email",
): Promise<void> {
  const db = openDb();
  try {
    const now = new Date().toISOString();
    const expiresAt = new Date(Date.now() + 5 * 60 * 1000).toISOString();
    const target = email.trim().toLowerCase();

    await db`DELETE FROM user_otps WHERE channel = ${channel} AND target = ${target}`;
    await db`
      INSERT INTO user_otps (id, created_at, updated_at, channel, target, code_hash, attempts, expires_at)
      VALUES (${randomUUID()}, ${now}, ${now}, ${channel}, ${target}, ${sha256Hex(code)}, 0, ${expiresAt})
    `;
  } finally {
    await db.close();
  }
}
