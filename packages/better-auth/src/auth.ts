import { drizzleAdapter } from "@better-auth/drizzle-adapter/relations-v2";
import { betterAuth } from "better-auth";
import { admin, jwt, openAPI, testUtils } from "better-auth/plugins";
import { adminAc, defaultRoles } from "better-auth/plugins/admin/access";
import { sql } from "drizzle-orm";
import { authEnv } from "./authEnv";
import { authDB, schema } from "./drizzle";
import { socialProviders } from "./options/socialProviders";

export const auth = betterAuth({
  baseURL: authEnv.AUTH_URL,
  basePath: authEnv.AUTH_PATH,
  trustedOrigins: authEnv.AUTH_TRUSTED_ORIGINS,
  database: drizzleAdapter(authDB, {
    provider: "pg",
    schemaName: "auth",
    schema,
  }),
  advanced: {
    database: {
      generateId: "uuid",
    },
  },
  plugins: [
    admin({
      adminRoles: ["admin", "superadmin"],
      roles: {
        ...defaultRoles,
        superadmin: adminAc,
      },
    }),
    jwt(),
    openAPI(),
    ...(process.env.E2E === "true" || process.env.NODE_ENV === "test"
      ? [testUtils()]
      : []),
  ],
  databaseHooks: {
    user: {
      create: {
        after: async (user) => {
          try {
            const uuidRegex =
              /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
            if (uuidRegex.test(user.id)) {
              await authDB.execute(
                sql`INSERT INTO public.users (id, name, email, avatar)
                    VALUES (${user.id}, ${user.name}, ${user.email}, ${user.image ?? null})
                    ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, avatar = EXCLUDED.avatar`,
              );
              if (user.role === "admin" || user.role === "superadmin") {
                await authDB.execute(
                  sql`INSERT INTO public.memberships (user_id, scope_type, scope_id, role, created_at)
                      VALUES (${user.id}, 'system', 0, 'superadmin', NOW())`,
                );
              }
            }
          } catch (e) {
            console.error(
              "[better-auth] Failed to sync user to public tables:",
              e,
            );
          }
        },
      },
      delete: {
        after: async (user) => {
          try {
            await authDB.execute(
              sql`DELETE FROM public.dioceses WHERE created_by = ${user.id} OR updated_by = ${user.id}`,
            );
            await authDB.execute(
              sql`DELETE FROM public.memberships WHERE user_id = ${user.id}`,
            );
            await authDB.execute(
              sql`DELETE FROM public.users WHERE id = ${user.id}`,
            );
          } catch (e) {
            console.error(
              "[better-auth] Failed to clean up user in public tables:",
              e,
            );
          }
        },
      },
    },
  },
  socialProviders,
});

export {
  type SeedSuperadminOptions,
  type SeedSuperadminResult,
  seedSuperadmin,
} from "./seedSuperadmin";
export { authDB, authEnv };
