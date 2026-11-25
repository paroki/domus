import { type ChildProcess, spawn } from "node:child_process";
import { createWriteStream, mkdirSync, rmSync } from "node:fs";
import path from "node:path";
import { SQL } from "bun";
import { config } from "./config";

interface ManagedProcess {
  name: string;
  proc: ChildProcess;
  exited: Promise<number | null>;
}

const processes: ManagedProcess[] = [];

function spawnLogged(
  name: string,
  command: string,
  args: string[],
  cwd: string,
  env: Record<string, string | undefined>,
  logFile: string,
): ManagedProcess {
  const logStream = createWriteStream(logFile, { flags: "w" });
  const proc = spawn(command, args, {
    cwd,
    env: { ...process.env, ...env },
    detached: true,
    stdio: ["ignore", "pipe", "pipe"],
  });

  proc.stdout?.pipe(logStream);
  proc.stderr?.pipe(logStream);

  const exited = new Promise<number | null>((resolve) => {
    proc.once("exit", (code) => {
      logStream.end();
      resolve(code);
    });
  });

  const managed = { name, proc, exited };
  processes.push(managed);
  return managed;
}

async function waitForHttp(
  url: string,
  managed: ManagedProcess,
  timeoutMs = 60_000,
): Promise<void> {
  const deadline = Date.now() + timeoutMs;
  let earlyExitCode: number | null | undefined;
  void managed.exited.then((code) => {
    earlyExitCode = code;
  });

  while (Date.now() < deadline) {
    if (earlyExitCode !== undefined) {
      throw new Error(
        `${managed.name} exited prematurely with code ${earlyExitCode}`,
      );
    }
    try {
      const res = await fetch(url);
      if (res.status === 200) return;
    } catch {
      // Server not listening yet
    }
    await Bun.sleep(250);
  }
  throw new Error(`Timed out waiting for ${managed.name} at ${url}`);
}

export async function startServers(): Promise<void> {
  if (!config.autoStart) return;

  mkdirSync(config.logsDir, { recursive: true });
  rmSync(config.apiRawCoverageDir, { recursive: true, force: true });
  mkdirSync(config.apiRawCoverageDir, { recursive: true });

  const db = new SQL(config.appDbUrl);
  try {
    await db`DROP SCHEMA IF EXISTS public CASCADE`;
    await db`CREATE SCHEMA public`;
    await db`CREATE EXTENSION IF NOT EXISTS vector`;
    await db`CREATE EXTENSION IF NOT EXISTS pg_search CASCADE`;
  } finally {
    await db.close();
  }

  const apiProc = spawnLogged(
    "api",
    "go",
    [
      "run",
      "-cover",
      "-coverpkg=github.com/paroki/domus/internal/...,github.com/paroki/domus/cmd/...",
      "./cmd/domus",
      "serve",
    ],
    config.repoRoot,
    {
      APP_PORT: String(config.appPort),
      APP_DB_URL: config.appDbUrl,
      APP_AUTO_MIGRATE: "true",
      APP_SECRET: config.appSecret,
      GOCOVERDIR: config.apiRawCoverageDir,
      USER_SUPER_ADMIN_EMAIL: config.superAdminEmail,
      USER_SUPER_ADMIN_NAME: config.superAdminName,
    },
    path.join(config.logsDir, "api.log"),
  );

  await waitForHttp(`${config.apiBaseUrl}/health`, apiProc, 180_000);

  const reactRouterBin = path.join(
    config.repoRoot,
    "ui/node_modules/.bin/react-router",
  );
  const uiProc = spawnLogged(
    "ui",
    "bun",
    [
      "--conditions=development",
      reactRouterBin,
      "dev",
      "--port",
      String(config.uiPort),
    ],
    path.join(config.repoRoot, "ui"),
    {
      VITE_COVERAGE: "true",
      VITE_API_TARGET: `http://127.0.0.1:${config.appPort}`,
    },
    path.join(config.logsDir, "ui.log"),
  );

  await waitForHttp(`${config.uiBaseUrl}/login`, uiProc);
}

export async function stopServers(): Promise<void> {
  const list = processes.splice(0, processes.length);
  await Promise.all(
    list.map(async ({ proc, exited }) => {
      if (proc.pid && proc.exitCode === null) {
        try {
          process.kill(-proc.pid, "SIGTERM");
        } catch {
          try {
            proc.kill("SIGTERM");
          } catch {
            // Already exited
          }
        }
      }
      const timer = setTimeout(() => {
        if (proc.pid && proc.exitCode === null) {
          try {
            process.kill(-proc.pid, "SIGKILL");
          } catch {
            // Already exited
          }
        }
      }, 5000);
      await exited;
      clearTimeout(timer);
    }),
  );
}
