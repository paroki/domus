import i18n from "~/shared/i18n";
import { UnderConstruction } from "~/shared/ui/UnderConstruction";

import { app } from "../app";
import type { Route } from "./+types/OrganizationPage";

export function meta({}: Route.MetaArgs) {
  return [{ title: `${i18n.t("apps.organization.name")} · Domus` }];
}

export default function OrganizationPage() {
  return <UnderConstruction app={app} />;
}
