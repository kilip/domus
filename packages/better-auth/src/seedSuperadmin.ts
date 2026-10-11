import { randomUUID } from "node:crypto";
import { sql } from "drizzle-orm";
import { authEnv } from "./authEnv";
import { authDB } from "./drizzle";

export interface SeedSuperadminOptions {
  db?: {
    execute: (query: ReturnType<typeof sql>) => Promise<unknown>;
  };
  email?: string;
  name?: string;
}

export interface SeedSuperadminResult {
  status: "skipped" | "created" | "promoted" | "already_superadmin";
  email?: string;
  userId?: string;
}

const UUID_REGEX =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

export async function seedSuperadmin(
  options: SeedSuperadminOptions = {},
): Promise<SeedSuperadminResult> {
  const db = options.db ?? authDB;
  const rawEmail = options.email ?? authEnv.SUPERADMIN_EMAIL;
  const email = rawEmail?.trim().toLowerCase();
  const name =
    (options.name ?? authEnv.SUPERADMIN_NAME)?.trim() || "Super Admin";

  if (!email) {
    console.log(
      "[better-auth] SUPERADMIN_EMAIL not set, skipping superadmin seeding.",
    );
    return { status: "skipped" };
  }

  try {
    const queryResult = await db.execute(
      sql`SELECT id, email, name, role, email_verified FROM auth.user WHERE LOWER(email) = ${email} LIMIT 1`,
    );

    const rows = (
      Array.isArray(queryResult)
        ? queryResult
        : ((queryResult as { rows?: unknown[] })?.rows ?? [])
    ) as Array<{
      id: string;
      email: string;
      name: string;
      role: string | null;
      email_verified: boolean;
    }>;

    const existingUser = rows[0];
    let userId: string;
    let resultStatus: "created" | "promoted" | "already_superadmin";

    if (!existingUser) {
      userId = randomUUID();
      await db.execute(
        sql`INSERT INTO auth.user (id, name, email, email_verified, role, created_at, updated_at)
            VALUES (${userId}, ${name}, ${email}, true, 'superadmin', NOW(), NOW())`,
      );
      resultStatus = "created";
    } else {
      userId = existingUser.id;
      await db.execute(
        sql`UPDATE auth.user
            SET role = 'superadmin',
                email_verified = true,
                updated_at = NOW()
            WHERE id = ${userId}`,
      );
      resultStatus =
        existingUser.role === "superadmin" ? "already_superadmin" : "promoted";
    }

    if (UUID_REGEX.test(userId)) {
      await db.execute(
        sql`INSERT INTO public.users (id, name, email, avatar)
            VALUES (${userId}, ${name}, ${email}, null)
            ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email`,
      );

      await db.execute(
        sql`INSERT INTO public.memberships (user_id, scope_type, scope_id, role, created_at)
            VALUES (${userId}, 'system', 0, 'superadmin', NOW())
            ON CONFLICT (user_id, scope_type, scope_id, role) DO NOTHING`,
      );
    } else {
      console.warn(
        `[better-auth] User ID ${userId} is not a valid UUID, skipping public.users and public.memberships sync.`,
      );
    }

    console.log(
      `[better-auth] Super admin (${email}) successfully ${resultStatus === "created" ? "created" : "ensured as superadmin"}.`,
    );

    return {
      status: resultStatus,
      email,
      userId,
    };
  } catch (err) {
    console.error("[better-auth] Failed to seed superadmin user:", err);
    throw err;
  }
}
