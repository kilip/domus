import type { RouteConfig, RouteConfigEntry } from "@react-router/dev/routes";
import { flatRoutes } from "@react-router/fs-routes";
import { systemRoutes } from "./features/System/routes";

/**
 * Route yang disumbang tiap modul. Tambah modul baru = bikin
 * `features/<Modul>/routes.ts`, lalu daftarkan di sini.
 * Semuanya masuk ke dalam layout `_app` (guard login + AppShell).
 */
const moduleRoutes: RouteConfigEntry[] = [...systemRoutes];

const APP_LAYOUT_FILE = "routes/_app.tsx";

/**
 * `flatRoutes()` tetap memegang route berbasis file di `app/routes/`
 * (login, launcher, placeholder `$moduleId/$page`). Route modul disuntik
 * sebagai anak layout `_app` supaya shell tidak di-mount ulang.
 * Route statis modul menang atas route dinamis `$moduleId/$page`.
 */
async function buildRoutes(): Promise<RouteConfigEntry[]> {
  const routes = await flatRoutes();
  const appLayout = routes.find((entry) => entry.file === APP_LAYOUT_FILE);
  if (!appLayout) {
    throw new Error(`Layout route "${APP_LAYOUT_FILE}" tidak ditemukan.`);
  }
  appLayout.children = [...(appLayout.children ?? []), ...moduleRoutes];
  return routes;
}

export default buildRoutes() satisfies Promise<RouteConfig>;
