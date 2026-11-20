-- Create "dioceses" table
CREATE TABLE IF NOT EXISTS "dioceses" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  "code" character varying NOT NULL,
  "name" character varying NOT NULL,
  "bishop_name" character varying NULL,
  "address" character varying NULL,
  "phone" character varying NULL,
  "email" character varying NULL,
  "website" character varying NULL,
  "active" boolean NOT NULL DEFAULT true,
  PRIMARY KEY ("id")
);
-- Create index "diocese_code" to table: "dioceses"
CREATE UNIQUE INDEX IF NOT EXISTS "diocese_code" ON "dioceses" ("code");
-- Create index "diocese_active" to table: "dioceses"
CREATE INDEX IF NOT EXISTS "diocese_active" ON "dioceses" ("active");

-- Create "parishes" table
CREATE TABLE IF NOT EXISTS "parishes" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  "code" character varying NOT NULL,
  "name" character varying NOT NULL,
  "patron_saint" character varying NULL,
  "pastor_name" character varying NULL,
  "address" character varying NULL,
  "phone" character varying NULL,
  "email" character varying NULL,
  "website" character varying NULL,
  "logo_url" character varying NULL,
  "letterhead" character varying NULL,
  "active" boolean NOT NULL DEFAULT true,
  "diocese_id" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "parishes_dioceses_parishes" FOREIGN KEY ("diocese_id") REFERENCES "dioceses" ("id") ON DELETE NO ACTION
);
-- Create index "parish_code" to table: "parishes"
CREATE UNIQUE INDEX IF NOT EXISTS "parish_code" ON "parishes" ("code");
-- Create index "parish_diocese_id_active" to table: "parishes"
CREATE INDEX IF NOT EXISTS "parish_diocese_id_active" ON "parishes" ("diocese_id", "active");

-- Create "parish_members" table
CREATE TABLE IF NOT EXISTS "parish_members" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "user_id" uuid NOT NULL,
  "role" character varying NOT NULL DEFAULT 'member',
  "parish_id" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "parish_members_parishes_members" FOREIGN KEY ("parish_id") REFERENCES "parishes" ("id") ON DELETE NO ACTION
);
-- Create index "parishmember_parish_id_user_id" to table: "parish_members"
CREATE UNIQUE INDEX IF NOT EXISTS "parishmember_parish_id_user_id" ON "parish_members" ("parish_id", "user_id");
-- Create index "parishmember_user_id" to table: "parish_members"
CREATE INDEX IF NOT EXISTS "parishmember_user_id" ON "parish_members" ("user_id");

-- Create "zones" table
CREATE TABLE IF NOT EXISTS "zones" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  "type" character varying NOT NULL DEFAULT 'zone',
  "code" character varying NOT NULL,
  "name" character varying NOT NULL,
  "coordinator_name" character varying NULL,
  "coordinator_phone" character varying NULL,
  "active" boolean NOT NULL DEFAULT true,
  "parish_id" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "zones_parishes_zones" FOREIGN KEY ("parish_id") REFERENCES "parishes" ("id") ON DELETE NO ACTION
);
-- Create index "zone_parish_id_code" to table: "zones"
CREATE UNIQUE INDEX IF NOT EXISTS "zone_parish_id_code" ON "zones" ("parish_id", "code");
-- Create index "zone_parish_id_type_active" to table: "zones"
CREATE INDEX IF NOT EXISTS "zone_parish_id_type_active" ON "zones" ("parish_id", "type", "active");

-- Create "wards" table
CREATE TABLE IF NOT EXISTS "wards" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  "code" character varying NOT NULL,
  "name" character varying NOT NULL,
  "patron_saint" character varying NULL,
  "head_name" character varying NULL,
  "head_phone" character varying NULL,
  "active" boolean NOT NULL DEFAULT true,
  "parish_id" uuid NOT NULL,
  "zone_id" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "wards_parishes_wards" FOREIGN KEY ("parish_id") REFERENCES "parishes" ("id") ON DELETE NO ACTION,
  CONSTRAINT "wards_zones_wards" FOREIGN KEY ("zone_id") REFERENCES "zones" ("id") ON DELETE NO ACTION
);
-- Create index "ward_parish_id_code" to table: "wards"
CREATE UNIQUE INDEX IF NOT EXISTS "ward_parish_id_code" ON "wards" ("parish_id", "code");
-- Create index "ward_parish_id_zone_id_active" to table: "wards"
CREATE INDEX IF NOT EXISTS "ward_parish_id_zone_id_active" ON "wards" ("parish_id", "zone_id", "active");

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
  "parish_id" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "lookups_parishes_lookups" FOREIGN KEY ("parish_id") REFERENCES "parishes" ("id") ON DELETE SET NULL
);
-- Create index "lookup_parish_id_category_code" to table: "lookups"
CREATE UNIQUE INDEX IF NOT EXISTS "lookup_parish_id_category_code" ON "lookups" ("parish_id", "category", "code");
-- Create index "lookup_parish_id_category_sort_order" to table: "lookups"
CREATE INDEX IF NOT EXISTS "lookup_parish_id_category_sort_order" ON "lookups" ("parish_id", "category", "sort_order");

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
  "parish_id" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "periods_parishes_periods" FOREIGN KEY ("parish_id") REFERENCES "parishes" ("id") ON DELETE NO ACTION
);
-- Create index "period_parish_id_code" to table: "periods"
CREATE UNIQUE INDEX IF NOT EXISTS "period_parish_id_code" ON "periods" ("parish_id", "code");
-- Create index "period_parish_id_type_is_current" to table: "periods"
CREATE INDEX IF NOT EXISTS "period_parish_id_type_is_current" ON "periods" ("parish_id", "type", "is_current");
