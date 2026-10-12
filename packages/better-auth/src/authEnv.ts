import { createEnv } from "@t3-oss/env-core";
import { z } from "zod";

export const authEnvConfig = () => {
  const env = createEnv({
    server: {
      AUTH_URL: z.url(),
      AUTH_HOST: z.string().optional(),
      AUTH_PORT: z.coerce.number().int().optional(),
      AUTH_PATH: z.string(),
      AUTH_DB_URL: z.string(),
      AUTH_TRUSTED_ORIGINS: z
        .string()
        .optional()
        .default("")
        .transform((value) =>
          value
            ? value
                .split(",")
                .map((origin) => origin.trim())
                .filter(Boolean)
            : [],
        ),
      AUTH_GOOGLE_ID: z.string().optional(),
      AUTH_GOOGLE_SECRET: z.string().optional(),
      AUTH_GITHUB_ID: z.string().optional(),
      AUTH_GITHUB_SECRET: z.string().optional(),
      API_INTERNAL_URL: z.url(),
      API_INTERNAL_KEY: z.string().min(32),
    },
    runtimeEnv: process.env,
    isServer: typeof window === "undefined",
    clientPrefix: "VITE_",
    client: {
      VITE_AUTH_URL: z.string().optional(),
      VITE_AUTH_PATH: z.string().optional(),
    },
  });

  const parsedUrl = new URL(env.AUTH_URL);
  const defaultPort = parsedUrl.port
    ? Number.parseInt(parsedUrl.port, 10)
    : parsedUrl.protocol === "https:"
      ? 443
      : 8001;

  return {
    ...env,
    AUTH_HOST: env.AUTH_HOST ?? parsedUrl.hostname,
    AUTH_PORT: env.AUTH_PORT ?? defaultPort,
  };
};

let cachedAuthEnv: ReturnType<typeof authEnvConfig> | null = null;

export const resetAuthEnvCache = () => {
  cachedAuthEnv = null;
};

const getAuthEnv = () => {
  if (!cachedAuthEnv) {
    cachedAuthEnv = authEnvConfig();
  }
  return cachedAuthEnv;
};

export const authEnv = new Proxy({} as ReturnType<typeof authEnvConfig>, {
  get(_target, prop, receiver) {
    return Reflect.get(getAuthEnv(), prop, receiver);
  },
  ownKeys(_target) {
    return Reflect.ownKeys(getAuthEnv());
  },
  getOwnPropertyDescriptor(_target, prop) {
    return {
      ...Reflect.getOwnPropertyDescriptor(getAuthEnv(), prop),
      configurable: true,
    };
  },
  has(_target, prop) {
    return Reflect.has(getAuthEnv(), prop);
  },
});
