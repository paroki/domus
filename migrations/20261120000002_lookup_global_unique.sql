-- Create index "lookup_global_category_code" to table: "lookups"
CREATE UNIQUE INDEX IF NOT EXISTS "lookup_global_category_code" ON "lookups" ("category", "code") WHERE "parish_id" IS NULL;
