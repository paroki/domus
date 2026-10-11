import { drizzleAdapter } from "@better-auth/drizzle-adapter/relations-v2";
import { betterAuth } from "better-auth";
import { admin, jwt, openAPI, testUtils } from "better-auth/plugins";
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
  plugins: [
    admin(),
    jwt(),
    openAPI(),
    ...(process.env.E2E === "true" || process.env.NODE_ENV === "test"
      ? [testUtils()]
      : []),
  ],
  socialProviders,
});

export { authEnv };
