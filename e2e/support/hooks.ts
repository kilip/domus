import {
  After,
  AfterAll,
  Before,
  BeforeAll,
  Status,
  setDefaultTimeout,
} from "@cucumber/cucumber";
import { type Browser, chromium } from "@playwright/test";
import { startApps, stopApps } from "./apps";
import {
  cleanTestDioceses,
  cleanTestParishes,
  cleanTestTerritories,
  removeUser,
} from "./auth";
import { env } from "./env";
import type { DomusWorld } from "./world";

setDefaultTimeout(30_000);

let browser: Browser;

BeforeAll({ timeout: 60_000 }, async () => {
  await startApps();
  browser = await chromium.launch({ headless: !env.headed });
  const page = await browser.newPage();
  try {
    await page.goto(env.baseUrl, { timeout: 20_000 });
  } catch {
    // Warm-up is best-effort
  } finally {
    await page.close();
  }
});

AfterAll({ timeout: 15_000 }, async () => {
  try {
    await browser?.close();
  } finally {
    await stopApps();
  }
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
  await cleanTestTerritories();
  await cleanTestParishes();
  await cleanTestDioceses();
  for (const id of this.userIds) await removeUser(id);
  this.userIds = [];
});
