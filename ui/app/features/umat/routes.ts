import { type RouteConfig, route } from "@react-router/dev/routes";

export default [
  route("umat", "./features/umat/pages/UmatDirectoryPage.tsx"),
  route("umat/new", "./features/umat/pages/UmatCreateWizardPage.tsx", {
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
  route("umat/referensi", "./features/umat/pages/UmatReferensiPage.tsx", {
    id: "umat-referensi",
  }),
  route(
    "umat/referensi/new",
    "./features/umat/pages/UmatReferensiFormPage.tsx",
    { id: "umat-referensi-new" },
  ),
  route(
    "umat/referensi/:id",
    "./features/umat/pages/UmatReferensiFormPage.tsx",
    { id: "umat-referensi-edit" },
  ),
  route("umat/:umatId", "./features/umat/pages/UmatDetailPage.tsx", {
    id: "umat-detail",
  }),
  route(
    "umat/:umatId/sakramen/new",
    "./features/umat/pages/UmatSakramenFormPage.tsx",
    { id: "umat-sakramen-new" },
  ),
  route(
    "umat/:umatId/sakramen/:sakramenId",
    "./features/umat/pages/UmatSakramenFormPage.tsx",
    { id: "umat-sakramen-edit" },
  ),
] satisfies RouteConfig;
