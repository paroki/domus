import {
  index,
  layout,
  type RouteConfig,
  type RouteConfigEntry,
  route,
} from "@react-router/dev/routes";

// Every features/<feature>/routes.ts is picked up automatically: add a feature
// folder with a routes.ts and its routes are registered, no edit needed here.
const featureRoutes = import.meta.glob<RouteConfigEntry[]>(
  "./features/*/routes.ts",
  { eager: true, import: "default" },
);

export default [
  route("login", "./features/user/pages/LoginPage.tsx"),
  route("org/join/:code", "./features/org/pages/OrgJoinPage.tsx"),
  layout("./features/user/components/RequireAuth.tsx", [
    route("onboarding", "./features/user/pages/OnboardingPage.tsx"),
    layout("./shell/AppShell.tsx", [
      index("./shell/HomePage.tsx"),
      ...Object.values(featureRoutes).flat(),
    ]),
  ]),
] satisfies RouteConfig;
