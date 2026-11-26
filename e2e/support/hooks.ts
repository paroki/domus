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
import {
  bootstrapOrgAdminSession,
  bootstrapOrgMemberSession,
  bootstrapParokiAdminSession,
  loginAsSuperAdmin,
  loginAsUser,
} from "./auth";
import { config } from "./config";
import { startMockLLMServer, stopMockLLMServer } from "./mock-llm";
import { startServers, stopServers } from "./servers";
import type { DomusWorld } from "./world";

setDefaultTimeout(30_000);

let browser: Browser;
let mockLlmBaseUrl = "";
const coverageMap = libCoverage.createCoverageMap({});

BeforeAll({ timeout: 90_000 }, async () => {
  mockLlmBaseUrl = await startMockLLMServer();
  await startServers();
  browser = await chromium.launch();
});

Before(async function (this: DomusWorld) {
  this.mockLlmBaseUrl = mockLlmBaseUrl;
  this.context = await browser.newContext({ baseURL: config.uiBaseUrl });
  this.page = await this.context.newPage();
  this.pages = createPages(this.page);
});

Before({ tags: "@super-admin" }, async function (this: DomusWorld) {
  const session = await loginAsSuperAdmin();
  await this.context.addCookies(session.cookies);
  this.client.setToken(session.rawToken);
});

Before({ tags: "@authenticated" }, async function (this: DomusWorld) {
  const session = await loginAsUser({
    email: "user@domus.test",
    name: "Regular User",
  });
  await this.context.addCookies(session.cookies);
  this.client.setToken(session.rawToken);
});

Before({ tags: "@paroki-admin" }, async function (this: DomusWorld) {
  const result = await bootstrapParokiAdminSession(this.client);
  await this.context.addCookies(result.session.cookies);
  this.client.setToken(result.session.rawToken);
  this.client.setParokiId(result.paroki.id);
  this.activeKeuskupanId = result.keuskupan.id;
  this.activeParokiId = result.paroki.id;
  this.activeParokiName = result.paroki.nama;
  this.activeParokiCode = result.paroki.kode;
});

Before({ tags: "@org-admin" }, async function (this: DomusWorld) {
  const result = await bootstrapOrgAdminSession(this.client);
  await this.context.addCookies(result.session.cookies);
  this.client.setToken(result.session.rawToken);
  this.client.setParokiId(result.paroki.id);
  this.activeParokiId = result.paroki.id;
  this.activeParokiName = result.paroki.nama;
  this.activeParokiCode = result.paroki.kode;
  this.activeOrgId = result.org.id;
  this.activeOrgName = result.org.name;
  this.activeOrgCode = result.org.code;
  this.activePeriodName = result.period.name;
  this.parokiAdminToken = result.parokiAdminToken;
});

Before({ tags: "@org-member" }, async function (this: DomusWorld) {
  const result = await bootstrapOrgMemberSession(this.client);
  await this.context.addCookies(result.session.cookies);
  this.client.setToken(result.session.rawToken);
  this.client.setParokiId(result.paroki.id);
  this.activeParokiId = result.paroki.id;
  this.activeParokiName = result.paroki.nama;
  this.activeParokiCode = result.paroki.kode;
  this.activeOrgId = result.orgA.id;
  this.activeOrgName = result.orgA.name;
  this.otherOrgId = result.orgB.id;
  this.otherOrgName = result.orgB.name;
  this.orgAdminToken = result.orgAdminToken;
  this.parokiAdminToken = result.parokiAdminToken;
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
  await stopMockLLMServer();

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
