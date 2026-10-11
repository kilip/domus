import { useLocation, useParams } from "react-router";
import { findMenuItem, getModule } from "./registry";

/** Modul dan halaman yang sedang aktif, dibaca dari URL (`/:moduleId/:page`). */
export function useActiveModule() {
  const params = useParams();
  const location = useLocation();
  const pathSegments = location.pathname.split("/").filter(Boolean);
  const moduleId = params.moduleId ?? pathSegments[0];
  const page = params.page ?? pathSegments[1];

  const activeModule = moduleId ? getModule(moduleId) : undefined;
  const activeItem =
    activeModule && page ? findMenuItem(activeModule, page) : undefined;

  return { activeModule, activeItem };
}
