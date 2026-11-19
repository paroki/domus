-- Create "casbin_rules" table
CREATE TABLE IF NOT EXISTS "casbin_rules" (
  "id" uuid NOT NULL,
  "ptype" character varying NOT NULL,
  "v0" character varying NOT NULL DEFAULT '',
  "v1" character varying NOT NULL DEFAULT '',
  "v2" character varying NOT NULL DEFAULT '',
  "v3" character varying NOT NULL DEFAULT '',
  "v4" character varying NOT NULL DEFAULT '',
  "v5" character varying NOT NULL DEFAULT '',
  PRIMARY KEY ("id")
);
-- Create index "casbinrule_ptype_v0_v1_v2_v3" to table: "casbin_rules"
CREATE INDEX IF NOT EXISTS "casbinrule_ptype_v0_v1_v2_v3" ON "casbin_rules" ("ptype", "v0", "v1", "v2", "v3");

-- Create "users" table
CREATE TABLE IF NOT EXISTS "users" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  "email" character varying NOT NULL,
  "name" character varying NOT NULL,
  "phone" character varying NULL,
  "avatar_url" character varying NULL,
  "status" character varying NOT NULL DEFAULT 'active',
  "default_parish_id" uuid NULL,
  "last_login_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "user_email" to table: "users"
CREATE UNIQUE INDEX IF NOT EXISTS "user_email" ON "users" ("email");
-- Create index "user_status" to table: "users"
CREATE INDEX IF NOT EXISTS "user_status" ON "users" ("status");

-- Create "user_accounts" table
CREATE TABLE IF NOT EXISTS "user_accounts" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "provider" character varying NOT NULL,
  "subject" character varying NOT NULL,
  "email" character varying NULL,
  "display_name" character varying NULL,
  "avatar_url" character varying NULL,
  "active_parish_id" uuid NULL,
  "last_used_at" timestamptz NULL,
  "user_id" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "user_accounts_users_accounts" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE NO ACTION
);
-- Create index "useraccount_provider_subject" to table: "user_accounts"
CREATE UNIQUE INDEX IF NOT EXISTS "useraccount_provider_subject" ON "user_accounts" ("provider", "subject");
-- Create index "useraccount_user_id_provider" to table: "user_accounts"
CREATE INDEX IF NOT EXISTS "useraccount_user_id_provider" ON "user_accounts" ("user_id", "provider");

-- Create "user_otps" table
CREATE TABLE IF NOT EXISTS "user_otps" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "channel" character varying NOT NULL DEFAULT 'email',
  "target" character varying NOT NULL,
  "code_hash" character varying NOT NULL,
  "attempts" bigint NOT NULL DEFAULT 0,
  "expires_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "userotp_channel_target" to table: "user_otps"
CREATE INDEX IF NOT EXISTS "userotp_channel_target" ON "user_otps" ("channel", "target");
-- Create index "userotp_expires_at" to table: "user_otps"
CREATE INDEX IF NOT EXISTS "userotp_expires_at" ON "user_otps" ("expires_at");

-- Create "user_sessions" table
CREATE TABLE IF NOT EXISTS "user_sessions" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "token_hash" character varying NOT NULL,
  "user_agent" character varying NOT NULL DEFAULT '',
  "ip_address" character varying NOT NULL DEFAULT '',
  "active_parish_id" uuid NULL,
  "last_active_at" timestamptz NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "user_id" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "user_sessions_users_sessions" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE NO ACTION
);
-- Create index "usersession_token_hash" to table: "user_sessions"
CREATE UNIQUE INDEX IF NOT EXISTS "usersession_token_hash" ON "user_sessions" ("token_hash");
-- Create index "usersession_user_id_expires_at" to table: "user_sessions"
CREATE INDEX IF NOT EXISTS "usersession_user_id_expires_at" ON "user_sessions" ("user_id", "expires_at");
