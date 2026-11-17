import i18n from "~/shared/i18n";
import { UnderConstruction } from "~/shared/ui/UnderConstruction";

import { app } from "../app";
import type { Route } from "./+types/SettingsPage";

export function meta(_: Route.MetaArgs) {
  return [{ title: `${i18n.t("apps.settings.name")} · Domus` }];
}

export default function SettingsPage() {
  return <UnderConstruction app={app} />;
}
