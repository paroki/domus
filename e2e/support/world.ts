import {
  type IWorldOptions,
  setWorldConstructor,
  World,
} from "@cucumber/cucumber";
import type { BrowserContext, Page } from "@playwright/test";
import { LauncherPage } from "../pages/LauncherPage";
import { LoginPage } from "../pages/LoginPage";
import { createSession } from "./auth";

export class DomusWorld extends World {
  context!: BrowserContext;
  page!: Page;
  userIds: string[] = [];

  private _loginPage?: LoginPage;
  private _launcherPage?: LauncherPage;

  constructor(options: IWorldOptions) {
    super(options);
  }

  get loginPage() {
    return (this._loginPage ??= new LoginPage(this.page));
  }

  get launcherPage() {
    return (this._launcherPage ??= new LauncherPage(this.page));
  }

  /** Login tanpa OAuth: bikin user via testUtils lalu pasang cookie sesi. */
  async loginAs(overrides: Record<string, unknown> = {}) {
    const { user, cookies } = await createSession(overrides);
    this.userIds.push(user.id);
    await this.context.addCookies(cookies);
    return user;
  }
}

setWorldConstructor(DomusWorld);
