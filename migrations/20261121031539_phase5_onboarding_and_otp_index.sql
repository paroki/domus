-- Modify "users" table
ALTER TABLE "users" ADD COLUMN "onboarded" boolean NOT NULL DEFAULT false;
-- Create index "userotp_channel_code_hash" to table: "user_otps"
CREATE INDEX IF NOT EXISTS "userotp_channel_code_hash" ON "user_otps" ("channel", "code_hash");
