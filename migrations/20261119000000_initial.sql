-- Create "meta" table
CREATE TABLE IF NOT EXISTS "meta" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "key" character varying NOT NULL,
  "value" character varying NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "meta_key_key" to table: "meta"
CREATE UNIQUE INDEX IF NOT EXISTS "meta_key_key" ON "meta" ("key");
