import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { resetAuthEnvCache } from "./authEnv";
import { seedSuperadmin } from "./seedSuperadmin";

describe("seedSuperadmin", () => {
  const originalEnv = process.env;

  beforeEach(() => {
    process.env = { ...originalEnv };
    process.env.AUTH_URL = "http://localhost:3000";
    process.env.AUTH_PATH = "/api/auth";
    process.env.AUTH_DB_URL = "postgres://user:pass@localhost:5432/db";
  });

  afterEach(() => {
    process.env = originalEnv;
    resetAuthEnvCache();
    vi.restoreAllMocks();
  });

  it("should skip seeding if SUPERADMIN_EMAIL is not configured", async () => {
    delete process.env.SUPERADMIN_EMAIL;
    const mockExecute = vi.fn();
    const mockDb = { execute: mockExecute };

    const result = await seedSuperadmin({ db: mockDb });

    expect(result.status).toBe("skipped");
    expect(mockExecute).not.toHaveBeenCalled();
  });

  it("should create new superadmin user when user does not exist", async () => {
    const executedQueries: string[] = [];
    const mockDb = {
      execute: vi
        .fn()
        .mockImplementation(
          async (query: { queryChunks?: Array<{ value?: string }> }) => {
            const text = query.queryChunks?.[0]?.value || "";
            executedQueries.push(text);
            if (text.includes("SELECT id, email")) {
              return { rows: [] };
            }
            return { rows: [] };
          },
        ),
    };

    const result = await seedSuperadmin({
      db: mockDb,
      email: "root@example.com",
      name: "Root Admin",
    });

    expect(result.status).toBe("created");
    expect(result.email).toBe("root@example.com");
    expect(result.userId).toMatch(
      /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i,
    );

    // Verify 4 executions: SELECT, INSERT auth.user, INSERT public.users, INSERT public.memberships
    expect(mockDb.execute).toHaveBeenCalledTimes(4);
  });

  it("should promote existing user to superadmin if role is different", async () => {
    const existingId = "11111111-2222-3333-4444-555555555555";
    const mockDb = {
      execute: vi
        .fn()
        .mockImplementation(
          async (query: { queryChunks?: Array<{ value?: string }> }) => {
            const text = query.queryChunks?.[0]?.value || "";
            if (text.includes("SELECT id, email")) {
              return {
                rows: [
                  {
                    id: existingId,
                    email: "existing@example.com",
                    name: "Existing User",
                    role: "user",
                    email_verified: false,
                  },
                ],
              };
            }
            return { rows: [] };
          },
        ),
    };

    const result = await seedSuperadmin({
      db: mockDb,
      email: "existing@example.com",
    });

    expect(result.status).toBe("promoted");
    expect(result.userId).toBe(existingId);
    expect(mockDb.execute).toHaveBeenCalledTimes(4);
  });

  it("should return already_superadmin if user is already superadmin", async () => {
    const existingId = "11111111-2222-3333-4444-555555555555";
    const mockDb = {
      execute: vi
        .fn()
        .mockImplementation(
          async (query: { queryChunks?: Array<{ value?: string }> }) => {
            const text = query.queryChunks?.[0]?.value || "";
            if (text.includes("SELECT id, email")) {
              return {
                rows: [
                  {
                    id: existingId,
                    email: "super@example.com",
                    name: "Super Admin",
                    role: "superadmin",
                    email_verified: true,
                  },
                ],
              };
            }
            return { rows: [] };
          },
        ),
    };

    const result = await seedSuperadmin({
      db: mockDb,
      email: "super@example.com",
    });

    expect(result.status).toBe("already_superadmin");
    expect(result.userId).toBe(existingId);
    expect(mockDb.execute).toHaveBeenCalledTimes(4);
  });

  it("should propagate database errors", async () => {
    const mockDb = {
      execute: vi.fn().mockRejectedValue(new Error("DB connection lost")),
    };

    await expect(
      seedSuperadmin({
        db: mockDb,
        email: "fail@example.com",
      }),
    ).rejects.toThrow("DB connection lost");
  });
});
