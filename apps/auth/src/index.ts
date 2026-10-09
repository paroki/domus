import { auth, authEnv } from "@domus/better-auth/auth";
import { cors } from "@elysia/cors";
import { openapi } from "@elysia/openapi";
import { Elysia } from "elysia";
import { OpenAPI } from "./openapi";

const app = new Elysia()
  .onRequest(({ request }) => {
    console.log(`[REQ] ${request.method} ${request.url}`);
  })
  .use(
    cors({
      origin: [authEnv.AUTH_URL, ...authEnv.AUTH_TRUSTED_ORIGINS],
      methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"],
      credentials: true,
      allowedHeaders: ["Content-Type", "Authorization"],
    }),
  )
  .use(
    openapi({
      documentation: {
        components: await OpenAPI.components,
        paths: await OpenAPI.getPaths(),
      },
    }),
  )
  .mount(auth.handler)
  .get("/", ({ redirect }) => redirect("/openapi"))
  .listen(authEnv.AUTH_PORT);

console.log(
  `🦊 Elysia is running at ${app.server?.hostname}:${app.server?.port}`,
);
