import i18n from "~/shared/i18n";
import { UnderConstruction } from "~/shared/ui/UnderConstruction";

import { app } from "../app";
import type { Route } from "./+types/UsersPage";

export function meta(_: Route.MetaArgs) {
  return [{ title: `${i18n.t("apps.users.name")} · Domus` }];
}

export default function UsersPage() {
  return <UnderConstruction app={app} />;
}
