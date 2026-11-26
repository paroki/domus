import { type RouteConfig, route } from "@react-router/dev/routes";

export default [
  route("umat", "./features/umat/pages/UmatDirectoryPage.tsx"),
  route("umat/new", "./features/umat/pages/UmatPlaceholderPage.tsx", {
    id: "umat-new",
  }),
  route("umat/keluarga", "./features/umat/pages/KeluargaListPage.tsx"),
  route("umat/keluarga/new", "./features/umat/pages/KeluargaCreatePage.tsx", {
    id: "umat-keluarga-new",
  }),
  route(
    "umat/keluarga/:keluargaId",
    "./features/umat/pages/KeluargaDetailPage.tsx",
    { id: "umat-keluarga-detail" },
  ),
  route("umat/:umatId", "./features/umat/pages/UmatPlaceholderPage.tsx", {
    id: "umat-detail",
  }),
] satisfies RouteConfig;
