import type { BetterAuthOptions } from "better-auth";
import { authEnv } from "../authEnv";

type Providers = BetterAuthOptions["socialProviders"];

function withGoogle(): Providers {
  let providers: Providers = {};

  if (authEnv.AUTH_GOOGLE_ID && authEnv.AUTH_GOOGLE_SECRET) {
    providers = {
      google: {
        clientId: authEnv.AUTH_GOOGLE_ID,
        clientSecret: authEnv.AUTH_GOOGLE_SECRET,
      },
    };
  }
  return providers;
}

function withGitHub(): Providers {
  let providers: Providers = {};
  if (authEnv.AUTH_GITHUB_ID && authEnv.AUTH_GITHUB_SECRET) {
    providers = {
      github: {
        clientId: authEnv.AUTH_GITHUB_ID,
        clientSecret: authEnv.AUTH_GITHUB_SECRET,
      },
    };
  }
  return providers;
}

export const socialProviders = {
  ...withGitHub(),
  ...withGoogle(),
} satisfies Providers;
