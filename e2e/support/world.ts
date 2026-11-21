import {
  type IWorldOptions,
  setWorldConstructor,
  World,
} from "@cucumber/cucumber";
import type { BrowserContext, Page } from "playwright";
import type { AppPages } from "../pages";
import { TestClient } from "./test-client";

export class DomusWorld extends World {
  context!: BrowserContext;
  page!: Page;
  pages!: AppPages;
  readonly client: TestClient;
  activeDioceseId?: string;
  activeParishId?: string;
  activeParishName?: string;
  activeParishCode?: string;
  activeOrgId?: string;
  activeOrgName?: string;
  activeOrgCode?: string;
  activePeriodName?: string;
  otherOrgId?: string;
  otherOrgName?: string;
  orgAdminToken?: string;
  parishAdminToken?: string;
  mockLlmBaseUrl?: string;

  constructor(options: IWorldOptions) {
    super(options);
    this.client = new TestClient();
  }
}

setWorldConstructor(DomusWorld);
