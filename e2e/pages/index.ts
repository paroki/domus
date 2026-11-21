import type { Page } from "playwright";
import { BasePage } from "./base.page";
import { LookupsPage } from "./ref/lookups.page";
import { ParishesPage } from "./ref/parishes.page";
import { SettingsPage } from "./ref/settings.page";
import { HomePage } from "./shell/home.page";
import { LoginPage } from "./user/login.page";

export function createPages(page: Page) {
  return {
    base: new BasePage(page),
    ref: {
      settings: new SettingsPage(page),
      parishes: new ParishesPage(page),
      lookups: new LookupsPage(page),
    },
    shell: {
      home: new HomePage(page),
    },
    user: {
      login: new LoginPage(page),
    },
  };
}

export type AppPages = ReturnType<typeof createPages>;
