import * as js from "@eslint/js";

export const baseConfigs = [
  { ignores: ["dist", "node_modules", "**/*.d.ts"] },
  js.configs.recommended,
];
