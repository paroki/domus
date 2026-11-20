import { Settings } from "lucide-react";

import type { AppDef } from "~/shared/app-def";

export const app: AppDef = {
  id: "ref",
  path: "/settings",
  order: 4,
  plate: 4,
  icon: Settings,
  roles: ["user:super-admin", "ref:parish-admin"],
};
