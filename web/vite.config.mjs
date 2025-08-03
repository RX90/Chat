import { defineConfig } from "vite";

export default defineConfig({
  root: "templates",
  publicDir: "../static",
  server: {
    port: 3000,
    proxy: {
      "/api": {
        target: "http://localhost:8080",
        changeOrigin: true,
        secure: false,
      },
    },
  },
});