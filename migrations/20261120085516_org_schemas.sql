-- Create index "dioceses_code_key" to table: "dioceses"
CREATE UNIQUE INDEX "dioceses_code_key" ON "dioceses" ("code");
-- Create index "parishes_code_key" to table: "parishes"
CREATE UNIQUE INDEX "parishes_code_key" ON "parishes" ("code");
-- Create index "user_sessions_token_hash_key" to table: "user_sessions"
CREATE UNIQUE INDEX "user_sessions_token_hash_key" ON "user_sessions" ("token_hash");
-- Create index "users_email_key" to table: "users"
CREATE UNIQUE INDEX "users_email_key" ON "users" ("email");
-- Create "organizations" table
CREATE TABLE "organizations" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  "parish_id" uuid NOT NULL,
  "zone_id" uuid NULL,
  "ward_id" uuid NULL,
  "kind" character varying NOT NULL DEFAULT 'section',
  "code" character varying NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "active" boolean NOT NULL DEFAULT true,
  "parent_id" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "organizations_organizations_children" FOREIGN KEY ("parent_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "organization_parish_id_code" to table: "organizations"
CREATE UNIQUE INDEX "organization_parish_id_code" ON "organizations" ("parish_id", "code");
-- Create index "organization_parish_id_kind_active" to table: "organizations"
CREATE INDEX "organization_parish_id_kind_active" ON "organizations" ("parish_id", "kind", "active");
-- Create index "organization_parish_id_parent_id" to table: "organizations"
CREATE INDEX "organization_parish_id_parent_id" ON "organizations" ("parish_id", "parent_id");
-- Create index "organization_parish_id_ward_id" to table: "organizations"
CREATE INDEX "organization_parish_id_ward_id" ON "organizations" ("parish_id", "ward_id");
-- Create index "organization_parish_id_zone_id" to table: "organizations"
CREATE INDEX "organization_parish_id_zone_id" ON "organizations" ("parish_id", "zone_id");
-- Create "org_invites" table
CREATE TABLE "org_invites" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "parish_id" uuid NOT NULL,
  "code" character varying NOT NULL,
  "role" character varying NOT NULL DEFAULT 'member',
  "default_position" character varying NULL,
  "max_uses" bigint NULL,
  "used_count" bigint NOT NULL DEFAULT 0,
  "expires_at" timestamptz NOT NULL,
  "revoked_at" timestamptz NULL,
  "created_by" uuid NOT NULL,
  "org_id" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "org_invites_organizations_invites" FOREIGN KEY ("org_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "org_invites_code_key" to table: "org_invites"
CREATE UNIQUE INDEX "org_invites_code_key" ON "org_invites" ("code");
-- Create index "orginvite_code" to table: "org_invites"
CREATE UNIQUE INDEX "orginvite_code" ON "org_invites" ("code");
-- Create index "orginvite_org_id_created_at" to table: "org_invites"
CREATE INDEX "orginvite_org_id_created_at" ON "org_invites" ("org_id", "created_at");
-- Create index "orginvite_parish_id_org_id" to table: "org_invites"
CREATE INDEX "orginvite_parish_id_org_id" ON "org_invites" ("parish_id", "org_id");
-- Create "org_members" table
CREATE TABLE "org_members" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "parish_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  "role" character varying NOT NULL DEFAULT 'member',
  "position" character varying NULL,
  "period_id" uuid NULL,
  "start_date" character varying NULL,
  "end_date" character varying NULL,
  "status" character varying NOT NULL DEFAULT 'active',
  "org_id" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "org_members_organizations_members" FOREIGN KEY ("org_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "orgmember_org_id_role_status" to table: "org_members"
CREATE INDEX "orgmember_org_id_role_status" ON "org_members" ("org_id", "role", "status");
-- Create index "orgmember_org_id_user_id" to table: "org_members"
CREATE UNIQUE INDEX "orgmember_org_id_user_id" ON "org_members" ("org_id", "user_id");
-- Create index "orgmember_parish_id_user_id" to table: "org_members"
CREATE INDEX "orgmember_parish_id_user_id" ON "org_members" ("parish_id", "user_id");
