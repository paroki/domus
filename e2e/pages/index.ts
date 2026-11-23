import type { Page } from "playwright";
import { BasePage } from "./base.page";
import { EventDetailPage } from "./event/event-detail.page";
import { EventsPage } from "./event/events.page";
import { OrgDetailPage } from "./org/org-detail.page";
import { OrgJoinPage } from "./org/org-join.page";
import { OrgsPage } from "./org/orgs.page";
import { LookupsPage } from "./ref/lookups.page";
import { ParishesPage } from "./ref/parishes.page";
import { PeriodsPage } from "./ref/periods.page";
import { ParishProfilePage } from "./ref/profile.page";
import { SettingsPage } from "./ref/settings.page";
import { TerritoryPage } from "./ref/territory.page";
import { SystemSettingsPage } from "./settings/system-settings.page";
import { HomePage } from "./shell/home.page";
import { LoginPage } from "./user/login.page";
import { MePage } from "./user/me.page";
import { OnboardingPage } from "./user/onboarding.page";
import { UsersPage } from "./user/users.page";

export function createPages(page: Page) {
  return {
    base: new BasePage(page),
    event: {
      events: new EventsPage(page),
      detail: new EventDetailPage(page),
    },
    org: {
      orgs: new OrgsPage(page),
      detail: new OrgDetailPage(page),
      join: new OrgJoinPage(page),
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
    settings: new SystemSettingsPage(page),
    user: {
      login: new LoginPage(page),
      users: new UsersPage(page),
      onboarding: new OnboardingPage(page),
      me: new MePage(page),
    },
  };
}

export type AppPages = ReturnType<typeof createPages>;
