import { type RouteConfig, route } from "@react-router/dev/routes";

export default [
  route("org", "./features/org/pages/OrgsPage.tsx"),
  route("org/:orgId", "./features/org/pages/OrgDetailPage.tsx"),
  route(
    "org/:orgId/periode/new",
    "./features/org/pages/OrgPeriodeFormPage.tsx",
    { id: "org-periode-new" },
  ),
  route(
    "org/:orgId/periode/:periodeId/edit",
    "./features/org/pages/OrgPeriodeFormPage.tsx",
    { id: "org-periode-edit" },
  ),
  route(
    "org/:orgId/periode/:periodeId/jabatan/new",
    "./features/org/pages/OrgJabatanFormPage.tsx",
    { id: "org-jabatan-new" },
  ),
  route(
    "org/:orgId/periode/:periodeId/jabatan/:jabatanId/edit",
    "./features/org/pages/OrgJabatanFormPage.tsx",
    { id: "org-jabatan-edit" },
  ),
] satisfies RouteConfig;
