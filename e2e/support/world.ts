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
  activeKeuskupanId?: string;
  activeParokiId?: string;
  activeParokiName?: string;
  activeParokiCode?: string;
  activeOrgId?: string;
  activeOrgName?: string;
  activeOrgCode?: string;
  activePeriodName?: string;
  otherOrgId?: string;
  otherOrgName?: string;
  orgAdminToken?: string;
  parokiAdminToken?: string;
  activeLiturgiEventId?: string;
  activeCantusId?: string;
  mockLlmBaseUrl?: string;

  constructor(options: IWorldOptions) {
    super(options);
    this.client = new TestClient();
  }
}

setWorldConstructor(DomusWorld);
