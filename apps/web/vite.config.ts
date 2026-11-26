import { defineConfig } from "vitest/config"
import { fileURLToPath } from "node:url"
import react from "@vitejs/plugin-react"
import tailwindcss from "@tailwindcss/vite"
export default defineConfig(({ command }) => ({
  plugins: [react(), tailwindcss()],
  resolve: { alias: { "@": fileURLToPath(new URL("./src", import.meta.url)) } },
  publicDir: command === "serve" ? "dev" : false,
  server: { proxy: { "/api": "http://localhost:5141" } },
  test: {
    coverage: {
      provider: "v8",
      include: ["src/componentes/**/*.{ts,tsx}"],
      reporter: [
        "text",
        ["lcov", { projectRoot: fileURLToPath(new URL("../../", import.meta.url)) }],
      ],
    },
  },
}))
