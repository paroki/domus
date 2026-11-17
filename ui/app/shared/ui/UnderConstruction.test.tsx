import assert from "node:assert/strict";
import test from "node:test";
import { Church } from "lucide-react";
import { renderToString } from "react-dom/server";
import { StaticRouter } from "react-router";

import "~/shared/i18n";
import type { AppDef } from "~/shared/app-def";
import { UnderConstruction } from "./UnderConstruction";

const app: AppDef = {
  id: "organization",
  path: "/org",
  order: 1,
  plate: 1,
  icon: Church,
};

test("UnderConstruction renders the localized title, description and a link home", () => {
  const html = renderToString(
    <StaticRouter location="/org">
      <UnderConstruction app={app} />
    </StaticRouter>,
  );

  assert.ok(html.includes("Halaman Organisasi sedang dibangun"));
  assert.ok(html.includes("Struktur organisasi, wilayah, dan lingkungan paroki."));
  assert.ok(html.includes('href="/"'));
  assert.ok(html.includes("bg-plate-1"));
});
