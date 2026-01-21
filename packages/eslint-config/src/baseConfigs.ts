import js from "@eslint/js";

export const baseConfigs = [
  { ignores: ["dist", "node_modules", "**/*.d.ts"] },
  { files: ["**/*.js", "**/*.ts", "**/*.tsx"] },
  js.configs.recommended,
];
