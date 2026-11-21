import type { Page } from "playwright";
import { BasePage } from "./base.page";
import { OrgDetailPage } from "./org/org-detail.page";
import { OrgsPage } from "./org/orgs.page";
import { LookupsPage } from "./ref/lookups.page";
import { ParishesPage } from "./ref/parishes.page";
import { PeriodsPage } from "./ref/periods.page";
import { ParishProfilePage } from "./ref/profile.page";
import { SettingsPage } from "./ref/settings.page";
import { TerritoryPage } from "./ref/territory.page";
import { HomePage } from "./shell/home.page";
import { LoginPage } from "./user/login.page";
import { UsersPage } from "./user/users.page";

export function createPages(page: Page) {
  return {
    base: new BasePage(page),
    org: {
      orgs: new OrgsPage(page),
      detail: new OrgDetailPage(page),
    },
    ref: {
      settings: new SettingsPage(page),
      parishes: new ParishesPage(page),
      lookups: new LookupsPage(page),
      profile: new ParishProfilePage(page),
      territory: new TerritoryPage(page),
      periods: new PeriodsPage(page),
    },
    shell: {
      home: new HomePage(page),
    },
    user: {
      login: new LoginPage(page),
      users: new UsersPage(page),
    },
  };
}

export type AppPages = ReturnType<typeof createPages>;
