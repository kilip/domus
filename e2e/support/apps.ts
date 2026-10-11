import { type ChildProcess, spawn } from "node:child_process";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { env } from "./env";

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
const rootDir = resolve(__dirname, "../..");

export interface AppServer {
  name: string;
  cwd: string;
  command: string;
  args: string[];
  url: string;
  checkUrl: string;
  env?: NodeJS.ProcessEnv;
  enabled: boolean;
}

const managedProcesses: Map<string, ChildProcess> = new Map();
let isCleanedUp = false;
let handlersRegistered = false;

function killProcessGroup(pid: number, signal: NodeJS.Signals = "SIGTERM") {
  try {
    if (process.platform === "win32") {
      process.kill(pid, signal);
    } else {
      process.kill(-pid, signal);
    }
  } catch (err: unknown) {
    const error = err as NodeJS.ErrnoException;
    if (error.code !== "ESRCH") {
      try {
        process.kill(pid, signal);
      } catch {}
    }
  }
}

async function stopChildProcess(child: ChildProcess): Promise<void> {
  if (!child.pid || child.exitCode !== null) return;

  killProcessGroup(child.pid, "SIGTERM");

  const exited = await new Promise<boolean>((resolve) => {
    const timer = setTimeout(() => resolve(false), 3000);
    child.once("exit", () => {
      clearTimeout(timer);
      resolve(true);
    });
  });

  if (!exited && child.exitCode === null) {
    killProcessGroup(child.pid, "SIGKILL");
  }
}

function stopAppsSync(): void {
  for (const child of managedProcesses.values()) {
    if (child.pid && child.exitCode === null) {
      killProcessGroup(child.pid, "SIGKILL");
    }
  }
  managedProcesses.clear();
}

function registerExitHandlers() {
  if (handlersRegistered) return;
  handlersRegistered = true;

  process.once("exit", () => {
    stopAppsSync();
  });

  process.once("SIGINT", () => {
    stopAppsSync();
    process.exit(130);
  });

  process.once("SIGTERM", () => {
    stopAppsSync();
    process.exit(143);
  });
}

export function getAppServers(): AppServer[] {
  const cleanEnv: NodeJS.ProcessEnv = { ...process.env };
  delete cleanEnv.NODE_OPTIONS;

  return [
    {
      name: "auth",
      cwd: resolve(rootDir, "apps/auth"),
      command: "bun",
      args: ["run", "src/main.ts"],
      url: env.authUrl,
      checkUrl: `${env.authUrl}/health`,
      env: {
        ...cleanEnv,
        E2E: "true",
        AUTH_PORT: new URL(env.authUrl).port || "8001",
      },
      enabled: true,
    },
    {
      name: "dash",
      cwd: resolve(rootDir, "apps/dash"),
      command: "bun",
      args: ["run", "dev"],
      url: env.baseUrl,
      checkUrl: env.baseUrl,
      env: {
        ...cleanEnv,
        PORT: new URL(env.baseUrl).port || "3001",
      },
      enabled: true,
    },
    {
      name: "api",
      cwd: resolve(rootDir, "apps/api"),
      command: "go",
      args: ["run", "./cmd/api"],
      url: env.apiUrl,
      checkUrl: `${env.apiUrl}/live`,
      env: {
        ...cleanEnv,
        API_PORT: new URL(env.apiUrl).port || "8002",
        AUTH_URL: env.authUrl,
        AUTH_JWKS_URL: `${env.authUrl}/jwks`,
      },
      enabled: env.startApi,
    },
  ];
}

export async function isUrlReady(
  url: string,
  timeoutMs = 1500,
): Promise<boolean> {
  try {
    const res = await fetch(url, {
      method: "GET",
      signal: AbortSignal.timeout(timeoutMs),
      redirect: "manual",
      headers: { Accept: "*/*" },
    });
    return res.status < 500;
  } catch {
    return false;
  }
}

async function startApp(app: AppServer): Promise<void> {
  const alreadyRunning = await isUrlReady(app.checkUrl);
  if (alreadyRunning) {
    console.log(
      `[e2e] ${app.name} is already running on ${app.url}, reusing existing server.`,
    );
    return;
  }

  console.log(
    `[e2e] Starting ${app.name} (${app.command} ${app.args.join(" ")}) on ${app.url}...`,
  );

  let outputBuffer = "";
  const child = spawn(app.command, app.args, {
    cwd: app.cwd,
    env: app.env,
    stdio: ["ignore", "pipe", "pipe"],
    detached: process.platform !== "win32",
  });

  child.stdout?.on("data", (data) => {
    outputBuffer += data.toString();
    if (outputBuffer.length > 20000) {
      outputBuffer = outputBuffer.slice(-20000);
    }
  });

  child.stderr?.on("data", (data) => {
    outputBuffer += data.toString();
    if (outputBuffer.length > 20000) {
      outputBuffer = outputBuffer.slice(-20000);
    }
  });

  managedProcesses.set(app.name, child);

  const startTime = Date.now();
  const timeout = 60_000;

  while (Date.now() - startTime < timeout) {
    if (child.exitCode !== null) {
      throw new Error(
        `[e2e] ${app.name} exited prematurely with code ${child.exitCode}.\nOutput:\n${outputBuffer}`,
      );
    }
    if (await isUrlReady(app.checkUrl)) {
      console.log(`[e2e] ${app.name} is ready on ${app.url}.`);
      return;
    }
    await new Promise((r) => setTimeout(r, 200));
  }

  throw new Error(
    `[e2e] Timed out waiting for ${app.name} (${app.url}) to become ready after ${timeout}ms.\nOutput:\n${outputBuffer}`,
  );
}

export async function startApps(): Promise<void> {
  if (!env.autoStart) {
    console.log("[e2e] Auto-start is disabled (E2E_AUTO_START=false).");
    return;
  }

  registerExitHandlers();

  const apps = getAppServers().filter((app) => app.enabled);
  try {
    await Promise.all(apps.map((app) => startApp(app)));
  } catch (error) {
    console.error(
      "[e2e] Failed to start apps, stopping any spawned processes...",
      error,
    );
    await stopApps();
    throw error;
  }
}

export async function stopApps(): Promise<void> {
  if (isCleanedUp) return;
  isCleanedUp = true;

  if (managedProcesses.size === 0) return;
  console.log(`[e2e] Stopping ${managedProcesses.size} auto-started app(s)...`);

  const stopPromises = Array.from(managedProcesses.entries()).map(
    async ([name, child]) => {
      await stopChildProcess(child);
      console.log(`[e2e] Stopped ${name}.`);
    },
  );

  await Promise.all(stopPromises);
  managedProcesses.clear();
}
