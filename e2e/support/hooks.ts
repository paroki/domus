import {
  After,
  AfterAll,
  Before,
  BeforeAll,
  Status,
  setDefaultTimeout,
} from "@cucumber/cucumber";
import { type Browser, chromium } from "@playwright/test";
import { removeUser } from "./auth";
import { env } from "./env";
import type { DomusWorld } from "./world";

setDefaultTimeout(30_000);

let browser: Browser;

BeforeAll(async () => {
  browser = await chromium.launch({ headless: !env.headed });
});

AfterAll(async () => {
  await browser.close();
});

Before(async function (this: DomusWorld) {
  this.context = await browser.newContext({
    baseURL: env.baseUrl,
    locale: "id-ID",
  });
  this.page = await this.context.newPage();
});

After(async function (this: DomusWorld, scenario) {
  if (scenario.result?.status === Status.FAILED) {
    const png = await this.page.screenshot({ fullPage: true });
    this.attach(png, "image/png");
  }
  await this.context.close();
  for (const id of this.userIds) await removeUser(id);
  this.userIds = [];
});
