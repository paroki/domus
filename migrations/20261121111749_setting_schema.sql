-- Create index "dioceses_code_key" to table: "dioceses"
CREATE UNIQUE INDEX "dioceses_code_key" ON "dioceses" ("code");
-- Create index "parishes_code_key" to table: "parishes"
CREATE UNIQUE INDEX "parishes_code_key" ON "parishes" ("code");
-- Create index "user_sessions_token_hash_key" to table: "user_sessions"
CREATE UNIQUE INDEX "user_sessions_token_hash_key" ON "user_sessions" ("token_hash");
-- Create index "users_email_key" to table: "users"
CREATE UNIQUE INDEX "users_email_key" ON "users" ("email");
-- Create "settings" table
CREATE TABLE "settings" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "key" character varying NOT NULL,
  "value" character varying NOT NULL DEFAULT '',
  PRIMARY KEY ("id")
);
-- Create index "settings_key_key" to table: "settings"
CREATE UNIQUE INDEX "settings_key_key" ON "settings" ("key");
-- Drop "meta" table
DROP TABLE "meta";
