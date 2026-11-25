import { BookOpen } from "lucide-react";

import type { AppDef } from "~/shared/app-def";

export const app: AppDef = {
  id: "ref",
  path: "/ref",
  order: 4,
  plate: 4,
  icon: BookOpen,
  roles: ["user:super-admin", "ref:paroki-admin"],
};
