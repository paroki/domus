export const env = {
  baseUrl: process.env.E2E_BASE_URL ?? "http://localhost:3001",
  headed: process.env.HEADED === "true",
};
