import { Users } from "lucide-react";

import type { AppDef } from "~/shared/app-def";

export const app: AppDef = {
  id: "umat",
  path: "/umat",
  order: 3,
  plate: 3,
  icon: Users,
  roles: ["umat:admin"],
};
