import { execFileSync } from "node:child_process";
import { existsSync, mkdirSync, readdirSync } from "node:fs";
import {
  After,
  AfterAll,
  Before,
  BeforeAll,
  setDefaultTimeout,
} from "@cucumber/cucumber";
import libCoverage, { type CoverageMapData } from "istanbul-lib-coverage";
import libReport from "istanbul-lib-report";
import reports from "istanbul-reports";
import { type Browser, chromium } from "playwright";
import { createPages } from "../pages";
import { loginAsSuperAdmin, loginAsUser } from "./auth";
import { config } from "./config";
import { startServers, stopServers } from "./servers";
import type { DomusWorld } from "./world";

setDefaultTimeout(30_000);

let browser: Browser;
const coverageMap = libCoverage.createCoverageMap({});

BeforeAll({ timeout: 90_000 }, async () => {
  await startServers();
  browser = await chromium.launch();
});

Before(async function (this: DomusWorld) {
  this.context = await browser.newContext({ baseURL: config.uiBaseUrl });
  this.page = await this.context.newPage();
  this.pages = createPages(this.page);
});

Before({ tags: "@super-admin" }, async function (this: DomusWorld) {
  const session = loginAsSuperAdmin();
  await this.context.addCookies(session.cookies);
  this.client.setToken(session.rawToken);
});

Before({ tags: "@authenticated" }, async function (this: DomusWorld) {
  const session = loginAsUser({
    email: "user@domus.test",
    name: "Regular User",
  });
  await this.context.addCookies(session.cookies);
  this.client.setToken(session.rawToken);
});

After(async function (this: DomusWorld) {
  try {
    const cov = await this.page.evaluate(
      () =>
        (window as unknown as { __coverage__?: CoverageMapData }).__coverage__,
    );
    if (cov) {
      coverageMap.merge(cov);
    }
  } catch {
    // Page may have closed or navigated away
  }
  await this.page?.close();
  await this.context?.close();
});

AfterAll(async () => {
  await browser?.close();
  await stopServers();

  if (Object.keys(coverageMap.data).length > 0) {
    mkdirSync(config.coverageDir, { recursive: true });
    mkdirSync(config.uiCoverageDir, { recursive: true });

    const rootCtx = libReport.createContext({
      dir: config.coverageDir,
      coverageMap,
    });
    reports.create("lcovonly").execute(rootCtx);
    reports.create("text-summary").execute(rootCtx);

    const htmlCtx = libReport.createContext({
      dir: config.uiCoverageDir,
      coverageMap,
    });
    reports.create("html").execute(htmlCtx);
  }

  if (
    config.autoStart &&
    existsSync(config.apiRawCoverageDir) &&
    readdirSync(config.apiRawCoverageDir).length > 0
  ) {
    execFileSync(
      "go",
      [
        "tool",
        "covdata",
        "textfmt",
        `-i=${config.apiRawCoverageDir}`,
        `-o=${config.apiCoverageFile}`,
      ],
      { cwd: config.repoRoot, stdio: "inherit" },
    );
  }
});
