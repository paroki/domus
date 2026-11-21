import type { Page } from "playwright";
import { BasePage } from "./base.page";
import { HomePage } from "./shell/home.page";
import { LoginPage } from "./user/login.page";

export function createPages(page: Page) {
  return {
    base: new BasePage(page),
    shell: {
      home: new HomePage(page),
    },
    user: {
      login: new LoginPage(page),
    },
  };
}

export type AppPages = ReturnType<typeof createPages>;
