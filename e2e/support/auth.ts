import { randomUUID } from "node:crypto";
// NOTE: sesuaikan import ini dengan export instance auth di @domus/better-auth
import { auth, authDB } from "@domus/better-auth/auth";
import { sql } from "drizzle-orm";
import { env } from "./env";

async function testHelpers() {
  const ctx = await auth.$context;
  if (!ctx.test) {
    throw new Error(
      "testUtils plugin belum aktif. Jalankan auth server dengan E2E=true.",
    );
  }
  return ctx.test;
}

/** Bikin user test di DB + cookie sesi siap pakai buat Playwright. */
export async function createSession(overrides: Record<string, unknown> = {}) {
  const t = await testHelpers();
  const id = (overrides.id as string) ?? randomUUID();
  const user = t.createUser({
    id,
    email: `e2e-${randomUUID()}@example.com`,
    name: "E2E User",
    ...overrides,
  });
  const saved = await t.saveUser(user);
  const cookies = await t.getCookies({
    userId: saved.id,
    domain: new URL(env.baseUrl).hostname,
  });
  return { user: saved, cookies };
}

export async function removeUser(id: string) {
  const t = await testHelpers();
  await t.deleteUser(id);
}

export async function cleanTestDioceses() {
  try {
    await authDB.execute(
      sql`DELETE FROM public.dioceses WHERE name LIKE 'E2E%'`,
    );
  } catch (err) {
    console.error("[e2e] Failed to clean test dioceses:", err);
  }
}

export async function cleanTestTerritories() {
  try {
    await authDB.execute(
      sql`DELETE FROM public.territories WHERE name LIKE 'E2E%'`,
    );
  } catch (err) {
    console.error("[e2e] Failed to clean test territories:", err);
  }
}

export async function cleanTestWorkspaces() {
  try {
    await authDB.execute(
      sql`DELETE FROM public.workspaces WHERE name LIKE 'E2E%'`,
    );
  } catch (err) {
    console.error("[e2e] Failed to clean test workspaces:", err);
  }
}

export async function cleanTestParishes() {
  try {
    await authDB.execute(
      sql`DELETE FROM public.parishes WHERE name LIKE 'E2E%'`,
    );
  } catch (err) {
    console.error("[e2e] Failed to clean test parishes:", err);
  }
}

export async function cleanTestMemberships() {
  try {
    await authDB.execute(
      sql`DELETE FROM public.memberships WHERE role LIKE 'finance%'`,
    );
  } catch (err) {
    console.error("[e2e] Failed to clean test memberships:", err);
  }
}
