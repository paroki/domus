import { data, Outlet } from "react-router";
import { getModule } from "~/shared/modules/registry";
import type { Route } from "../../routes/+types/_app.$moduleId";

/** Modul yang tidak ada di registry berujung ke 404 (ErrorBoundary di root). */
export function clientLoader({ params }: Route.ClientLoaderArgs) {
  if (!getModule(params.moduleId)) throw data("Not Found", { status: 404 });
  return null;
}

export default function ModuleRoute() {
  return <Outlet />;
}
