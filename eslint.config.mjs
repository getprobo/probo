import { configs } from "@probo/eslint-config";
import { defineConfig, globalIgnores } from "eslint/config";

// Workspaces that are linted by this root config. Each gets the shared rule
// sets below; everything else is ignored so a bare `eslint .` keeps the same
// scope as the previous per-workspace configs.
const appDirs = ["apps/console/**", "apps/trust/**", "apps/compliance-portal/**"];
const reactDirs = [...appDirs, "packages/ui/**"];
const lintedDirs = [...reactDirs, "packages/eslint-config/**"];

export default defineConfig([
  // Keep `configs.base` global so its `globalIgnores` (dist, __generated__,
  // *.d.ts, ...) stay global rather than being scoped by a wrapping `files`.
  configs.base,
  globalIgnores([
    "examples/**",
    "pkg/**",
    "packages/coredata/**",
    "packages/cookie-banner/**",
    "packages/emails/**",
    "packages/eslint-relay-plugin-types/**",
    "packages/helpers/**",
    "packages/hooks/**",
    "packages/i18n/**",
    "packages/n8n-node/**",
    "packages/prosemirror/**",
    "packages/react-lazy/**",
    "packages/relay/**",
    "packages/routes/**",
    "packages/tsconfig/**",
  ]),
  {
    files: lintedDirs,
    extends: [configs.ts, configs.imports, configs.stylistic],
    // Linting runs from the repo root, so pin the project service root and let
    // it resolve each file to its nearest package tsconfig.json.
    languageOptions: {
      parserOptions: {
        tsconfigRootDir: import.meta.dirname,
      },
    },
  },
  {
    files: reactDirs,
    extends: [configs.react],
  },
  {
    files: appDirs,
    extends: [configs.relay],
  },
  {
    files: reactDirs,
    ignores: ["packages/ui/tailwind.config.js"],
    extends: [configs.languageOptions.browser],
  },
  {
    files: ["packages/eslint-config/**"],
    extends: [configs.languageOptions.node],
  },
  {
    files: ["packages/ui/tailwind.config.js"],
    extends: [configs.languageOptions.node],
    languageOptions: {
      sourceType: "commonjs",
    },
  },
]);
