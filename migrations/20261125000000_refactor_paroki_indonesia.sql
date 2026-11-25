-- Refactor ecclesiastical & territorial tables and tenant columns to Bahasa Indonesia.

-- 1. Drop old English ref tables and recreate Indonesian tables.
DROP TABLE IF EXISTS "wards" CASCADE;
DROP TABLE IF EXISTS "zones" CASCADE;
DROP TABLE IF EXISTS "parish_members" CASCADE;
DROP TABLE IF EXISTS "periods" CASCADE;
DROP TABLE IF EXISTS "lookups" CASCADE;
DROP TABLE IF EXISTS "parishes" CASCADE;
DROP TABLE IF EXISTS "dioceses" CASCADE;

-- Create "keuskupan" table
CREATE TABLE IF NOT EXISTS "keuskupan" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  "kode" character varying NOT NULL,
  "nama" character varying NOT NULL,
  "uskup" character varying NULL,
  "alamat" character varying NULL,
  "telepon" character varying NULL,
  "email" character varying NULL,
  "website" character varying NULL,
  "active" boolean NOT NULL DEFAULT true,
  PRIMARY KEY ("id")
);
CREATE UNIQUE INDEX IF NOT EXISTS "keuskupan_kode" ON "keuskupan" ("kode");
CREATE INDEX IF NOT EXISTS "keuskupan_active" ON "keuskupan" ("active");

-- Create "paroki" table
CREATE TABLE IF NOT EXISTS "paroki" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  "kode" character varying NOT NULL,
  "nama" character varying NOT NULL,
  "pelindung" character varying NULL,
  "pastor" character varying NULL,
  "alamat" character varying NULL,
  "telepon" character varying NULL,
  "email" character varying NULL,
  "website" character varying NULL,
  "logo_url" character varying NULL,
  "kop_surat" character varying NULL,
  "active" boolean NOT NULL DEFAULT true,
  "keuskupan_id" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "paroki_keuskupan_paroki" FOREIGN KEY ("keuskupan_id") REFERENCES "keuskupan" ("id") ON DELETE NO ACTION
);
CREATE UNIQUE INDEX IF NOT EXISTS "paroki_kode" ON "paroki" ("kode");
CREATE INDEX IF NOT EXISTS "paroki_keuskupan_id_active" ON "paroki" ("keuskupan_id", "active");

-- Create "paroki_members" table
CREATE TABLE IF NOT EXISTS "paroki_members" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "user_id" uuid NOT NULL,
  "role" character varying NOT NULL DEFAULT 'member',
  "paroki_id" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "paroki_members_paroki_members" FOREIGN KEY ("paroki_id") REFERENCES "paroki" ("id") ON DELETE NO ACTION
);
CREATE UNIQUE INDEX IF NOT EXISTS "parokimember_paroki_id_user_id" ON "paroki_members" ("paroki_id", "user_id");
CREATE INDEX IF NOT EXISTS "parokimember_user_id" ON "paroki_members" ("user_id");

-- Create "wilayah" table
CREATE TABLE IF NOT EXISTS "wilayah" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  "jenis" character varying NOT NULL DEFAULT 'wilayah',
  "kode" character varying NOT NULL,
  "nama" character varying NOT NULL,
  "koordinator" character varying NULL,
  "telepon_koordinator" character varying NULL,
  "active" boolean NOT NULL DEFAULT true,
  "paroki_id" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "wilayah_paroki_wilayah" FOREIGN KEY ("paroki_id") REFERENCES "paroki" ("id") ON DELETE NO ACTION
);
CREATE UNIQUE INDEX IF NOT EXISTS "wilayah_paroki_id_kode" ON "wilayah" ("paroki_id", "kode");
CREATE INDEX IF NOT EXISTS "wilayah_paroki_id_jenis_active" ON "wilayah" ("paroki_id", "jenis", "active");

-- Create "lingkungan" table
CREATE TABLE IF NOT EXISTS "lingkungan" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  "kode" character varying NOT NULL,
  "nama" character varying NOT NULL,
  "pelindung" character varying NULL,
  "ketua" character varying NULL,
  "telepon_ketua" character varying NULL,
  "active" boolean NOT NULL DEFAULT true,
  "paroki_id" uuid NOT NULL,
  "wilayah_id" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "lingkungan_paroki_lingkungan" FOREIGN KEY ("paroki_id") REFERENCES "paroki" ("id") ON DELETE NO ACTION,
  CONSTRAINT "lingkungan_wilayah_lingkungan" FOREIGN KEY ("wilayah_id") REFERENCES "wilayah" ("id") ON DELETE NO ACTION
);
CREATE UNIQUE INDEX IF NOT EXISTS "lingkungan_paroki_id_kode" ON "lingkungan" ("paroki_id", "kode");
CREATE INDEX IF NOT EXISTS "lingkungan_paroki_id_wilayah_id_active" ON "lingkungan" ("paroki_id", "wilayah_id", "active");

-- Create "lookups" table
CREATE TABLE IF NOT EXISTS "lookups" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  "category" character varying NOT NULL,
  "code" character varying NOT NULL,
  "label" character varying NOT NULL,
  "sort_order" bigint NOT NULL DEFAULT 0,
  "active" boolean NOT NULL DEFAULT true,
  "paroki_id" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "lookups_paroki_lookups" FOREIGN KEY ("paroki_id") REFERENCES "paroki" ("id") ON DELETE SET NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS "lookup_paroki_id_category_code" ON "lookups" ("paroki_id", "category", "code");
CREATE UNIQUE INDEX IF NOT EXISTS "lookup_global_category_code" ON "lookups" ("category", "code") WHERE "paroki_id" IS NULL;
CREATE INDEX IF NOT EXISTS "lookup_paroki_id_category_sort_order" ON "lookups" ("paroki_id", "category", "sort_order");

-- Create "periods" table
CREATE TABLE IF NOT EXISTS "periods" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  "type" character varying NOT NULL DEFAULT 'fiscal',
  "code" character varying NOT NULL,
  "name" character varying NOT NULL,
  "start_date" character varying NOT NULL,
  "end_date" character varying NOT NULL,
  "status" character varying NOT NULL DEFAULT 'open',
  "is_current" boolean NOT NULL DEFAULT false,
  "paroki_id" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "periods_paroki_periods" FOREIGN KEY ("paroki_id") REFERENCES "paroki" ("id") ON DELETE NO ACTION
);
CREATE UNIQUE INDEX IF NOT EXISTS "period_paroki_id_code" ON "periods" ("paroki_id", "code");
CREATE INDEX IF NOT EXISTS "period_paroki_id_type_is_current" ON "periods" ("paroki_id", "type", "is_current");

-- 2. Update user, org, event, and agent tenant/territorial columns and indexes.
ALTER TABLE "users" RENAME COLUMN "default_parish_id" TO "default_paroki_id";
ALTER TABLE "user_accounts" RENAME COLUMN "active_parish_id" TO "active_paroki_id";
ALTER TABLE "user_sessions" RENAME COLUMN "active_parish_id" TO "active_paroki_id";

DROP INDEX IF EXISTS "organization_parish_id_code";
DROP INDEX IF EXISTS "organization_parish_id_kind_active";
DROP INDEX IF EXISTS "organization_parish_id_parent_id";
DROP INDEX IF EXISTS "organization_parish_id_ward_id";
DROP INDEX IF EXISTS "organization_parish_id_zone_id";
ALTER TABLE "organizations" RENAME COLUMN "parish_id" TO "paroki_id";
ALTER TABLE "organizations" RENAME COLUMN "zone_id" TO "wilayah_id";
ALTER TABLE "organizations" RENAME COLUMN "ward_id" TO "lingkungan_id";
UPDATE "organizations" SET "kind" = 'wilayah' WHERE "kind" = 'zone';
UPDATE "organizations" SET "kind" = 'lingkungan' WHERE "kind" = 'ward';
CREATE UNIQUE INDEX IF NOT EXISTS "organization_paroki_id_code" ON "organizations" ("paroki_id", "code");
CREATE INDEX IF NOT EXISTS "organization_paroki_id_kind_active" ON "organizations" ("paroki_id", "kind", "active");
CREATE INDEX IF NOT EXISTS "organization_paroki_id_parent_id" ON "organizations" ("paroki_id", "parent_id");
CREATE INDEX IF NOT EXISTS "organization_paroki_id_wilayah_id" ON "organizations" ("paroki_id", "wilayah_id");
CREATE INDEX IF NOT EXISTS "organization_paroki_id_lingkungan_id" ON "organizations" ("paroki_id", "lingkungan_id");

DROP INDEX IF EXISTS "orginvite_parish_id_org_id";
ALTER TABLE "org_invites" RENAME COLUMN "parish_id" TO "paroki_id";
CREATE INDEX IF NOT EXISTS "orginvite_paroki_id_org_id" ON "org_invites" ("paroki_id", "org_id");

DROP INDEX IF EXISTS "orgmember_parish_id_user_id";
ALTER TABLE "org_members" RENAME COLUMN "parish_id" TO "paroki_id";
CREATE INDEX IF NOT EXISTS "orgmember_paroki_id_user_id" ON "org_members" ("paroki_id", "user_id");

DROP INDEX IF EXISTS "event_parish_id_start_at";
DROP INDEX IF EXISTS "event_parish_id_org_id_start_at";
DROP INDEX IF EXISTS "event_parish_id_visibility_status_start_at";
ALTER TABLE "events" RENAME COLUMN "parish_id" TO "paroki_id";
CREATE INDEX IF NOT EXISTS "event_paroki_id_start_at" ON "events" ("paroki_id", "start_at");
CREATE INDEX IF NOT EXISTS "event_paroki_id_org_id_start_at" ON "events" ("paroki_id", "org_id", "start_at");
CREATE INDEX IF NOT EXISTS "event_paroki_id_visibility_status_start_at" ON "events" ("paroki_id", "visibility", "status", "start_at");

DROP INDEX IF EXISTS "eventrsvp_parish_id_user_id";
ALTER TABLE "event_rsvps" RENAME COLUMN "parish_id" TO "paroki_id";
CREATE INDEX IF NOT EXISTS "eventrsvp_paroki_id_user_id" ON "event_rsvps" ("paroki_id", "user_id");

DROP INDEX IF EXISTS "agentsession_parish_id_user_id_last_message_at";
ALTER TABLE "agent_sessions" RENAME COLUMN "parish_id" TO "paroki_id";
CREATE INDEX IF NOT EXISTS "agentsession_paroki_id_user_id_last_message_at" ON "agent_sessions" ("paroki_id", "user_id", "last_message_at");

DROP INDEX IF EXISTS "agentmessage_parish_id_user_id_created_at";
ALTER TABLE "agent_messages" RENAME COLUMN "parish_id" TO "paroki_id";
CREATE INDEX IF NOT EXISTS "agentmessage_paroki_id_user_id_created_at" ON "agent_messages" ("paroki_id", "user_id", "created_at");

-- 3. Rebuild agent_memories index & BM25 index on paroki_id.
DROP INDEX IF EXISTS "agent_memories_bm25_idx";
DROP INDEX IF EXISTS "agentmemory_parish_id_user_id_created_at";
ALTER TABLE "agent_memories" RENAME COLUMN "parish_id" TO "paroki_id";
CREATE INDEX IF NOT EXISTS "agentmemory_paroki_id_user_id_created_at" ON "agent_memories" ("paroki_id", "user_id", "created_at");
CREATE INDEX IF NOT EXISTS "agent_memories_bm25_idx" ON "agent_memories" USING bm25 ("id", "content", "category", "paroki_id", "user_id") WITH (key_field = 'id');

-- 4. Update Casbin rules for Indonesian domain & role/resource identifiers.
UPDATE "casbin_rules" SET "v1" = 'ref:paroki-admin' WHERE "v1" = 'ref:parish-admin';
UPDATE "casbin_rules" SET "v1" = 'ref:keuskupan' WHERE "v1" = 'ref:dioceses';
UPDATE "casbin_rules" SET "v1" = 'ref:paroki' WHERE "v1" = 'ref:parishes';
UPDATE "casbin_rules" SET "v2" = 'ref:keuskupan' WHERE "v2" = 'ref:dioceses';
UPDATE "casbin_rules" SET "v2" = 'ref:paroki' WHERE "v2" = 'ref:parishes';
UPDATE "casbin_rules" SET "v1" = regexp_replace("v1", '^parish:', 'paroki:') WHERE "v1" LIKE 'parish:%';
UPDATE "casbin_rules" SET "v2" = regexp_replace("v2", '^parish:', 'paroki:') WHERE "v2" LIKE 'parish:%';
