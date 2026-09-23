import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  envPrefix: "PUBLIC_",
  envDir: "..",
  server: {
    port: 5181,
  },
  optimizeDeps: {
    exclude: ["@probo/example-cookie-banner-shared"],
  },
});
