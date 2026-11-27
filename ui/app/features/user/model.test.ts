import assert from "node:assert/strict";
import test from "node:test";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { MemoryRouter } from "react-router";
import type { User } from "./api/types";
import { AccountList } from "./components/AccountList";
import { TokenList } from "./components/TokenList";
import {
  buildOnboardingSteps,
  canManageUsers,
  deduplicateOrgMembers,
  formatUserAgent,
  getAvailableUserDetailTabs,
  getAvailableUserScopes,
  getBirthDateBounds,
  getInitials,
  isCantusAdmin,
  isLiturgiAdmin,
  isOrgAdmin,
  isParokiAdmin,
  isSuperAdmin,
  isUmatAdmin,
  ME_ROUTES,
  parseUserDetailTab,
  providerLabel,
  ROLE_CANTUS_ADMIN,
  ROLE_LITURGI_ADMIN,
  ROLE_ORG_ADMIN,
  ROLE_PAROKI_ADMIN,
  ROLE_SUPER_ADMIN,
  ROLE_UMAT_ADMIN,
  ROLE_USER,
  sanitizeRedirect,
  shouldRedirectToOnboarding,
  USER_GENDERS,
} from "./model";
import MePage from "./pages/MePage";
import TelegramLinkPage from "./pages/TelegramLinkPage";
import TokenCreatePage from "./pages/TokenCreatePage";
import WhatsAppLinkPage from "./pages/WhatsAppLinkPage";

test("formatUserAgent detects common browsers and OS", () => {
  assert.equal(
    formatUserAgent(
      "Mozilla/5.0 (X11; Linux x86_64; rv:130.0) Gecko/20100101 Firefox/130.0",
    ),
    "Firefox · Linux",
  );
  assert.equal(
    formatUserAgent(
      "Mozilla/5.0 (Linux; Android 14) AppleWebKit/537.36 Chrome/128.0.0.0 Mobile Safari/537.36",
    ),
    "Chrome · Android",
  );
  assert.equal(formatUserAgent(""), "Perangkat tidak dikenal");
});

test("providerLabel maps known providers", () => {
  assert.equal(providerLabel("email"), "Email OTP");
  assert.equal(providerLabel("google"), "Google");
  assert.equal(providerLabel("github"), "GitHub");
  assert.equal(providerLabel("telegram"), "Telegram");
  assert.equal(providerLabel("whatsapp"), "WhatsApp");
  assert.equal(providerLabel("custom"), "custom");
});

test("isSuperAdmin checks user:super-admin role", () => {
  const admin = { roles: [ROLE_SUPER_ADMIN, ROLE_USER] } as User;
  const member = { roles: [ROLE_USER] } as User;
  assert.equal(isSuperAdmin(admin), true);
  assert.equal(isSuperAdmin(member), false);
  assert.equal(isSuperAdmin(null), false);
});

test("isCantusAdmin checks isCantusAdmin flag and cantus:admin role without granting /users access", () => {
  const byFlag = { isCantusAdmin: true, roles: [ROLE_USER] } as User;
  const byRole = { roles: [ROLE_CANTUS_ADMIN, ROLE_USER] } as User;
  const plain = { isCantusAdmin: false, roles: [ROLE_USER] } as User;

  assert.equal(isCantusAdmin(byFlag), true);
  assert.equal(isCantusAdmin(byRole), true);
  assert.equal(isCantusAdmin(plain), false);
  assert.equal(isCantusAdmin(null), false);
  assert.equal(canManageUsers(byRole), false);
  assert.deepEqual(getAvailableUserScopes([ROLE_CANTUS_ADMIN]), []);
  assert.deepEqual(getAvailableUserDetailTabs([ROLE_CANTUS_ADMIN]), []);
});

test("role helpers and getAvailableUserScopes resolve directory scopes", () => {
  const superUser = { roles: [ROLE_SUPER_ADMIN, ROLE_USER] } as User;
  const parokiUser = { roles: [ROLE_PAROKI_ADMIN, ROLE_USER] } as User;
  const orgUser = { roles: [ROLE_ORG_ADMIN, ROLE_USER] } as User;
  const umatUser = { roles: [ROLE_UMAT_ADMIN, ROLE_USER] } as User;
  const liturgiUser = { roles: [ROLE_LITURGI_ADMIN, ROLE_USER] } as User;
  const plainUser = { roles: [ROLE_USER] } as User;

  assert.equal(isParokiAdmin(parokiUser), true);
  assert.equal(isUmatAdmin(umatUser), true);
  assert.equal(isLiturgiAdmin(liturgiUser), true);
  assert.equal(isLiturgiAdmin(plainUser), false);
  assert.equal(isOrgAdmin(orgUser), true);
  assert.equal(canManageUsers(superUser), true);
  assert.equal(canManageUsers(parokiUser), true);
  assert.equal(canManageUsers(umatUser), true);
  assert.equal(canManageUsers(orgUser), true);
  assert.equal(canManageUsers(plainUser), false);

  assert.deepEqual(getAvailableUserScopes(superUser.roles), ["global"]);
  assert.deepEqual(getAvailableUserScopes(parokiUser.roles), ["org"]);
  assert.deepEqual(getAvailableUserScopes(umatUser.roles), ["org"]);
  assert.deepEqual(getAvailableUserScopes(orgUser.roles), ["org"]);
  assert.deepEqual(getAvailableUserScopes([ROLE_SUPER_ADMIN, ROLE_ORG_ADMIN]), [
    "global",
    "org",
  ]);
  assert.deepEqual(getAvailableUserScopes(plainUser.roles), []);
  assert.deepEqual(getAvailableUserScopes(null), []);
});

test("sanitizeRedirect rejects protocol-relative and external URLs", () => {
  assert.equal(sanitizeRedirect("/me"), "/me");
  assert.equal(sanitizeRedirect("/users?q=a"), "/users?q=a");
  assert.equal(sanitizeRedirect("//evil.example.com"), "/");
  assert.equal(sanitizeRedirect("https://evil.example.com"), "/");
  assert.equal(sanitizeRedirect(null), "/");
});

test("buildOnboardingSteps builds dynamic steps from otpChannels", () => {
  assert.deepEqual(buildOnboardingSteps(["email"]), ["biodata"]);
  assert.deepEqual(buildOnboardingSteps(["email", "whatsapp"]), [
    "biodata",
    "whatsapp",
  ]);
  assert.deepEqual(buildOnboardingSteps(["email", "telegram"]), [
    "biodata",
    "telegram",
  ]);
  assert.deepEqual(buildOnboardingSteps(["email", "whatsapp", "telegram"]), [
    "biodata",
    "whatsapp",
    "telegram",
  ]);
});

test("shouldRedirectToOnboarding avoids /onboarding loop and handles states", () => {
  assert.equal(shouldRedirectToOnboarding(false, "/onboarding"), false);
  assert.equal(shouldRedirectToOnboarding(false, "/"), true);
  assert.equal(shouldRedirectToOnboarding(false, "/me"), true);
  assert.equal(shouldRedirectToOnboarding(true, "/"), false);
  assert.equal(shouldRedirectToOnboarding(true, "/me"), false);
  assert.equal(shouldRedirectToOnboarding(undefined, "/"), false);
  assert.equal(shouldRedirectToOnboarding(undefined, "/me"), false);
});

test("USER_GENDERS and getBirthDateBounds define biodata constraints", () => {
  assert.deepEqual(USER_GENDERS, ["pria", "wanita"]);
  assert.deepEqual(getBirthDateBounds(new Date("2026-11-21T15:30:00Z")), {
    min: "1906-11-21",
    max: "2013-11-21",
  });
});

test("getInitials derives two-letter initials from name or email", () => {
  assert.equal(getInitials("Budi Santoso", "budi@example.com"), "BS");
  assert.equal(getInitials("Maria", "maria@example.com"), "MA");
  assert.equal(getInitials("", "admin@parish.id"), "AD");
  assert.equal(getInitials("", ""), "?");
});

test("parseUserDetailTab and getAvailableUserDetailTabs respect caller roles", () => {
  assert.deepEqual(getAvailableUserDetailTabs([ROLE_SUPER_ADMIN]), [
    "profile",
    "paroki",
    "org",
  ]);
  assert.deepEqual(getAvailableUserDetailTabs([ROLE_PAROKI_ADMIN]), [
    "paroki",
    "org",
  ]);
  assert.deepEqual(getAvailableUserDetailTabs([ROLE_UMAT_ADMIN]), ["paroki"]);
  assert.deepEqual(
    getAvailableUserDetailTabs([ROLE_UMAT_ADMIN, ROLE_ORG_ADMIN]),
    ["paroki", "org"],
  );
  assert.deepEqual(getAvailableUserDetailTabs([ROLE_ORG_ADMIN]), ["org"]);
  assert.deepEqual(getAvailableUserDetailTabs([ROLE_USER]), []);

  assert.equal(parseUserDetailTab(null), "profile");
  assert.equal(parseUserDetailTab("profile"), "profile");
  assert.equal(parseUserDetailTab("paroki"), "paroki");
  assert.equal(parseUserDetailTab("org"), "org");
  assert.equal(parseUserDetailTab("unknown"), "profile");
  assert.equal(parseUserDetailTab("profile", ["paroki", "org"]), "paroki");
  assert.equal(parseUserDetailTab("paroki", ["org"]), "org");
});

test("deduplicateOrgMembers merges multi-org memberships per user", () => {
  const result = deduplicateOrgMembers([
    {
      org: { id: "o1", name: "OMK" },
      members: [
        { userId: "u2", name: "Budi", email: "budi@p.id", role: "member" },
        { userId: "u1", name: "Ani", email: "ani@p.id", role: "admin" },
      ],
    },
    {
      org: { id: "o2", name: "Koor" },
      members: [
        { userId: "u1", name: "Ani", email: "ani@p.id", role: "member" },
      ],
    },
  ]);

  assert.equal(result.length, 2);
  assert.equal(result[0]?.id, "u1");
  assert.equal(result[0]?.orgs.length, 2);
  assert.equal(result[1]?.id, "u2");
});

function renderWithProviders(
  element: React.ReactElement,
  initialEntry = "/me",
) {
  const qc = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return renderToStaticMarkup(
    createElement(
      QueryClientProvider,
      { client: qc },
      createElement(MemoryRouter, { initialEntries: [initialEntry] }, element),
    ),
  );
}

test("ME_ROUTES defines /me sub-routes for token creation and account linking", () => {
  assert.equal(ME_ROUTES.index, "/me");
  assert.equal(ME_ROUTES.verify, "/me/verify");
  assert.equal(ME_ROUTES.tokenNew, "/me/tokens/new");
  assert.equal(ME_ROUTES.linkWhatsapp, "/me/accounts/whatsapp");
  assert.equal(ME_ROUTES.linkTelegram, "/me/accounts/telegram");
});

test("MePage SSR renders read-first profile <dl> without inline creation forms", () => {
  const html = renderWithProviders(
    createElement(MePage, {
      loaderData: undefined,
      params: {},
      matches: [] as never,
    }),
  );
  assert.ok(html.includes("Profil &amp; Keamanan"));
  assert.ok(html.includes("Profil Saya"));
  assert.ok(html.includes("<dl"));
  assert.ok(html.includes("Ubah"));
  assert.equal(html.includes("Simpan Profil"), false);
  assert.equal(html.includes("Nama token"), false);
});

test("TokenList and AccountList SSR link to dedicated sub-pages without inline forms", () => {
  const tokenHtml = renderWithProviders(createElement(TokenList));
  assert.ok(tokenHtml.includes('href="/me/tokens/new"'));
  assert.equal(tokenHtml.includes("Nama token"), false);

  const accountHtml = renderWithProviders(createElement(AccountList));
  assert.ok(accountHtml.includes("Akun Tertaut"));
  assert.equal(accountHtml.includes("Kirim Kode OTP WhatsApp"), false);
});

test("Me sub-pages SSR render InductivePageHeader and dedicated forms", () => {
  const tokenPageHtml = renderWithProviders(
    createElement(TokenCreatePage, {
      loaderData: undefined,
      params: {},
      matches: [] as never,
    }),
    "/me/tokens/new",
  );
  assert.ok(tokenPageHtml.includes("Buat Token API"));
  assert.ok(tokenPageHtml.includes("Nama token"));
  assert.ok(tokenPageHtml.includes("Masa berlaku"));

  const waPageHtml = renderWithProviders(
    createElement(WhatsAppLinkPage, {
      loaderData: undefined,
      params: {},
      matches: [] as never,
    }),
    "/me/accounts/whatsapp",
  );
  assert.ok(waPageHtml.includes("Hubungkan WhatsApp"));
  assert.ok(waPageHtml.includes("Nomor WhatsApp"));

  const tgPageHtml = renderWithProviders(
    createElement(TelegramLinkPage, {
      loaderData: undefined,
      params: {},
      matches: [] as never,
    }),
    "/me/accounts/telegram",
  );
  assert.ok(tgPageHtml.includes("Hubungkan Telegram"));
});
