import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { authEnv, authEnvConfig, resetAuthEnvCache } from "./authEnv";

describe("authEnvConfig", () => {
  const originalEnv = process.env;

  beforeEach(() => {
    process.env = { ...originalEnv };
  });

  afterEach(() => {
    process.env = originalEnv;
    resetAuthEnvCache();
  });

  it("should parse valid environment variables and derive default host and port from URL with port", () => {
    process.env.AUTH_URL = "http://localhost:3000";
    process.env.AUTH_PATH = "/api/auth";
    process.env.AUTH_DB_URL = "postgres://user:pass@localhost:5432/db";
    delete process.env.AUTH_HOST;
    delete process.env.AUTH_PORT;

    const env = authEnvConfig();

    expect(env.AUTH_URL).toBe("http://localhost:3000");
    expect(env.AUTH_PATH).toBe("/api/auth");
    expect(env.AUTH_DB_URL).toBe("postgres://user:pass@localhost:5432/db");
    expect(env.AUTH_HOST).toBe("localhost");
    expect(env.AUTH_PORT).toBe(3000);
  });

  it("should default AUTH_PORT to 443 when protocol is https and no port is specified in URL", () => {
    process.env.AUTH_URL = "https://auth.example.com";
    process.env.AUTH_PATH = "/api/auth";
    process.env.AUTH_DB_URL = "postgres://user:pass@localhost:5432/db";
    delete process.env.AUTH_HOST;
    delete process.env.AUTH_PORT;

    const env = authEnvConfig();

    expect(env.AUTH_HOST).toBe("auth.example.com");
    expect(env.AUTH_PORT).toBe(443);
  });

  it("should default AUTH_PORT to 8001 when protocol is http and no port is specified in URL", () => {
    process.env.AUTH_URL = "http://auth.example.com";
    process.env.AUTH_PATH = "/api/auth";
    process.env.AUTH_DB_URL = "postgres://user:pass@localhost:5432/db";
    delete process.env.AUTH_HOST;
    delete process.env.AUTH_PORT;

    const env = authEnvConfig();

    expect(env.AUTH_HOST).toBe("auth.example.com");
    expect(env.AUTH_PORT).toBe(8001);
  });

  it("should use explicit AUTH_HOST and AUTH_PORT overrides when provided", () => {
    process.env.AUTH_URL = "http://localhost:3000";
    process.env.AUTH_HOST = "custom-host.internal";
    process.env.AUTH_PORT = "9000";
    process.env.AUTH_PATH = "/api/auth";
    process.env.AUTH_DB_URL = "postgres://user:pass@localhost:5432/db";

    const env = authEnvConfig();

    expect(env.AUTH_HOST).toBe("custom-host.internal");
    expect(env.AUTH_PORT).toBe(9000);
  });

  it("should throw validation error when AUTH_URL is missing", () => {
    delete process.env.AUTH_URL;
    process.env.AUTH_PATH = "/api/auth";
    process.env.AUTH_DB_URL = "postgres://user:pass@localhost:5432/db";

    expect(() => authEnvConfig()).toThrow();
  });

  it("should throw validation error when AUTH_URL is invalid", () => {
    process.env.AUTH_URL = "invalid-url";
    process.env.AUTH_PATH = "/api/auth";
    process.env.AUTH_DB_URL = "postgres://user:pass@localhost:5432/db";

    expect(() => authEnvConfig()).toThrow();
  });

  it("should throw validation error when AUTH_PATH is missing", () => {
    process.env.AUTH_URL = "http://localhost:3000";
    delete process.env.AUTH_PATH;
    process.env.AUTH_DB_URL = "postgres://user:pass@localhost:5432/db";

    expect(() => authEnvConfig()).toThrow();
  });

  it("should throw validation error when AUTH_DB_URL is missing", () => {
    process.env.AUTH_URL = "http://localhost:3000";
    process.env.AUTH_PATH = "/api/auth";
    delete process.env.AUTH_DB_URL;

    expect(() => authEnvConfig()).toThrow();
  });

  it("should parse AUTH_TRUSTED_ORIGINS into an array or default to empty array", () => {
    process.env.AUTH_URL = "http://localhost:3000";
    process.env.AUTH_PATH = "/api/auth";
    process.env.AUTH_DB_URL = "postgres://user:pass@localhost:5432/db";
    delete process.env.AUTH_TRUSTED_ORIGINS;

    const envDefault = authEnvConfig();
    expect(envDefault.AUTH_TRUSTED_ORIGINS).toEqual([]);

    process.env.AUTH_TRUSTED_ORIGINS =
      "http://localhost:3001, http://localhost:8002";
    const envWithOrigins = authEnvConfig();
    expect(envWithOrigins.AUTH_TRUSTED_ORIGINS).toEqual([
      "http://localhost:3001",
      "http://localhost:8002",
    ]);
  });

  it("should access authEnv via proxy lazily", () => {
    process.env.AUTH_URL = "http://localhost:3000";
    process.env.AUTH_PATH = "/api/auth";
    process.env.AUTH_DB_URL = "postgres://user:pass@localhost:5432/db";
    delete process.env.AUTH_HOST;
    delete process.env.AUTH_PORT;

    expect(authEnv.AUTH_URL).toBe("http://localhost:3000");
    expect(authEnv.AUTH_PORT).toBe(3000);
    expect(authEnv.AUTH_HOST).toBe("localhost");
  });

  it("should parse SUPERADMIN_EMAIL and default SUPERADMIN_NAME", () => {
    process.env.AUTH_URL = "http://localhost:3000";
    process.env.AUTH_PATH = "/api/auth";
    process.env.AUTH_DB_URL = "postgres://user:pass@localhost:5432/db";
    delete process.env.SUPERADMIN_EMAIL;
    delete process.env.SUPERADMIN_NAME;

    const envDefault = authEnvConfig();
    expect(envDefault.SUPERADMIN_EMAIL).toBeUndefined();
    expect(envDefault.SUPERADMIN_NAME).toBe("Super Admin");

    process.env.SUPERADMIN_EMAIL = "admin@example.com";
    process.env.SUPERADMIN_NAME = "Root Admin";

    const envCustom = authEnvConfig();
    expect(envCustom.SUPERADMIN_EMAIL).toBe("admin@example.com");
    expect(envCustom.SUPERADMIN_NAME).toBe("Root Admin");
  });
});
