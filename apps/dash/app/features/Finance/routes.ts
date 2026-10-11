import {
  index,
  layout,
  prefix,
  type RouteConfigEntry,
  route,
} from "@react-router/dev/routes";

/**
 * Route modul Finance (`/fin/*`).
 *
 * File ini harus murni: hanya boleh import dari `@react-router/dev/routes`
 * (jalan di Node saat build, jadi jangan import React atau komponen).
 * Path file relatif ke folder `app/`, tanpa `~/` dan tanpa `./`.
 */
export const financeRoutes: RouteConfigEntry[] = prefix("fin", [
  layout("features/Finance/FinanceLayout.tsx", [
    index("features/Finance/FinanceHomePage.tsx"),
    route(":page", "features/Finance/FinancePlaceholderPage.tsx"),
  ]),
]);
