-- Create "events" table
CREATE TABLE IF NOT EXISTS "events" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  "parish_id" uuid NOT NULL,
  "org_id" uuid NOT NULL,
  "created_by" uuid NOT NULL,
  "title" character varying NOT NULL,
  "description" character varying NULL,
  "location" character varying NULL,
  "start_at" timestamptz NOT NULL,
  "end_at" timestamptz NOT NULL,
  "visibility" character varying NOT NULL DEFAULT 'public',
  "rsvp_enabled" boolean NOT NULL DEFAULT true,
  "rsvp_deadline" timestamptz NULL,
  "status" character varying NOT NULL DEFAULT 'scheduled',
  PRIMARY KEY ("id")
);
-- Create index "event_parish_id_start_at" to table: "events"
CREATE INDEX IF NOT EXISTS "event_parish_id_start_at" ON "events" ("parish_id", "start_at");
-- Create index "event_parish_id_org_id_start_at" to table: "events"
CREATE INDEX IF NOT EXISTS "event_parish_id_org_id_start_at" ON "events" ("parish_id", "org_id", "start_at");
-- Create index "event_parish_id_visibility_status_start_at" to table: "events"
CREATE INDEX IF NOT EXISTS "event_parish_id_visibility_status_start_at" ON "events" ("parish_id", "visibility", "status", "start_at");
-- Create index "event_org_id" to table: "events"
CREATE INDEX IF NOT EXISTS "event_org_id" ON "events" ("org_id");
-- Create "event_rsvps" table
CREATE TABLE IF NOT EXISTS "event_rsvps" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "parish_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  "status" character varying NOT NULL,
  "note" character varying NULL,
  "event_id" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "event_rsvps_events_rsvps" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "eventrsvp_event_id_user_id" to table: "event_rsvps"
CREATE UNIQUE INDEX IF NOT EXISTS "eventrsvp_event_id_user_id" ON "event_rsvps" ("event_id", "user_id");
-- Create index "eventrsvp_event_id_status_updated_at" to table: "event_rsvps"
CREATE INDEX IF NOT EXISTS "eventrsvp_event_id_status_updated_at" ON "event_rsvps" ("event_id", "status", "updated_at");
-- Create index "eventrsvp_parish_id_user_id" to table: "event_rsvps"
CREATE INDEX IF NOT EXISTS "eventrsvp_parish_id_user_id" ON "event_rsvps" ("parish_id", "user_id");
