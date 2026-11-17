import i18n from "~/shared/i18n";
import { UnderConstruction } from "~/shared/ui/UnderConstruction";

import { app } from "../app";
import type { Route } from "./+types/EventPage";

export function meta({}: Route.MetaArgs) {
  return [{ title: `${i18n.t("apps.event.name")} · Domus` }];
}

export default function EventPage() {
  return <UnderConstruction app={app} />;
}
