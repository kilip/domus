import { drizzleAdapter } from "@better-auth/drizzle-adapter/relations-v2";
import { betterAuth } from "better-auth";
import { admin, jwt, openAPI, testUtils } from "better-auth/plugins";
import { adminAc, defaultRoles } from "better-auth/plugins/admin/access";
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
        before: async (user) => {
          // Providers such as GitHub can return an unverified email. Reject, otherwise
          // someone could claim another person's invitation by setting that email.
          if (!user.emailVerified) return false;

          let res: Response;
          try {
            res = await fetch(
              `${authEnv.API_INTERNAL_URL}/internal/users/resolve`,
              {
                method: "POST",
                headers: {
                  "Content-Type": "application/json",
                  "X-Internal-Key": authEnv.API_INTERNAL_KEY,
                },
                body: JSON.stringify({ email: user.email.toLowerCase() }),
                signal: AbortSignal.timeout(5000),
              },
            );
          } catch {
            return false; // API unreachable: fail closed
          }
          if (!res.ok) return false; // not invited

          const { role } = (await res.json()) as { role: string };
          return { data: { ...user, role } };
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
