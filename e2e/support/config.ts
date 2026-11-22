import path from "node:path";

const e2eDir = path.resolve(import.meta.dir, "..");
const repoRoot = path.resolve(e2eDir, "..");
const reportsDir = path.join(e2eDir, "reports");
const coverageDir = path.join(repoRoot, "coverage");

const appPort = Number(process.env.APP_PORT ?? 8080);
const uiPort = Number(process.env.UI_PORT ?? 5173);

export const config = {
  repoRoot,
  e2eDir,
  reportsDir,
  logsDir: path.join(reportsDir, "logs"),
  uiCoverageDir: path.join(reportsDir, "coverage"),
  coverageDir,
  apiRawCoverageDir: path.join(coverageDir, "api-raw"),
  apiCoverageFile: path.join(coverageDir, "api-e2e.txt"),
  appPort,
  uiPort,
  apiBaseUrl: process.env.E2E_API_URL ?? `http://localhost:${appPort}/api`,
  uiBaseUrl: process.env.E2E_UI_URL ?? `http://localhost:${uiPort}`,
  autoStart: process.env.E2E_AUTO_START !== "false",
  appDbUrl:
    process.env.APP_DB_URL ??
    "postgres://domus:domus@localhost:5432/domus_test?sslmode=disable",
  appSecret: process.env.APP_SECRET ?? "domus-e2e-test-secret-key-min-32-bytes",
  superAdminEmail:
    process.env.USER_SUPER_ADMIN_EMAIL ?? "superadmin@domus.test",
  superAdminName: process.env.USER_SUPER_ADMIN_NAME ?? "Super Admin",
};
