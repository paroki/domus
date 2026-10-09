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
      AUTH_TRUSTED_ORIGINS: z.string().transform((value) => value.split(",")),
    },
    runtimeEnv: process.env,
    isServer: typeof window === "undefined",
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

export const authEnv = authEnvConfig();
