import { vanillaExtractPlugin } from "@vanilla-extract/vite-plugin";
import react from "@vitejs/plugin-react";
import path from "path";
import { defineConfig } from "vite";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react(), vanillaExtractPlugin()],
  server: {
    proxy: {
      "/api/v1/": {
        target: `http://${process.env.VITE_API_URL}`,
        changeOrigin: true,
      },
      "/api/v1/graphql": {
        target: `ws://${process.env.VITE_API_URL}`,
        changeOrigin: true,
        ws: true,
      },
    },
    port: process.env.PORT ? parseInt(process.env.PORT, 10) : 3000,
  },
  resolve: {
    alias: {
      "@/generated": path.resolve(__dirname, "./generated"),
      "@": path.resolve(__dirname, "./src"),
    },
  },
});
