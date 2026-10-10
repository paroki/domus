import {
  adminClient,
  inferAdditionalFields,
  jwtClient,
  organizationClient,
} from "better-auth/client/plugins";
import { createAuthClient } from "better-auth/react";
import type { auth } from "./auth";

export const { signIn, signOut, getSession, token } = createAuthClient({
  baseURL: "http://localhost:8001",
  basePath: "",
  plugins: [
    adminClient(),
    organizationClient(),
    jwtClient(),
    inferAdditionalFields<typeof auth>(),
  ],
});
