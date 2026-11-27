import { type RouteConfig, route } from "@react-router/dev/routes";

export default [
  route("liturgi", "./features/liturgi/pages/LiturgiSchedulesPage.tsx"),
  route(
    "liturgi/new",
    "./features/liturgi/pages/LiturgiScheduleCreatePage.tsx",
    { id: "liturgi-new" },
  ),
  route(
    "liturgi/ministries",
    "./features/liturgi/pages/LiturgiMinistriesPage.tsx",
    { id: "liturgi-ministries" },
  ),
  route("liturgi/:eventId", "./features/liturgi/pages/LiturgiDetailPage.tsx", {
    id: "liturgi-detail",
  }),
  route(
    "liturgi/:eventId/petugas/new",
    "./features/liturgi/pages/LiturgiPetugasFormPage.tsx",
    { id: "liturgi-petugas-new" },
  ),
  route(
    "liturgi/:eventId/petugas/:petugasId/edit",
    "./features/liturgi/pages/LiturgiPetugasFormPage.tsx",
    { id: "liturgi-petugas-edit" },
  ),
  route(
    "liturgi/:eventId/petugas/:petugasId/respond",
    "./features/liturgi/pages/LiturgiPetugasRespondPage.tsx",
    { id: "liturgi-petugas-respond" },
  ),
  route(
    "liturgi/:eventId/lagu/new",
    "./features/liturgi/pages/LiturgiLaguFormPage.tsx",
    { id: "liturgi-lagu-new" },
  ),
  route(
    "liturgi/:eventId/lagu/:laguId/edit",
    "./features/liturgi/pages/LiturgiLaguFormPage.tsx",
    { id: "liturgi-lagu-edit" },
  ),
] satisfies RouteConfig;
