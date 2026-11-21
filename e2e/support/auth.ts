import { Database } from "bun:sqlite";
import { createHash, randomBytes, randomUUID } from "node:crypto";
import { config } from "./config";

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
      "INSERT INTO user_sessions (id, created_at, updated_at, token_hash, user_agent, ip_address, last_active_at, expires_at, user_id) VALUES (?, ?, ?, ?, 'Playwright E2E', '127.0.0.1', ?, ?, ?)",
    ).run(randomUUID(), now, now, sha256Hex(rawToken), now, expiresAt, userId);

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
}): AuthSession {
  return loginAsUser({
    email: options?.email ?? config.superAdminEmail,
    name: options?.name ?? config.superAdminName,
    onboarded: options?.onboarded ?? true,
  });
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
