import path from "path"
import react from "@vitejs/plugin-react"
import { defineConfig } from "vite"

// In Docker Compose, set VITE_PROXY_API=http://api:8080 so the dev server proxies to the api service.
const apiProxyTarget = process.env.VITE_PROXY_API ?? "http://localhost:8080"

export default defineConfig({
  plugins: [react()],
  server: {
    host: true,
    port: 5173,
    proxy: {
      "/api": {
        target: apiProxyTarget,
        changeOrigin: true,
      },
    },
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
})
