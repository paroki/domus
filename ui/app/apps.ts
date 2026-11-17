import type { AppDef } from "~/shared/app-def";

// Launcher registry: every features/<feature>/app.ts exporting `app` is
// included automatically, sorted by `order`.
const modules = import.meta.glob<{ app: AppDef }>("./features/*/app.ts", {
  eager: true,
});

export const apps: AppDef[] = Object.values(modules)
  .map((m) => m.app)
  .sort((a, b) => a.order - b.order);
