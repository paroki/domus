import { type RouteConfig, route } from "@react-router/dev/routes";

export default [
  route("settings/gerejawi", "./features/ref/pages/GerejawiHubPage.tsx"),
  route(
    "settings/gerejawi/keuskupan",
    "./features/ref/pages/KeuskupanListPage.tsx",
  ),
  route(
    "settings/gerejawi/keuskupan/new",
    "./features/ref/pages/KeuskupanFormPage.tsx",
    { id: "ref-keuskupan-new" },
  ),
  route(
    "settings/gerejawi/keuskupan/:id",
    "./features/ref/pages/KeuskupanFormPage.tsx",
    { id: "ref-keuskupan-edit" },
  ),
  route(
    "settings/gerejawi/kevikepan",
    "./features/ref/pages/KevikepanListPage.tsx",
  ),
  route(
    "settings/gerejawi/kevikepan/new",
    "./features/ref/pages/KevikepanFormPage.tsx",
    { id: "ref-kevikepan-new" },
  ),
  route(
    "settings/gerejawi/kevikepan/:id",
    "./features/ref/pages/KevikepanFormPage.tsx",
    { id: "ref-kevikepan-edit" },
  ),
  route("settings/gerejawi/paroki", "./features/ref/pages/ParokiListPage.tsx"),
  route(
    "settings/gerejawi/paroki/new",
    "./features/ref/pages/ParokiCreateWizardPage.tsx",
  ),
  route(
    "settings/gerejawi/paroki/:id",
    "./features/ref/pages/ParokiDetailPage.tsx",
  ),
  route("settings/sipil", "./features/ref/pages/SipilHubPage.tsx"),
  route("settings/sipil/kota", "./features/ref/pages/KotaListPage.tsx"),
  route("settings/sipil/kota/new", "./features/ref/pages/KotaFormPage.tsx", {
    id: "ref-kota-new",
  }),
  route("settings/sipil/kota/:id", "./features/ref/pages/KotaFormPage.tsx", {
    id: "ref-kota-edit",
  }),
  route(
    "settings/sipil/kecamatan",
    "./features/ref/pages/KecamatanListPage.tsx",
  ),
  route(
    "settings/sipil/kecamatan/new",
    "./features/ref/pages/KecamatanFormPage.tsx",
    { id: "ref-kecamatan-new" },
  ),
  route(
    "settings/sipil/kecamatan/:id",
    "./features/ref/pages/KecamatanFormPage.tsx",
    { id: "ref-kecamatan-edit" },
  ),
  route(
    "settings/sipil/kelurahan",
    "./features/ref/pages/KelurahanListPage.tsx",
  ),
  route(
    "settings/sipil/kelurahan/new",
    "./features/ref/pages/KelurahanFormPage.tsx",
    { id: "ref-kelurahan-new" },
  ),
  route(
    "settings/sipil/kelurahan/:id",
    "./features/ref/pages/KelurahanFormPage.tsx",
    { id: "ref-kelurahan-edit" },
  ),
  route("settings/paroki", "./features/ref/pages/ParokiHubPage.tsx"),
  route(
    "settings/paroki/profile",
    "./features/ref/pages/ParokiProfilePage.tsx",
  ),
  route("settings/paroki/wilayah", "./features/ref/pages/WilayahListPage.tsx"),
  route(
    "settings/paroki/wilayah/new",
    "./features/ref/pages/WilayahFormPage.tsx",
    { id: "ref-wilayah-new" },
  ),
  route(
    "settings/paroki/wilayah/:id",
    "./features/ref/pages/WilayahFormPage.tsx",
    { id: "ref-wilayah-edit" },
  ),
  route(
    "settings/paroki/lingkungan",
    "./features/ref/pages/LingkunganListPage.tsx",
  ),
  route(
    "settings/paroki/lingkungan/new",
    "./features/ref/pages/LingkunganFormPage.tsx",
    { id: "ref-lingkungan-new" },
  ),
  route(
    "settings/paroki/lingkungan/:id",
    "./features/ref/pages/LingkunganFormPage.tsx",
    { id: "ref-lingkungan-edit" },
  ),
  route("settings/paroki/periods", "./features/ref/pages/PeriodsPage.tsx"),
  route("settings/lookups", "./features/ref/pages/LookupsPage.tsx"),
  route("settings/lookups/new", "./features/ref/pages/LookupFormPage.tsx", {
    id: "ref-lookup-new",
  }),
  route("settings/lookups/:id", "./features/ref/pages/LookupFormPage.tsx", {
    id: "ref-lookup-edit",
  }),
] satisfies RouteConfig;
