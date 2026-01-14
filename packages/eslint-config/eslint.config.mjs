import js from "@eslint/js";
import globals from "globals";
import reactHooks from "eslint-plugin-react-hooks";
import tseslint from "typescript-eslint";
import stylistic from "@stylistic/eslint-plugin";
import prettier from "eslint-plugin-prettier";
import prettierConfig from "eslint-config-prettier";

/**
 * Base ESLint configuration with strict and stylistic rules
 */
export function createConfig({
    files = ["**/*.{ts,tsx}"],
    languageOptions = {},
    react = false,
    project = true,
    tsconfigRootDir,
} = {}) {
    const configs = [
        { ignores: ["dist", "node_modules", "**/*.d.ts"] },
        js.configs.recommended,
        ...tseslint.configs.recommendedTypeChecked,
        ...tseslint.configs.stylisticTypeChecked,
        {
            files,
            languageOptions: (() => {
                const mergedGlobals = {
                    ...globals.node,
                    ...globals.es2021,
                    ...(languageOptions.globals || {}),
                };
                const mergedParserOptions = {
                    project,
                    ...(tsconfigRootDir && { tsconfigRootDir }),
                    ...(languageOptions.parserOptions || {}),
                };
                const otherLanguageOptions = Object.fromEntries(
                    Object.entries(languageOptions).filter(([key]) => key !== "globals" && key !== "parserOptions"),
                );

                return {
                    ecmaVersion: 2020,
                    sourceType: "module",
                    globals: mergedGlobals,
                    parserOptions: mergedParserOptions,
                    ...otherLanguageOptions,
                };
            })(),
            plugins: {
                "@stylistic": stylistic,
                prettier,
            },
            rules: {
                // Prettier integration
                "prettier/prettier": "error",

                // TypeScript strict rules
                "@typescript-eslint/no-unused-vars": [
                    "error",
                    {
                        argsIgnorePattern: "^_",
                        varsIgnorePattern: "^_",
                    },
                ],
                "@typescript-eslint/explicit-function-return-type": "off",
                "@typescript-eslint/no-explicit-any": "warn",
                "@typescript-eslint/no-floating-promises": "error",
                "@typescript-eslint/no-misused-promises": "error",
                "@typescript-eslint/await-thenable": "error",
                "@typescript-eslint/no-unnecessary-type-assertion": "error",

                // Stylistic rules
                "@stylistic/quotes": ["error", "double", { avoidEscape: true }],
                "@stylistic/semi": ["error", "always"],
                "@stylistic/comma-dangle": ["error", "always-multiline"],
                "@stylistic/indent": ["error", 4, { SwitchCase: 1 }],
                "@stylistic/object-curly-spacing": ["error", "always"],
                "@stylistic/array-bracket-spacing": ["error", "never"],
                "@stylistic/comma-spacing": ["error", { before: false, after: true }],
                "@stylistic/key-spacing": ["error", { beforeColon: false, afterColon: true }],
                "@stylistic/space-before-blocks": ["error", "always"],
                "@stylistic/space-before-function-paren": [
                    "error",
                    { anonymous: "always", named: "never", asyncArrow: "always" },
                ],
                "@stylistic/space-infix-ops": "error",
                "@stylistic/arrow-spacing": ["error", { before: true, after: true }],
                "@stylistic/no-trailing-spaces": "error",
                "@stylistic/eol-last": ["error", "always"],
                "@stylistic/max-len": ["warn", { code: 120, ignoreUrls: true, ignoreStrings: true }],
            },
        },
        prettierConfig,
    ];

    if (react) {
        configs.push({
            files,
            plugins: {
                "react-hooks": reactHooks,
            },
            rules: {
                ...reactHooks.configs.recommended.rules,
            },
        });
    }

    return configs;
}

/**
 * Browser/React configuration
 */
export function browser(project = true, tsconfigRootDir) {
    return createConfig({
        files: ["**/*.{ts,tsx}"],
        languageOptions: {
            globals: {
                ...globals.browser,
            },
        },
        react: true,
        project,
        tsconfigRootDir,
    });
}

/**
 * Node.js configuration
 */
export function node(project = true, tsconfigRootDir) {
    return createConfig({
        files: ["**/*.{ts,tsx}"],
        languageOptions: {
            globals: {
                ...globals.node,
            },
        },
        react: false,
        project,
        tsconfigRootDir,
    });
}
