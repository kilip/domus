import { prefix, type RouteConfigEntry, route } from "@react-router/dev/routes";

/**
 * Route modul System (`/sys/*`).
 *
 * File ini harus murni: hanya boleh import dari `@react-router/dev/routes`
 * (jalan di Node saat build, jadi jangan import React atau komponen).
 * Path file relatif ke folder `app/`, tanpa `~/` dan tanpa `./`.
 */
export const systemRoutes: RouteConfigEntry[] = prefix("sys", [
  route("diocese", "features/System/Diocese/DioceseListPage.tsx"),
  route("diocese/create", "features/System/Diocese/DioceseCreatePage.tsx"),
  route("diocese/update/:id", "features/System/Diocese/DioceseUpdatePage.tsx"),
  route("parish", "features/System/Parish/ParishListPage.tsx"),
  route("parish/create", "features/System/Parish/ParishCreatePage.tsx"),
  route("parish/update/:id", "features/System/Parish/ParishUpdatePage.tsx"),
  route("territorial", "features/System/Territorial/TerritorialListPage.tsx"),
  route(
    "territorial/create",
    "features/System/Territorial/TerritorialCreatePage.tsx",
  ),
  route(
    "territorial/update/:id",
    "features/System/Territorial/TerritorialUpdatePage.tsx",
  ),
  route("unit", "features/System/Unit/UnitListPage.tsx"),
  route("unit/create", "features/System/Unit/UnitCreatePage.tsx"),
  route("unit/update/:id", "features/System/Unit/UnitUpdatePage.tsx"),
]);
