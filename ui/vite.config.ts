import { reactRouter } from "@react-router/dev/vite";
import tailwindcss from "@tailwindcss/vite";
import { sentryVitePlugin } from "@sentry/vite-plugin";
import { defineConfig } from "vite";

export default defineConfig({
  // Expose only APP_SENTRY_* (never APP_DB_URL etc.) to the client bundle.
  envPrefix: ["VITE_", "APP_SENTRY_"],
  build: { sourcemap: true },
  plugins: [
    tailwindcss(),
    reactRouter(),
    // Uploads source maps only when APP_SENTRY_AUTH_TOKEN is set (CI/release build).
    sentryVitePlugin({
      org: "doyolabs",
      project: "domus-ui",
      authToken: process.env.APP_SENTRY_AUTH_TOKEN,
      disable: !process.env.APP_SENTRY_AUTH_TOKEN,
      release: { name: process.env.APP_SENTRY_RELEASE },
      sourcemaps: { filesToDeleteAfterUpload: ["build/client/**/*.map"] },
    }),
  ],
  resolve: {
    tsconfigPaths: true,
  },
  // ponytail: targets default backend port 8080; read process.env.VITE_API_TARGET when multi-port dev setups needed.
  server: {
    host: true,
    allowedHosts: ["dev.itstoni.com"],
    proxy: {
      "/api": {
        target: "http://127.0.0.1:8080",
        changeOrigin: true,
      },
      "/mcp": {
        target: "http://127.0.0.1:8080",
        changeOrigin: true,
      },
    },
  },
});
