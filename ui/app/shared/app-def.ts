import type { LucideIcon } from "lucide-react";

/** Entry a feature registers in the launcher (see ~/apps and features/<f>/app.ts). */
export type AppDef = {
  /** Feature id; also the i18n key under `apps.<id>`. */
  id: string;
  /** Route path, e.g. "/org". */
  path: string;
  /** Launcher sort order, ascending. */
  order: number;
  /** Icon plate color token: plate-1 … plate-6. */
  plate: 1 | 2 | 3 | 4 | 5 | 6;
  icon: LucideIcon;
  /** Optional required global roles; if set, caller must hold at least one. */
  roles?: string[];
};
