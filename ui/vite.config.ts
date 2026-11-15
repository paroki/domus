import { reactRouter } from "@react-router/dev/vite";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "vite";

export default defineConfig({
  plugins: [tailwindcss(), reactRouter()],
  resolve: {
    tsconfigPaths: true,
  },
  // ponytail: targets default backend port 8080; read process.env.API_TARGET when multi-port dev setups needed.
  server: {
    host: true,
    allowedHosts: ["dev.itstoni.com"],
    proxy: {
      "/api": {
        target: "http://127.0.0.1:8080",
        changeOrigin: true
      },
    },
  },
});
