import { type RouteConfig, route } from "@react-router/dev/routes";

export default [
  route("cantus", "./features/cantus/pages/CantusDirectoryPage.tsx"),
  route("cantus/new", "./features/cantus/pages/CantusCreateWizardPage.tsx", {
    id: "cantus-new",
  }),
  route("cantus/:cantusId", "./features/cantus/pages/CantusDetailPage.tsx", {
    id: "cantus-detail",
  }),
] satisfies RouteConfig;
