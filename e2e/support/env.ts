export const env = {
  baseUrl: process.env.E2E_BASE_URL ?? "http://localhost:3001",
  authUrl: process.env.AUTH_URL ?? "http://localhost:8001",
  apiUrl: process.env.API_URL ?? "http://localhost:8002",
  apiInternalUrl: process.env.API_INTERNAL_URL ?? "http://localhost:8002",
  apiInternalKey:
    process.env.API_INTERNAL_KEY ?? "dev-internal-key-at-least-32-chars-long",
  authJwksUrl: process.env.AUTH_JWKS_URL ?? "http://localhost:8001/jwks",
  headed: process.env.HEADED === "true",
  autoStart: process.env.E2E_AUTO_START !== "false",
  startApi: process.env.E2E_START_API !== "false",
};
