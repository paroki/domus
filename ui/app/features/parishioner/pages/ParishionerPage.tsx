import i18n from "~/shared/i18n";
import { UnderConstruction } from "~/shared/ui/UnderConstruction";

import { app } from "../app";
import type { Route } from "./+types/ParishionerPage";

export function meta({}: Route.MetaArgs) {
  return [{ title: `${i18n.t("apps.parishioner.name")} · Domus` }];
}

export default function ParishionerPage() {
  return <UnderConstruction app={app} />;
}
