import { Database } from "bun:sqlite";
import { createHash, randomBytes, randomUUID } from "node:crypto";
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

function openDb(): Database {
  return new Database(config.dbPath);
}

function formatSqliteTime(d: Date): string {
  return d.toISOString().replace("T", " ").replace("Z", " +0000 UTC");
}

export function loginAsUser(options: {
  email: string;
  name?: string;
  onboarded?: boolean;
  activeParishId?: string;
}): AuthSession {
  const db = openDb();
  try {
    const now = formatSqliteTime(new Date());
    const expiresAt = formatSqliteTime(
      new Date(Date.now() + 24 * 60 * 60 * 1000),
    );
    const email = options.email.trim().toLowerCase();
    const name = options.name ?? email;
    const onboarded = (options.onboarded ?? true) ? 1 : 0;

    const existing = db
      .query<{ id: string }, [string]>(
        "SELECT id FROM users WHERE email = ? AND deleted_at IS NULL",
      )
      .get(email);

    let userId: string;
    if (existing) {
      userId = existing.id;
      db.query(
        "UPDATE users SET status = 'active', onboarded = ?, updated_at = ? WHERE id = ?",
      ).run(onboarded, now, userId);
    } else {
      userId = randomUUID();
      db.query(
        "INSERT INTO users (id, created_at, updated_at, email, name, status, onboarded) VALUES (?, ?, ?, ?, ?, 'active', ?)",
      ).run(userId, now, now, email, name, onboarded);
    }

    const account = db
      .query<{ id: string }, [string, string]>(
        "SELECT id FROM user_accounts WHERE provider = ? AND subject = ?",
      )
      .get("email", email);

    if (!account) {
      db.query(
        "INSERT INTO user_accounts (id, created_at, updated_at, provider, subject, email, display_name, user_id) VALUES (?, ?, ?, 'email', ?, ?, ?, ?)",
      ).run(randomUUID(), now, now, email, email, name, userId);
    }

    const rawToken = randomBytes(32).toString("base64url");
    db.query(
      "INSERT INTO user_sessions (id, created_at, updated_at, token_hash, user_agent, ip_address, active_parish_id, last_active_at, expires_at, user_id) VALUES (?, ?, ?, ?, 'Playwright E2E', '127.0.0.1', ?, ?, ?, ?)",
    ).run(
      randomUUID(),
      now,
      now,
      sha256Hex(rawToken),
      options.activeParishId ?? null,
      now,
      expiresAt,
      userId,
    );

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
    db.close();
  }
}

export function loginAsSuperAdmin(options?: {
  email?: string;
  name?: string;
  onboarded?: boolean;
  activeParishId?: string;
}): AuthSession {
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

  const superAdmin = loginAsSuperAdmin();
  client.setToken(superAdmin.rawToken);

  const dioceseCode = `e2e-d-${suffix}`;
  const dioceseName = `Keuskupan E2E ${suffix}`;
  const dioceseRes = await client.post<{
    data?: { id?: string; code?: string; name?: string };
  }>("/ref/dioceses", {
    code: dioceseCode,
    name: dioceseName,
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
    data?: { id?: string; code?: string; name?: string };
  }>("/ref/parishes", {
    dioceseId,
    code: parishCode,
    name: parishName,
  });
  if (parishRes.status !== 201 || !parishRes.data?.data?.id) {
    throw new Error(
      `Failed to bootstrap parish: ${parishRes.status} ${JSON.stringify(parishRes.data)}`,
    );
  }
  const parishId = parishRes.data.data.id;

  const email = options?.email ?? "parishadmin.e2e@domus.test";
  const name = options?.name ?? "Admin Paroki E2E";
  const memberRes = await client.post(`/ref/parishes/${parishId}/members`, {
    email,
    name,
    role: "admin",
  });
  if (memberRes.status !== 201 && memberRes.status !== 409) {
    throw new Error(
      `Failed to appoint parish admin: ${memberRes.status} ${JSON.stringify(memberRes.data)}`,
    );
  }

  const session = loginAsUser({
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

export function seedOtpCode(
  email: string,
  code = "123456",
  channel = "email",
): void {
  const db = openDb();
  try {
    const now = formatSqliteTime(new Date());
    const expiresAt = formatSqliteTime(new Date(Date.now() + 5 * 60 * 1000));
    const target = email.trim().toLowerCase();

    db.query("DELETE FROM user_otps WHERE channel = ? AND target = ?").run(
      channel,
      target,
    );
    db.query(
      "INSERT INTO user_otps (id, created_at, updated_at, channel, target, code_hash, attempts, expires_at) VALUES (?, ?, ?, ?, ?, ?, 0, ?)",
    ).run(randomUUID(), now, now, channel, target, sha256Hex(code), expiresAt);
  } finally {
    db.close();
  }
}
