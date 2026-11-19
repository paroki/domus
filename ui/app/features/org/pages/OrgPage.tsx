import i18n from "~/shared/i18n";
import { UnderConstruction } from "~/shared/ui/UnderConstruction";

import { app } from "../app";
import type { Route } from "./+types/OrgPage";

export function meta({}: Route.MetaArgs) {
  return [{ title: `${i18n.t("apps.org.name")} · Domus` }];
}

export default function OrgPage() {
  return <UnderConstruction app={app} />;
}
