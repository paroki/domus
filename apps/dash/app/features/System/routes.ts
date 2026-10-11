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
  route("parish", "features/System/Parish/ParishListPage.tsx"),
  route("territorial", "features/System/Territorial/TerritorialListPage.tsx"),
]);
