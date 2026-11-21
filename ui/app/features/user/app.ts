import { UserCog } from "lucide-react";

import type { AppDef } from "~/shared/app-def";

export const app: AppDef = {
  id: "user",
  path: "/users",
  order: 5,
  plate: 6,
  icon: UserCog,
  roles: ["user:super-admin", "ref:parish-admin", "org:admin"],
};
