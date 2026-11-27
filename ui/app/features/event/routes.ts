import { type RouteConfig, route } from "@react-router/dev/routes";

export default [
  route("event", "./features/event/pages/EventsPage.tsx"),
  route("event/new", "./features/event/pages/EventCreatePage.tsx"),
  route("event/:eventId", "./features/event/pages/EventDetailPage.tsx"),
] satisfies RouteConfig;
