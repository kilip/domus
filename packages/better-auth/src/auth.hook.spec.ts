import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

vi.hoisted(() => {
  process.env.AUTH_URL = "http://localhost:8001";
  process.env.AUTH_PATH = "/api/auth";
  process.env.AUTH_DB_URL = "postgres://user:pass@localhost:5432/db";
  process.env.API_INTERNAL_URL = "http://localhost:8002";
  process.env.API_INTERNAL_KEY = "12345678901234567890123456789012";
});

import { drizzleAdapter } from "@better-auth/drizzle-adapter/relations-v2";
import { PGlite } from "@electric-sql/pglite";
import type { User } from "better-auth";
import { betterAuth } from "better-auth";
import { drizzle } from "drizzle-orm/pglite";
import { auth } from "./auth";
import { resetAuthEnvCache } from "./authEnv";
import * as schema from "./drizzle/schema/better-auth";

describe("auth databaseHooks.user.create.before", () => {
  const originalEnv = process.env;

  beforeEach(() => {
    process.env = { ...originalEnv };
    process.env.AUTH_URL = "http://localhost:8001";
    process.env.AUTH_PATH = "/api/auth";
    process.env.AUTH_DB_URL = "postgres://user:pass@localhost:5432/db";
    process.env.API_INTERNAL_URL = "http://localhost:8002";
    process.env.API_INTERNAL_KEY = "12345678901234567890123456789012";
    resetAuthEnvCache();
  });

  afterEach(() => {
    process.env = originalEnv;
    resetAuthEnvCache();
    vi.restoreAllMocks();
  });

  const getBeforeHook = () => {
    const hook = auth.options.databaseHooks?.user?.create?.before;
    expect(hook).toBeDefined();
    return hook;
  };

  it("should reject unverified email without calling internal resolve endpoint", async () => {
    const fetchSpy = vi.spyOn(globalThis, "fetch");
    const beforeHook = getBeforeHook();
    if (!beforeHook) return;

    const user: User = {
      id: "test-id",
      name: "Test User",
      email: "test@example.com",
      emailVerified: false,
      createdAt: new Date(),
      updatedAt: new Date(),
    };

    const result = await beforeHook(user);

    expect(result).toBe(false);
    expect(fetchSpy).not.toHaveBeenCalled();
  });

  it("should reject when internal API returns 404 (not invited)", async () => {
    const fetchSpy = vi.spyOn(globalThis, "fetch").mockResolvedValue({
      ok: false,
      status: 404,
    } as Response);

    const beforeHook = getBeforeHook();
    if (!beforeHook) return;

    const user: User = {
      id: "test-id",
      name: "Test User",
      email: "Uninvited@Example.COM",
      emailVerified: true,
      createdAt: new Date(),
      updatedAt: new Date(),
    };

    const result = await beforeHook(user);

    expect(result).toBe(false);
    expect(fetchSpy).toHaveBeenCalledWith(
      "http://localhost:8002/internal/users/resolve",
      expect.objectContaining({
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-Internal-Key": "12345678901234567890123456789012",
        },
        body: JSON.stringify({ email: "uninvited@example.com" }),
      }),
    );
  });

  it("should fail closed (reject) when internal API encounters a network error or timeout", async () => {
    vi.spyOn(globalThis, "fetch").mockRejectedValue(
      new Error("Connection refused"),
    );

    const beforeHook = getBeforeHook();
    if (!beforeHook) return;

    const user: User = {
      id: "test-id",
      name: "Test User",
      email: "test@example.com",
      emailVerified: true,
      createdAt: new Date(),
      updatedAt: new Date(),
    };

    const result = await beforeHook(user);

    expect(result).toBe(false);
  });

  it("should return user data with superadmin role when resolve endpoint returns 200 with superadmin", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({
        id: "01946890-0000-7000-8000-000000000001",
        status: "active",
        role: "superadmin",
      }),
    } as Response);

    const beforeHook = getBeforeHook();
    if (!beforeHook) return;

    const user: User = {
      id: "initial-id",
      name: "Super Admin",
      email: "Admin@Example.com",
      emailVerified: true,
      createdAt: new Date(),
      updatedAt: new Date(),
    };

    const result = await beforeHook(user);

    expect(result).toEqual({
      data: {
        ...user,
        role: "superadmin",
      },
    });
  });

  it("should return user data with user role when resolve endpoint returns 200 with user role", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({
        id: "01946890-0000-7000-8000-000000000002",
        status: "invited",
        role: "user",
      }),
    } as Response);

    const beforeHook = getBeforeHook();
    if (!beforeHook) return;

    const user: User = {
      id: "initial-id",
      name: "Normal User",
      email: "user@example.com",
      emailVerified: true,
      createdAt: new Date(),
      updatedAt: new Date(),
    };

    const result = await beforeHook(user);

    expect(result).toEqual({
      data: {
        ...user,
        role: "user",
      },
    });
  });

  it("should integrate with Better Auth and persist user with assigned role, without writing to public tables", async () => {
    const client = new PGlite();
    await client.exec(`
      CREATE SCHEMA IF NOT EXISTS auth;
      CREATE TABLE "auth"."user" (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        name TEXT NOT NULL,
        email TEXT NOT NULL UNIQUE,
        email_verified BOOLEAN NOT NULL DEFAULT false,
        image TEXT,
        created_at TIMESTAMP NOT NULL DEFAULT now(),
        updated_at TIMESTAMP NOT NULL DEFAULT now(),
        role TEXT,
        banned BOOLEAN DEFAULT false,
        ban_reason TEXT,
        ban_expires TIMESTAMP
      );
      CREATE TABLE "auth"."session" (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        expires_at TIMESTAMP NOT NULL,
        token TEXT NOT NULL UNIQUE,
        created_at TIMESTAMP NOT NULL DEFAULT now(),
        updated_at TIMESTAMP NOT NULL DEFAULT now(),
        ip_address TEXT,
        user_agent TEXT,
        user_id UUID NOT NULL REFERENCES "auth"."user"(id)
      );
      CREATE TABLE "auth"."account" (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        account_id TEXT NOT NULL,
        provider_id TEXT NOT NULL,
        user_id UUID NOT NULL REFERENCES "auth"."user"(id),
        access_token TEXT,
        refresh_token TEXT,
        id_token TEXT,
        access_token_expires_at TIMESTAMP,
        refresh_token_expires_at TIMESTAMP,
        scope TEXT,
        password TEXT,
        created_at TIMESTAMP NOT NULL DEFAULT now(),
        updated_at TIMESTAMP NOT NULL DEFAULT now()
      );
      CREATE TABLE "auth"."verification" (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        identifier TEXT NOT NULL,
        value TEXT NOT NULL,
        expires_at TIMESTAMP NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT now(),
        updated_at TIMESTAMP
      );
    `);

    const db = drizzle({ client });

    vi.spyOn(globalThis, "fetch").mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({
        id: "01946890-0000-7000-8000-000000000003",
        status: "invited",
        role: "superadmin",
      }),
    } as Response);

    const testAuth = betterAuth({
      baseURL: "http://localhost:8001",
      database: drizzleAdapter(db, {
        provider: "pg",
        schemaName: "auth",
        schema,
      }),
      advanced: {
        database: {
          generateId: "uuid",
        },
      },
      plugins: auth.options.plugins,
      databaseHooks: auth.options.databaseHooks,
    });

    const ctx = await testAuth.$context;
    const createdUser = await ctx.internalAdapter.createUser(
      {
        email: "invited@example.com",
        name: "Invited User",
        emailVerified: true,
        createdAt: new Date(),
        updatedAt: new Date(),
      },
      { method: "oauth" },
    );

    expect(createdUser).not.toBeNull();
    expect(createdUser.email).toBe("invited@example.com");
    expect(createdUser.role).toBe("superadmin");
    // Path B: Better Auth generates its own valid UUID
    expect(createdUser.id).toMatch(
      /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i,
    );

    // Verify row in database
    const dbRows = await client.query<{ id: string; role: string }>(
      `SELECT * FROM "auth"."user" WHERE email = 'invited@example.com'`,
    );
    expect(dbRows.rows).toHaveLength(1);
    expect(dbRows.rows[0]?.role).toBe("superadmin");
    expect(dbRows.rows[0]?.id).toBe(createdUser.id);
  });
});
