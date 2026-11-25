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
  activeParishId?: string;
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
      VALUES (${randomUUID()}, ${now}, ${now}, ${sha256Hex(rawToken)}, 'Playwright E2E', '127.0.0.1', ${options.activeParishId ?? null}, ${now}, ${expiresAt}, ${userId})
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
  activeParishId?: string;
}): Promise<AuthSession> {
  return loginAsUser({
    email: options?.email ?? config.superAdminEmail,
    name: options?.name ?? config.superAdminName,
    onboarded: options?.onboarded ?? true,
    activeParishId: options?.activeParishId,
  });
}

let parishBootstrapCounter = 0;

export async function bootstrapParishAdminSession(
  client: TestClient,
  options?: {
    email?: string;
    name?: string;
    parishCodeSuffix?: string;
  },
): Promise<{
  session: AuthSession;
  diocese: { id: string; code: string; name: string };
  parish: { id: string; code: string; name: string };
}> {
  parishBootstrapCounter += 1;
  const suffix =
    options?.parishCodeSuffix ??
    `${Date.now().toString(36)}-${parishBootstrapCounter}`;

  const superAdmin = await loginAsSuperAdmin();
  client.setToken(superAdmin.rawToken);

  const dioceseCode = `e2e-d-${suffix}`;
  const dioceseName = `Keuskupan E2E ${suffix}`;
  const dioceseRes = await client.post<{
    data?: { id?: string; kode?: string; nama?: string };
  }>("/ref/keuskupan", {
    kode: dioceseCode,
    nama: dioceseName,
  });
  if (dioceseRes.status !== 201 || !dioceseRes.data?.data?.id) {
    throw new Error(
      `Failed to bootstrap diocese: ${dioceseRes.status} ${JSON.stringify(dioceseRes.data)}`,
    );
  }
  const dioceseId = dioceseRes.data.data.id;

  const parishCode = `e2e-p-${suffix}`;
  const parishName = `Paroki E2E ${suffix}`;
  const parishRes = await client.post<{
    data?: { id?: string; kode?: string; nama?: string };
  }>("/ref/paroki", {
    keuskupanId: dioceseId,
    kode: parishCode,
    nama: parishName,
  });
  if (parishRes.status !== 201 || !parishRes.data?.data?.id) {
    throw new Error(
      `Failed to bootstrap parish: ${parishRes.status} ${JSON.stringify(parishRes.data)}`,
    );
  }
  const parishId = parishRes.data.data.id;

  const email = options?.email ?? "parishadmin.e2e@domus.test";
  const name = options?.name ?? "Admin Paroki E2E";
  const memberRes = await client.post(`/ref/paroki/${parishId}/members`, {
    email,
    name,
    role: "admin",
  });
  if (memberRes.status !== 201 && memberRes.status !== 409) {
    throw new Error(
      `Failed to appoint parish admin: ${memberRes.status} ${JSON.stringify(memberRes.data)}`,
    );
  }

  const session = await loginAsUser({
    email,
    name,
    onboarded: true,
    activeParishId: parishId,
  });
  client.setToken(session.rawToken);

  return {
    session,
    diocese: { id: dioceseId, code: dioceseCode, name: dioceseName },
    parish: { id: parishId, code: parishCode, name: parishName },
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
  parish: { id: string; code: string; name: string };
  org: { id: string; code: string; name: string };
  period: { id: string; name: string };
  parishAdminToken: string;
}> {
  const { session: parishAdminSession, parish } =
    await bootstrapParishAdminSession(client);
  client.setParishId(parish.id);
  const suffix = `${Date.now().toString(36)}-${parishBootstrapCounter}`;

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
    activeParishId: parish.id,
  });
  client.setToken(session.rawToken);

  return {
    session,
    parish,
    org: { id: orgId, code: orgCode, name: orgName },
    period: { id: periodRes.data.data.id, name: periodName },
    parishAdminToken: parishAdminSession.rawToken,
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
  parish: { id: string; code: string; name: string };
  orgA: { id: string; code: string; name: string };
  orgB: { id: string; code: string; name: string };
  orgAdminToken: string;
  parishAdminToken: string;
}> {
  const adminBoot = await bootstrapOrgAdminSession(client, {
    orgName: options?.orgName,
  });
  const orgAdminToken = adminBoot.session.rawToken;
  const parishAdminToken = adminBoot.parishAdminToken;

  client.setParishId(adminBoot.parish.id);
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

  client.setToken(parishAdminToken);
  const suffix = `${Date.now().toString(36)}-${parishBootstrapCounter}`;
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
    activeParishId: adminBoot.parish.id,
  });
  client.setToken(session.rawToken);

  return {
    session,
    parish: adminBoot.parish,
    orgA: adminBoot.org,
    orgB: { id: orgBId, code: orgBCode, name: orgBName },
    orgAdminToken,
    parishAdminToken,
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
