import { data } from "react-router";
import { ControlPanel } from "~/shared/layout/ControlPanel";
import { UnderConstruction } from "~/shared/layout/UnderConstruction";
import { findMenuItem, getModule, modulePath } from "~/shared/modules/registry";
import { useActiveModule } from "~/shared/modules/useActiveModule";
import type { Route } from "../../routes/+types/_app.$moduleId.$page";

export function meta({ params }: Route.MetaArgs) {
  const mod = getModule(params.moduleId);
  const item = mod && findMenuItem(mod, params.page);
  return [
    { title: mod && item ? `${item.label} | ${mod.name} | Domus` : "Domus" },
  ];
}

/** Halaman yang tidak ada di menu modul berujung ke 404. */
export function clientLoader({ params }: Route.ClientLoaderArgs) {
  const mod = getModule(params.moduleId);
  if (!mod || !findMenuItem(mod, params.page)) {
    throw data("Not Found", { status: 404 });
  }
  return null;
}

/**
 * Placeholder untuk semua halaman modul yang belum dibuat.
 * Ganti dengan route statis `_app.<module>.<slug>.tsx` saat halamannya siap.
 */
export default function ModulePage() {
  const { activeModule, activeItem } = useActiveModule();
  if (!activeModule || !activeItem) return null;

  return (
    <>
      <ControlPanel
        breadcrumbs={[
          { title: activeModule.name, to: modulePath(activeModule.id) },
          ...(activeItem.group ? [{ title: activeItem.group }] : []),
          { title: activeItem.label },
        ]}
      />
      <section className="mx-auto w-full max-w-5xl px-4 py-5">
        <UnderConstruction
          title={activeItem.label}
          backTo={modulePath(activeModule.id)}
        />
      </section>
    </>
  );
}
