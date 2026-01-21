import js from "@eslint/js";
import type { Linter } from "eslint";

export const baseConfigs = [
  { ignores: ["dist", "node_modules", "**/*.d.ts"] } satisfies Pick<Linter.Config, "ignores">,
  { files: ["**/*.js", "**/*.ts", "**/*.tsx"] } satisfies Pick<Linter.Config, "files">,
  js.configs.recommended,
];
