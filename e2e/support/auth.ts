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
  const email = (
    (overrides.email as string) ?? `e2e-${randomUUID()}@example.com`
  ).toLowerCase();
  const name = (overrides.name as string) ?? "E2E User";
  const role = (overrides.role as string) ?? "user";

  // Pre-register user in API database (public.users)
  await authDB.execute(
    sql`INSERT INTO public.users (id, email, name, status, auth_user_id)
        VALUES (${id}, ${email}, ${name}, 'active', ${id})
        ON CONFLICT (email) WHERE deleted_at IS NULL DO UPDATE SET auth_user_id = ${id}, status = 'active'`,
  );

  // If role is admin or superadmin, insert system:0 superadmin membership
  if (role === "admin" || role === "superadmin") {
    await authDB.execute(
      sql`INSERT INTO public.memberships (user_id, scope_type, scope_id, role, created_at)
          VALUES (${id}, 'system', 0, 'superadmin', NOW())
          ON CONFLICT DO NOTHING`,
    );
  }

  const user = t.createUser({
    id,
    email,
    name,
    emailVerified: true,
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
  try {
    await t.deleteUser(id);
  } catch (err) {
    console.error("[e2e] Failed to delete auth user:", err);
  }
  try {
    await authDB.execute(
      sql`DELETE FROM public.memberships WHERE user_id = ${id}`,
    );
    await authDB.execute(
      sql`DELETE FROM public.users WHERE id = ${id} OR auth_user_id = ${id}`,
    );
  } catch (err) {
    console.error("[e2e] Failed to clean test user from public tables:", err);
  }
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
