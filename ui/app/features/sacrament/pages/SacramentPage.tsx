import i18n from "~/shared/i18n";
import { UnderConstruction } from "~/shared/ui/UnderConstruction";

import { app } from "../app";
import type { Route } from "./+types/SacramentPage";

export function meta({}: Route.MetaArgs) {
  return [{ title: `${i18n.t("apps.sacrament.name")} · Domus` }];
}

export default function SacramentPage() {
  return <UnderConstruction app={app} />;
}
