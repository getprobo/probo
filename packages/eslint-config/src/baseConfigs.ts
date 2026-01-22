import js from "@eslint/js";
import type { Linter } from "eslint";

export const baseConfigs = [
  { ignores: ["dist", "node_modules", "**/*.d.ts"] } satisfies Pick<Linter.Config, "ignores">,
  js.configs.recommended,
];
