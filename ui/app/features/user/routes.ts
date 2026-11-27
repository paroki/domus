import { type RouteConfig, route } from "@react-router/dev/routes";

export default [
  route("me", "./features/user/pages/MePage.tsx"),
  route("me/verify", "./features/user/pages/VerifyPage.tsx"),
  route("me/tokens/new", "./features/user/pages/TokenCreatePage.tsx"),
  route("me/accounts/whatsapp", "./features/user/pages/WhatsAppLinkPage.tsx"),
  route("me/accounts/telegram", "./features/user/pages/TelegramLinkPage.tsx"),
  route("users", "./features/user/pages/UserPage.tsx"),
  route("users/new", "./features/user/pages/UserCreatePage.tsx"),
  route("users/:id", "./features/user/pages/UserDetailPage.tsx"),
  route("users/:id/paroki/new", "./features/user/pages/UserParokiAddPage.tsx"),
  route("users/:id/org/new", "./features/user/pages/UserOrgAddPage.tsx"),
] satisfies RouteConfig;
