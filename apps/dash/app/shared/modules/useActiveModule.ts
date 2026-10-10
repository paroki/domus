import { useParams } from "react-router";
import { findMenuItem, getModule } from "./registry";

/** Modul dan halaman yang sedang aktif, dibaca dari URL (`/:moduleId/:page`). */
export function useActiveModule() {
  const { moduleId, page } = useParams();
  const activeModule = moduleId ? getModule(moduleId) : undefined;
  const activeItem =
    activeModule && page ? findMenuItem(activeModule, page) : undefined;

  return { activeModule, activeItem };
}
