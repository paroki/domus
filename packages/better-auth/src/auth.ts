import { drizzleAdapter } from "@better-auth/drizzle-adapter/relations-v2";
import { betterAuth } from "better-auth";
import { admin, jwt, openAPI, organization } from "better-auth/plugins";
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
    organization({
      teams: {
        enabled: true,
        defaultTeam: {
          enabled: false,
        },
      },
    }),
    jwt(),
    openAPI(),
  ],
  socialProviders,
});

export { authEnv };
