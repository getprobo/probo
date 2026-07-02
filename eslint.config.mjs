import { configs } from "@probo/eslint-config";
import { defineConfig, globalIgnores } from "eslint/config";

// Workspaces that are linted by this root config. Each gets the shared rule
// sets below; everything else is ignored so a bare `eslint .` keeps the same
// scope as the previous per-workspace configs.
const appDirs = ["apps/console/**", "apps/trust/**", "apps/compliance-portal/**"];
const reactDirs = [...appDirs, "packages/ui/**", "packages/relay/**", "packages/routes/**"];
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
    // compliance-portal mutates through the awaitable useMutation bound in
    // #/lib/relay/useMutation (over @probo/relay's createUseMutation), never
    // react-relay's useMutation directly. Scoped to this app only: console and
    // trust still use react-relay's useMutation.
    files: ["apps/compliance-portal/**"],
    rules: {
      "no-restricted-imports": [
        "error",
        {
          paths: [
            {
              name: "react-relay",
              importNames: ["useMutation"],
              message:
                "Use useMutation from #/lib/relay/useMutation, not react-relay.",
            },
            {
              name: "i18next",
              importNames: ["default"],
              message:
                "Don't import i18next's default (global singleton). Build a dedicated instance via `import { createInstance } from \"i18next\"`.",
            },
          ],
        },
      ],
    },
  },
  {
    // The v2 kit styles with tailwind-variants/lite (no tailwind-merge): the
    // numbered scales (text-1…9, rounded-1…6, shadow-1…6) collide with the
    // color/utility namespaces and tailwind-merge would drop the scale class.
    files: ["packages/ui/src/v2/**"],
    rules: {
      "no-restricted-imports": [
        "error",
        {
          paths: [
            {
              name: "tailwind-variants",
              message:
                "Import from tailwind-variants/lite (no tailwind-merge) in v2.",
            },
            {
              name: "tailwind-merge",
              message: "The v2 kit does not use tailwind-merge.",
            },
            {
              name: "clsx",
              message: "The v2 kit does not use clsx; style via tailwind-variants/lite.",
            },
          ],
        },
      ],
    },
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
