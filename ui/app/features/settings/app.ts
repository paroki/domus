import { Settings } from "lucide-react";

import type { AppDef } from "~/shared/app-def";

export const app: AppDef = {
  id: "settings",
  path: "/settings",
  order: 5,
  plate: 6,
  icon: Settings,
  roles: ["user:super-admin"],
};
