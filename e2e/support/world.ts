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

  constructor(options: IWorldOptions) {
    super(options);
    this.client = new TestClient();
  }
}

setWorldConstructor(DomusWorld);
