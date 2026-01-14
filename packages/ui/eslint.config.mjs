import { browser } from "@probo/eslint-config";

export default [
    {
        ignores: [
            "dist",
            "eslint.config.mjs",
            "*.config.{js,mjs,ts}",
            ".storybook/**",
            "src/Atoms/Icons/**/*.tsx", // Icon files have parsing issues, likely auto-generated
        ],
    },
    ...browser(["./tsconfig.app.json", "./tsconfig.node.json"], import.meta.dirname),
];
