export const env = {
  baseUrl: process.env.E2E_BASE_URL ?? "http://localhost:3001",
  authUrl: process.env.AUTH_URL ?? "http://localhost:8001",
  apiUrl: process.env.API_URL ?? "http://localhost:8002",
  headed: process.env.HEADED === "true",
  autoStart: process.env.E2E_AUTO_START !== "false",
  startApi: process.env.E2E_START_API !== "false",
};
