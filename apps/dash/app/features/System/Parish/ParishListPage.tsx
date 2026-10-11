import { useTranslation } from "react-i18next";
import i18n from "~/i18n";
import { ControlPanel } from "~/shared/layout/ControlPanel";
import { UnderConstruction } from "~/shared/layout/UnderConstruction";
import { modulePath } from "~/shared/modules/registry";
import type { Route } from "./+types/ParishListPage";

export function meta(_: Route.MetaArgs) {
  return [
    {
      title: `${i18n.t("modules.sys.menu.parish")} | ${i18n.t("modules.sys.name")} | Domus`,
    },
  ];
}

/** Daftar paroki. Ganti `UnderConstruction` dengan tabel + aksi CRUD. */
export default function ParishListPage() {
  const { t } = useTranslation();

  return (
    <>
      <ControlPanel
        breadcrumbs={[
          { title: t("modules.sys.name"), to: modulePath("sys") },
          { title: t("modules.sys.menu.parish") },
        ]}
      />
      <section className="mx-auto w-full max-w-5xl px-4 py-5">
        <UnderConstruction
          title={t("modules.sys.menu.parish")}
          backTo={modulePath("sys")}
        />
      </section>
    </>
  );
}
