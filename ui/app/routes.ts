import {
  index,
  layout,
  type RouteConfig,
  type RouteConfigEntry,
} from "@react-router/dev/routes";

// Every features/<feature>/routes.ts is picked up automatically: add a feature
// folder with a routes.ts and its routes are registered, no edit needed here.
const featureRoutes = import.meta.glob<RouteConfigEntry[]>(
  "./features/*/routes.ts",
  { eager: true, import: "default" },
);

export default [
  layout("./shell/AppShell.tsx", [
    index("./shell/HomePage.tsx"),
    ...Object.values(featureRoutes).flat(),
  ]),
] satisfies RouteConfig;
