-- Create "agent_providers" table
CREATE TABLE "agent_providers" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NOT NULL DEFAULT '',
  "provider" character varying NOT NULL,
  "api_key" character varying NOT NULL,
  "base_url" character varying NOT NULL DEFAULT '',
  "models" jsonb NOT NULL,
  PRIMARY KEY ("id")
);
