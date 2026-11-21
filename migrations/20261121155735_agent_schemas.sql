-- Enable pgvector and ParadeDB BM25 extensions
CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS pg_search CASCADE;
-- Create "agent_sessions" table
CREATE TABLE IF NOT EXISTS "agent_sessions" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  "parish_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  "channel" character varying NOT NULL DEFAULT 'web',
  "title" character varying NOT NULL DEFAULT 'Percakapan Baru',
  "last_message_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "agentsession_parish_id_user_id_last_message_at" to table: "agent_sessions"
CREATE INDEX IF NOT EXISTS "agentsession_parish_id_user_id_last_message_at" ON "agent_sessions" ("parish_id", "user_id", "last_message_at");
-- Create index "agentsession_parish_id_user_id_channel_last_message_at" to table: "agent_sessions"
CREATE INDEX IF NOT EXISTS "agentsession_parish_id_user_id_channel_last_message_at" ON "agent_sessions" ("parish_id", "user_id", "channel", "last_message_at");
-- Create "agent_messages" table
CREATE TABLE IF NOT EXISTS "agent_messages" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "parish_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  "agent_name" character varying NOT NULL DEFAULT 'gabriel',
  "role" character varying NOT NULL,
  "content" text NOT NULL DEFAULT '',
  "thinking" text NULL,
  "tool_calls" jsonb NULL,
  "tool_call_id" character varying NULL,
  "tool_name" character varying NULL,
  "session_id" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "agent_messages_agent_sessions_messages" FOREIGN KEY ("session_id") REFERENCES "agent_sessions" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "agentmessage_session_id_created_at" to table: "agent_messages"
CREATE INDEX IF NOT EXISTS "agentmessage_session_id_created_at" ON "agent_messages" ("session_id", "created_at");
-- Create index "agentmessage_parish_id_user_id_created_at" to table: "agent_messages"
CREATE INDEX IF NOT EXISTS "agentmessage_parish_id_user_id_created_at" ON "agent_messages" ("parish_id", "user_id", "created_at");
-- Create "agent_memories" table
CREATE TABLE IF NOT EXISTS "agent_memories" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  "parish_id" uuid NOT NULL,
  "user_id" uuid NULL,
  "scope" character varying NOT NULL DEFAULT 'user',
  "category" character varying NOT NULL DEFAULT 'preference',
  "content" text NOT NULL,
  "is_pinned" boolean NOT NULL DEFAULT false,
  "source_session_id" uuid NULL,
  PRIMARY KEY ("id")
);
-- Create index "agentmemory_parish_id_user_id_is_pinned" to table: "agent_memories"
CREATE INDEX IF NOT EXISTS "agentmemory_parish_id_user_id_is_pinned" ON "agent_memories" ("parish_id", "user_id", "is_pinned");
-- Create index "agentmemory_parish_id_scope_category" to table: "agent_memories"
CREATE INDEX IF NOT EXISTS "agentmemory_parish_id_scope_category" ON "agent_memories" ("parish_id", "scope", "category");
-- Half-precision vector column for semantic memory search (kept outside Ent schema)
ALTER TABLE "agent_memories" ADD COLUMN IF NOT EXISTS "embedding" halfvec(768) NULL;
-- ParadeDB BM25 indexes for full-text search
CREATE INDEX IF NOT EXISTS "agent_messages_bm25_idx" ON "agent_messages" USING bm25 ("id", "content", "parish_id", "user_id", "session_id", "role") WITH (key_field = 'id');
CREATE INDEX IF NOT EXISTS "agent_memories_bm25_idx" ON "agent_memories" USING bm25 ("id", "content", "category", "scope", "parish_id", "user_id") WITH (key_field = 'id');
-- HNSW index for halfvec cosine similarity on active memories
CREATE INDEX IF NOT EXISTS "agent_memories_embedding_hnsw_idx" ON "agent_memories" USING hnsw ("embedding" halfvec_cosine_ops) WITH (m = 16, ef_construction = 64) WHERE "deleted_at" IS NULL AND "embedding" IS NOT NULL;
