-- Reconcile agent_sessions, agent_messages, and agent_memories with current Ent schemas and hybrid RAG indexes.
DROP INDEX IF EXISTS "agentsession_parish_id_user_id_channel_last_message_at";
ALTER TABLE "agent_sessions"
  DROP COLUMN IF EXISTS "channel",
  ADD COLUMN IF NOT EXISTS "provider_id" uuid NULL,
  ADD COLUMN IF NOT EXISTS "model_id" character varying NULL;

DROP INDEX IF EXISTS "agent_messages_bm25_idx";
ALTER TABLE "agent_messages"
  DROP COLUMN IF EXISTS "thinking",
  ADD COLUMN IF NOT EXISTS "meta" jsonb NULL;

DROP INDEX IF EXISTS "agent_memories_bm25_idx";
DROP INDEX IF EXISTS "agent_memories_embedding_hnsw_idx";
DROP INDEX IF EXISTS "agentmemory_parish_id_user_id_is_pinned";
DROP INDEX IF EXISTS "agentmemory_parish_id_scope_category";
DELETE FROM "agent_memories" WHERE "user_id" IS NULL;
UPDATE "agent_memories" SET "category" = 'fact' WHERE "category" NOT IN ('preference', 'fact', 'context');
ALTER TABLE "agent_memories"
  DROP COLUMN IF EXISTS "deleted_at",
  DROP COLUMN IF EXISTS "scope",
  DROP COLUMN IF EXISTS "is_pinned",
  ALTER COLUMN "user_id" SET NOT NULL,
  ALTER COLUMN "category" SET DEFAULT 'fact';
CREATE INDEX IF NOT EXISTS "agentmemory_parish_id_user_id_created_at" ON "agent_memories" ("parish_id", "user_id", "created_at");
ALTER TABLE "agent_memories" ADD COLUMN IF NOT EXISTS "embedding" halfvec(768) NULL;
CREATE INDEX IF NOT EXISTS "agent_memories_bm25_idx" ON "agent_memories" USING bm25 ("id", "content", "category", "parish_id", "user_id") WITH (key_field = 'id');
CREATE INDEX IF NOT EXISTS "agent_memories_embedding_hnsw_idx" ON "agent_memories" USING hnsw ("embedding" halfvec_cosine_ops) WITH (m = 16, ef_construction = 64) WHERE "embedding" IS NOT NULL;
