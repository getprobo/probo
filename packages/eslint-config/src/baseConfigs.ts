import js from "@eslint/js";
import { globalIgnores } from "eslint/config";
import type { FlatConfig } from "typescript-eslint";

export const baseConfigs: FlatConfig.ConfigArray = [
  { files: ["**/*.js", "**/*.mjs", "**/*.ts", "**/*.tsx"] },
  globalIgnores([
    ".turbo/",
    "**/dist/",
    "**/node_modules/",
    "**/__generated__/",
    "**/*.d.ts",
  ]),
  js.configs.recommended,
];
