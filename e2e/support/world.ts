import {
  type IWorldOptions,
  setWorldConstructor,
  World,
} from "@cucumber/cucumber";
import type { BrowserContext, Page } from "@playwright/test";
import { LauncherPage } from "../pages/LauncherPage";
import { LoginPage } from "../pages/LoginPage";
import { DiocesePage } from "../pages/system/DiocesePage";
import { createSession } from "./auth";

export class DomusWorld extends World {
  context!: BrowserContext;
  page!: Page;
  userIds: string[] = [];

  private _loginPage?: LoginPage;
  private _launcherPage?: LauncherPage;
  private _diocesePage?: DiocesePage;

  constructor(options: IWorldOptions) {
    super(options);
  }

  get loginPage() {
    this._loginPage ??= new LoginPage(this.page);
    return this._loginPage;
  }

  get launcherPage() {
    this._launcherPage ??= new LauncherPage(this.page);
    return this._launcherPage;
  }

  get diocesePage() {
    this._diocesePage ??= new DiocesePage(this.page);
    return this._diocesePage;
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
