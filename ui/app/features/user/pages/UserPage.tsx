import i18n from "~/shared/i18n";
import { UnderConstruction } from "~/shared/ui/UnderConstruction";

import { app } from "../app";
import type { Route } from "./+types/UserPage";

export function meta({}: Route.MetaArgs) {
  return [{ title: `${i18n.t("apps.user.name")} · Domus` }];
}

export default function UserPage() {
  return <UnderConstruction app={app} />;
}
