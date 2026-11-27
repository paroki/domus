import assert from "node:assert/strict";
import test from "node:test";
import { Church } from "lucide-react";
import { renderToString } from "react-dom/server";
import { StaticRouter } from "react-router";

import "~/shared/i18n";
import { IosInstallDialog } from "~/shared/pwa/IosInstallDialog";
import { PwaInstallBanner } from "~/shared/pwa/PwaInstallBanner";
import {
  type BeforeInstallPromptEvent,
  setDeferredInstallPrompt,
} from "~/shared/pwa/usePwaInstall";
import { Button } from "./Button";
import { ConfirmDialog } from "./ConfirmDialog";
import {
  InductiveDangerZone,
  InductiveEmptyState,
  InductiveFormSection,
  InductivePageHeader,
  InductiveStepCard,
} from "./iui";

test("IUI primitives render mobile-first responsive density and full-width CTAs", () => {
  const html = renderToString(
    <StaticRouter location="/org">
      <InductivePageHeader
        breadcrumbs={[{ label: "Beranda", to: "/" }, { label: "Organisasi" }]}
        title="Organisasi"
        subtitle="Kelola paroki"
        action={<Button type="button">Tambah</Button>}
      />
      <InductiveStepCard
        to="/org/wilayah"
        title="Wilayah"
        description="Daftar wilayah"
        icon={Church}
      />
      <InductiveFormSection title="Formulir">
        <p>Konten</p>
      </InductiveFormSection>
      <InductiveDangerZone
        title="Zona Bahaya"
        description="Hapus permanen"
        action={<Button type="button">Hapus</Button>}
      />
      <InductiveEmptyState
        icon={Church}
        title="Kosong"
        description="Belum ada data"
        action={<Button type="button">Buat</Button>}
      />
    </StaticRouter>,
  );

  assert.ok(html.includes("w-full"));
  assert.ok(html.includes("sm:w-auto"));
  assert.ok(html.includes("p-4"));
  assert.ok(html.includes("sm:p-5"));
  assert.ok(html.includes("sm:p-6"));
  assert.ok(html.includes("p-6"));
  assert.ok(html.includes("sm:p-10"));
});

test("ConfirmDialog renders mobile-first card padding and stacked full-width buttons", () => {
  const html = renderToString(
    <ConfirmDialog
      open
      title="Hapus data?"
      description="Tindakan ini tidak dapat dibatalkan."
      onConfirm={() => {}}
      onCancel={() => {}}
    />,
  );

  assert.ok(html.includes("p-4"));
  assert.ok(html.includes("sm:p-6"));
  assert.ok(html.includes("flex-col-reverse"));
  assert.ok(html.includes("sm:flex-row"));
  assert.ok(html.includes("w-full sm:w-auto"));
});

test("PwaInstallBanner renders safe-area bottom offset and 44px dismiss target", () => {
  setDeferredInstallPrompt({
    prompt: async () => {},
    userChoice: Promise.resolve({ outcome: "dismissed" }),
  } as unknown as BeforeInstallPromptEvent);

  try {
    const html = renderToString(<PwaInstallBanner />);
    assert.ok(html.includes("bottom-[max(1rem,env(safe-area-inset-bottom))]"));
    assert.ok(html.includes("min-w-0 flex-1"));
    assert.ok(html.includes("size-11"));
  } finally {
    setDeferredInstallPrompt(null);
  }
});

test("IosInstallDialog renders mobile-first padding, 44px close target, and full-width CTA", () => {
  const html = renderToString(<IosInstallDialog open onClose={() => {}} />);

  assert.ok(html.includes("p-4"));
  assert.ok(html.includes("sm:p-6"));
  assert.ok(html.includes("size-11"));
  assert.ok(html.includes("w-full sm:w-auto"));
});
